package AdminRoot

import (
	"Proyecto1/AdminDisco"
	"Proyecto1/EstructuraDisco"
	"Proyecto1/ManejoArchivo"
	"Proyecto1/SistemaDeArchivos"
	"Proyecto1/Usuario"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func Mkusr(user string, pass string, grp string, buffer *bytes.Buffer) {
	fmt.Fprintf(buffer, "MKUSR---------------------------------------------------------------------\n")
	if Usuario.Dato.GetIDUsuario() == "" && Usuario.Dato.GetIDParticion() == "" {
		fmt.Fprint(buffer, "Error MKUSR: No hay un usuario logueado.\n")
		return
	}
	if Usuario.Dato.GetIDUsuario() != "root" {
		fmt.Fprint(buffer, "Error MKUSR: El usuario no tiene permiso de escritura.\n")
		return
	}

	ParticionesMount := AdminDisco.GetMountedPartitions()
	var filepath string
	var id string
	for _, partitions := range ParticionesMount {
		for _, partition := range partitions {
			if partition.LoggedIn {
				filepath = partition.Ruta
				id = partition.ID
				break
			}
		}
		if filepath != "" {
			break
		}
	}

	file, err := ManejoArchivo.AbrirArchivo(filepath, buffer)
	if err != nil {
		return
	}
	defer file.Close()

	var TempMBR EstructuraDisco.MRB
	if err := ManejoArchivo.LeerObjeto(file, &TempMBR, 0, buffer); err != nil {
		return
	}

	var index int = -1
	for i := 0; i < 4; i++ {
		if TempMBR.Partitions[i].PartSize != 0 && strings.Contains(string(TempMBR.Partitions[i].PartId[:]), id) {
			if TempMBR.Partitions[i].PartStatus[0] == '1' {
				index = i
				break
			}
		}
	}
	if index == -1 {
		fmt.Fprint(buffer, "Error MKUSR: No se encontró la partición.\n")
		return
	}

	var tempSuperblock EstructuraDisco.Superblock
	if err := ManejoArchivo.LeerObjeto(file, &tempSuperblock, int64(TempMBR.Partitions[index].PartStart), buffer); err != nil {
		return
	}

	indexInode := SistemaDeArchivos.BuscarStart("/users.txt", file, tempSuperblock, buffer)
	if indexInode == -1 {
		fmt.Fprint(buffer, "Error MKUSR: No se encontró el archivo /users.txt\n")
		return
	}

	var crrInode EstructuraDisco.Inode
	if err := ManejoArchivo.LeerObjeto(file, &crrInode, int64(tempSuperblock.SB_Inode_Start+indexInode*int32(binary.Size(EstructuraDisco.Inode{}))), buffer); err != nil {
		return
	}

	data := readAllFileBlocks(&crrInode, file, tempSuperblock, buffer)
	cleanedData := LimpiarNull(data)

	if userExists(cleanedData, user) {
		fmt.Fprintf(buffer, "Error MKUR: El usuario ya existe registrado. %s\n", user)
		return
	}

	if !grupExiste(cleanedData, grp) {
		fmt.Fprintf(buffer, "Error MKUR: El grupo no existe registrado. %s\n", grp)
		return
	}

	lastGroupID := getLastGroupID(cleanedData) + 1
	newUserData := fmt.Sprintf("%d,U,%s,%s,%s\n", lastGroupID, grp, user, pass)

	if err := writeNewUserData(&crrInode, cleanedData, newUserData, file, tempSuperblock, buffer); err != nil {
		fmt.Fprintf(buffer, "Error MKUR: No se pudo escribir el nuevo usuario. %s\n", err)
		return
	}
	fmt.Fprintf(buffer, "Usuario creado con éxito: %s.\n", user)
}

func userExists(data string, user string) bool {
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		fields := strings.Split(line, ",")
		if len(fields) >= 4 && fields[1] == "U" && fields[3] == user {
			return true
		}
	}
	return false
}

func grupExiste(data string, grupo string) bool {
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		fields := strings.Split(line, ",")
		if len(fields) >= 2 && fields[1] == "G" && fields[2] == grupo {
			return true
		}
	}
	return false
}

func LimpiarNull(data string) string {
	return strings.TrimRight(data, "\x00")
}

func readAllFileBlocks(inode *EstructuraDisco.Inode, file *os.File, superblock EstructuraDisco.Superblock, buffer *bytes.Buffer) string {
	var data string
	for _, block := range inode.IN_Block {
		if block != -1 {
			var fileBlock EstructuraDisco.FileBlock
			if err := ManejoArchivo.LeerObjeto(file, &fileBlock, int64(superblock.SB_Block_Start+block*int32(binary.Size(EstructuraDisco.FileBlock{}))), buffer); err != nil {
				continue
			}
			data += string(fileBlock.B_Content[:])
		}
	}
	return data
}

func getLastGroupID(data string) int {
	lines := strings.Split(data, "\n")
	valor := 0
	for _, line := range lines {
		fields := strings.Split(line, ",")
		if len(fields) >= 4 && fields[1] == "U" {
			v, err := strconv.Atoi(fields[0])
			if err == nil {
				valor = v
			}
		}
	}
	return valor
}

func writeNewUserData(inode *EstructuraDisco.Inode, existingData, newUserData string, file *os.File, superblock EstructuraDisco.Superblock, buffer *bytes.Buffer) error {
	fullData := existingData + newUserData
	var currentBlock int32 = 0
	var currentOffset int = 0

	for currentOffset < len(fullData) {
		if currentBlock >= int32(len(inode.IN_Block)) {
			fmt.Fprintf(buffer, "Error Grupo Y Usuario: No hay suficientes bloques disponibles.\n")
			return fmt.Errorf("no hay suficientes bloques disponibles")
		}

		if inode.IN_Block[currentBlock] == -1 {
			newBlockIndex, err := createNewFileBlock(inode, &superblock, file, buffer)
			if err != nil {
				return err
			}
			inode.IN_Block[currentBlock] = newBlockIndex
		}

		var fileBlock EstructuraDisco.FileBlock
		start := currentOffset
		end := currentOffset + 64
		if end > len(fullData) {
			end = len(fullData)
		}
		copy(fileBlock.B_Content[:], fullData[start:end])

		if err := ManejoArchivo.EscribirObjeto(file, fileBlock, int64(superblock.SB_Block_Start+inode.IN_Block[currentBlock]*int32(binary.Size(EstructuraDisco.FileBlock{}))), buffer); err != nil {
			fmt.Fprintf(buffer, "Error Grupo Y Usuario: No se pudo escribir el bloque actualizado. %s\n", err)
			return fmt.Errorf("error al escribir el bloque actualizado: %v", err)
		}

		currentOffset = end
		currentBlock++
	}

	inode.IN_Size = int32(len(fullData))
	if err := ManejoArchivo.EscribirObjeto(file, *inode, int64(superblock.SB_Inode_Start+inode.IN_Block[0]*int32(binary.Size(EstructuraDisco.Inode{}))), buffer); err != nil {
		fmt.Fprintf(buffer, "Error Grupo Y Usuario: No se pudo actualizar el inodo. %s\n", err)
		return fmt.Errorf("error al actualizar el inodo: %v", err)
	}

	return nil
}

func Mkgrp(grupos string, buffer *bytes.Buffer) {
	fmt.Fprintf(buffer, "MKGRP---------------------------------------------------------------------\n")
	if Usuario.Dato.GetIDUsuario() == "" && Usuario.Dato.GetIDParticion() == "" {
		fmt.Fprint(buffer, "Error MKGRP: No hay un usuario logueado.\n")
		return
	}
	if Usuario.Dato.GetIDUsuario() != "root" {
		fmt.Fprint(buffer, "Error MKGRP: El usuario no tiene permiso de escritura.\n")
		return
	}

	ParticionesMount := AdminDisco.GetMountedPartitions()
	var filepath string
	var id string

	for _, particiones := range ParticionesMount {
		for _, particion := range particiones {
			if particion.LoggedIn {
				filepath = particion.Ruta
				id = particion.ID
				break
			}
		}
		if filepath != "" {
			break
		}
	}

	file, err := ManejoArchivo.AbrirArchivo(filepath, buffer)
	if err != nil {
		return
	}
	defer file.Close()

	var TempMBR EstructuraDisco.MRB
	if err := ManejoArchivo.LeerObjeto(file, &TempMBR, 0, buffer); err != nil {
		return
	}

	var index int = -1
	for i := 0; i < 4; i++ {
		if TempMBR.Partitions[i].PartSize != 0 && strings.Contains(string(TempMBR.Partitions[i].PartId[:]), id) {
			if TempMBR.Partitions[i].PartStatus[0] == '1' {
				index = i
				break
			}
		}
	}

	if index == -1 {
		fmt.Fprintf(buffer, "Error MKGRP: No se encontró la partición.\n")
		return
	}

	var tempSuperblock EstructuraDisco.Superblock
	if err := ManejoArchivo.LeerObjeto(file, &tempSuperblock, int64(TempMBR.Partitions[index].PartStart), buffer); err != nil {
		return
	}

	indexInode := SistemaDeArchivos.BuscarStart("/users.txt", file, tempSuperblock, buffer)
	if indexInode == -1 {
		fmt.Fprintf(buffer, "Error MKGRP: No se encontró el archivo usuarios.txt\n")
		return
	}

	var crrInode EstructuraDisco.Inode
	if err := ManejoArchivo.LeerObjeto(file, &crrInode, int64(tempSuperblock.SB_Inode_Start+indexInode*int32(binary.Size(EstructuraDisco.Inode{}))), buffer); err != nil {
		return
	}

	newGroupID, err := getNewGroupID(file, &tempSuperblock, &crrInode, grupos, buffer)
	if err != nil {
		fmt.Fprintf(buffer, "Error MKGRP: %s\n", err.Error())
		return
	}

	newGroupEntry := fmt.Sprintf("%d,G,%s\n", newGroupID, grupos)

	if err := writeNewGroupEntry(file, &tempSuperblock, &crrInode, newGroupEntry, buffer); err != nil {
		fmt.Fprintf(buffer, "Error MKGRP: %s\n", err.Error())
		return
	}

	if err := ManejoArchivo.EscribirObjeto(file, crrInode, int64(tempSuperblock.SB_Inode_Start+indexInode*int32(binary.Size(EstructuraDisco.Inode{}))), buffer); err != nil {
		return
	}

	fmt.Fprintf(buffer, "Grupo creado exitosamente: %s.\n", grupos)
}

func getNewGroupID(file *os.File, superblock *EstructuraDisco.Superblock, inode *EstructuraDisco.Inode, grupo string, buffer *bytes.Buffer) (int, error) {
	lastID := 0
	for _, block := range inode.IN_Block {
		if block != -1 {
			var fileBlock EstructuraDisco.FileBlock
			if err := ManejoArchivo.LeerObjeto(file, &fileBlock, int64(superblock.SB_Block_Start+block*int32(binary.Size(EstructuraDisco.FileBlock{}))), buffer); err != nil {
				continue
			}
			content := strings.TrimRight(string(fileBlock.B_Content[:]), "\x00")
			lines := strings.Split(content, "\n")
			for _, line := range lines {
				if strings.Contains(line, "G") {
					parts := strings.Split(line, ",")
					if len(parts) > 1 {
						existingGroup := strings.TrimSpace(parts[2])
						if existingGroup == grupo {
							return 0, fmt.Errorf("el grupo '%s' ya existe", grupo)
						}
					}
					if len(parts) > 0 {
						id, err := strconv.Atoi(strings.TrimSpace(parts[0]))
						if err == nil && id > lastID {
							lastID = id
						}
					}
				}
			}
		}
	}
	return lastID + 1, nil
}

func writeNewGroupEntry(file *os.File, superblock *EstructuraDisco.Superblock, inode *EstructuraDisco.Inode, newEntry string, buffer *bytes.Buffer) error {
	for i, block := range inode.IN_Block {
		if block != -1 {
			var fileBlock EstructuraDisco.FileBlock
			if err := ManejoArchivo.LeerObjeto(file, &fileBlock, int64(superblock.SB_Block_Start+block*int32(binary.Size(EstructuraDisco.FileBlock{}))), buffer); err != nil {
				return fmt.Errorf("no se pudo leer el FileBlock")
			}

			content := strings.TrimRight(string(fileBlock.B_Content[:]), "\x00")
			remainingSpace := 64 - len(content)

			if len(newEntry) <= remainingSpace {
				// El nuevo grupo cabe en este bloque
				copy(fileBlock.B_Content[len(content):], []byte(newEntry))
				return ManejoArchivo.EscribirObjeto(file, fileBlock, int64(superblock.SB_Block_Start+block*int32(binary.Size(EstructuraDisco.FileBlock{}))), buffer)
			}
		} else {
			// Encontramos un bloque vacío, vamos a crear uno nuevo
			newBlockIndex, err := createNewFileBlock(inode, superblock, file, buffer)
			if err != nil {
				return err
			}
			inode.IN_Block[i] = newBlockIndex

			var newFileBlock EstructuraDisco.FileBlock
			copy(newFileBlock.B_Content[:], []byte(newEntry))
			return ManejoArchivo.EscribirObjeto(file, newFileBlock, int64(superblock.SB_Block_Start+newBlockIndex*int32(binary.Size(EstructuraDisco.FileBlock{}))), buffer)
		}
	}

	return fmt.Errorf("no hay espacio disponible para crear el nuevo grupo")
}

func createNewFileBlock(inode *EstructuraDisco.Inode, superblock *EstructuraDisco.Superblock, file *os.File, buffer *bytes.Buffer) (int32, error) {
	fmt.Println(inode.IN_Block)
	var newBlockIndex int32 = -1
	for i := 0; i < int(superblock.SB_Blocks_Count); i++ {
		var blockStatus byte
		if err := ManejoArchivo.LeerObjeto(file, &blockStatus, int64(superblock.SB_Bm_Block_Start+int32(i)), buffer); err != nil {
			return -1, err
		}
		if blockStatus == 0 {
			newBlockIndex = int32(i)
			break
		}
	}

	if newBlockIndex == -1 {
		return -1, fmt.Errorf("no hay bloques disponibles")
	}

	if err := ManejoArchivo.EscribirObjeto(file, byte(1), int64(superblock.SB_Bm_Block_Start+newBlockIndex), buffer); err != nil {
		return -1, err
	}

	return newBlockIndex, nil
}

/*
func Rmgrp(grupo string, buffer *bytes.Buffer) {
	fmt.Fprintf(buffer, "RMGRP---------------------------------------------------------------------\n")
	if Usuario.Dato.GetIDUsuario() == "" && Usuario.Dato.GetIDParticion() == "" {
		fmt.Fprint(buffer, "Error RMGRP: No hay un usuario logueado.\n")
		return
	}
	if Usuario.Dato.GetIDUsuario() != "root" {
		fmt.Fprint(buffer, "Error RMGRP: El usuario no tiene permiso de escritura.\n")
		return
	}

	ParticionesMount := AdminDisco.GetMountedPartitions()
	var filepath string
	var id string

	for _, particiones := range ParticionesMount {
		for _, particion := range particiones {
			if particion.LoggedIn {
				filepath = particion.Ruta
				id = particion.ID
				break
			}
		}
		if filepath != "" {
			break
		}
	}

	file, err := ManejoArchivo.AbrirArchivo(filepath, buffer)
	if err != nil {
		return
	}
	defer file.Close()

	var TempMBR EstructuraDisco.MRB
	if err := ManejoArchivo.LeerObjeto(file, &TempMBR, 0, buffer); err != nil {
		return
	}

	var index int = -1
	for i := 0; i < 4; i++ {
		if TempMBR.Partitions[i].PartSize != 0 && strings.Contains(string(TempMBR.Partitions[i].PartId[:]), id) {
			if TempMBR.Partitions[i].PartStatus[0] == '1' {
				index = i
				break
			}
		}
	}

	if index == -1 {
		fmt.Fprintf(buffer, "Error RMGRP: No se encontró la partición.\n")
		return
	}

	var tempSuperblock EstructuraDisco.Superblock
	if err := ManejoArchivo.LeerObjeto(file, &tempSuperblock, int64(TempMBR.Partitions[index].PartStart), buffer); err != nil {
		return
	}

	indexInode := SistemaDeArchivos.BuscarStart("/users.txt", file, tempSuperblock, buffer)
	if indexInode == -1 {
		fmt.Fprintf(buffer, "Error RMGRP: No se encontró el archivo usuarios.txt\n")
		return
	}

	var crrInode EstructuraDisco.Inode
	if err := ManejoArchivo.LeerObjeto(file, &crrInode, int64(tempSuperblock.SB_Inode_Start+indexInode*int32(binary.Size(EstructuraDisco.Inode{}))), buffer); err != nil {
		return
	}

	if err := removeGroup(file, &tempSuperblock, &crrInode, grupo, buffer); err != nil {
		fmt.Fprintf(buffer, "Error RMGRP: %s\n", err.Error())
		return
	}

	if err := ManejoArchivo.EscribirObjeto(file, crrInode, int64(tempSuperblock.SB_Inode_Start+indexInode*int32(binary.Size(EstructuraDisco.Inode{}))), buffer); err != nil {
		return
	}

	fmt.Fprintf(buffer, "Grupo eliminado con éxito: %s.\n", grupo)

}

func removeGroup(file *os.File, superblock *EstructuraDisco.Superblock, inode *EstructuraDisco.Inode, grupo string, buffer *bytes.Buffer) error {
	var newContent strings.Builder
	groupFound := false

	for _, block := range inode.IN_Block {
		if block != -1 {
			var fileBlock EstructuraDisco.FileBlock
			if err := ManejoArchivo.LeerObjeto(file, &fileBlock, int64(superblock.SB_Block_Start+block*int32(binary.Size(EstructuraDisco.FileBlock{}))), buffer); err != nil {
				return fmt.Errorf("no se pudo leer el FileBlock")
			}
			content := strings.TrimRight(string(fileBlock.B_Content[:]), "\x00")
			lines := strings.Split(content, "\n")
			for _, line := range lines {
				if strings.Contains(line, "G") {
					parts := strings.Split(line, ",")
					if len(parts) > 2 && strings.TrimSpace(parts[2]) == grupo {
						groupFound = true
						continue
					}
				}
				newContent.WriteString(line + "\n")
			}
		}
	}

	if !groupFound {
		return fmt.Errorf("el grupo '%s' no existe", grupo)
	}

	return writeUpdatedContent(file, superblock, inode, newContent.String(), buffer)
}

func Rmusr(usuario string, buffer *bytes.Buffer) {
	fmt.Fprintf(buffer, "RMUSR---------------------------------------------------------------------\n")

	if Usuario.Dato.GetIDUsuario() == "" && Usuario.Dato.GetIDParticion() == "" {
		fmt.Fprint(buffer, "Error RMUSR: No hay un usuario logueado.\n")
		return
	}
	if Usuario.Dato.GetIDUsuario() != "root" {
		fmt.Fprint(buffer, "Error RMUSR: El usuario no tiene permiso de escritura.\n")
		return
	}

	ParticionesMount := AdminDisco.GetMountedPartitions()
	var filepath string
	var id string

	for _, particiones := range ParticionesMount {
		for _, particion := range particiones {
			if particion.LoggedIn {
				filepath = particion.Ruta
				id = particion.ID
				break
			}
		}
		if filepath != "" {
			break
		}
	}

	file, err := ManejoArchivo.AbrirArchivo(filepath, buffer)
	if err != nil {
		return
	}
	defer file.Close()

	var TempMBR EstructuraDisco.MRB
	if err := ManejoArchivo.LeerObjeto(file, &TempMBR, 0, buffer); err != nil {
		return
	}

	var index int = -1
	for i := 0; i < 4; i++ {
		if TempMBR.Partitions[i].PartSize != 0 && strings.Contains(string(TempMBR.Partitions[i].PartId[:]), id) {
			if TempMBR.Partitions[i].PartStatus[0] == '1' {
				index = i
				break
			}
		}
	}

	if index == -1 {
		fmt.Fprintf(buffer, "Error RMUSR: No se encontró la partición.\n")
		return
	}

	var tempSuperblock EstructuraDisco.Superblock
	if err := ManejoArchivo.LeerObjeto(file, &tempSuperblock, int64(TempMBR.Partitions[index].PartStart), buffer); err != nil {
		return
	}

	indexInode := SistemaDeArchivos.BuscarStart("/users.txt", file, tempSuperblock, buffer)
	if indexInode == -1 {
		fmt.Fprintf(buffer, "Error RMUSR: No se encontró el archivo usuarios.txt\n")
		return
	}

	var crrInode EstructuraDisco.Inode
	if err := ManejoArchivo.LeerObjeto(file, &crrInode, int64(tempSuperblock.SB_Inode_Start+indexInode*int32(binary.Size(EstructuraDisco.Inode{}))), buffer); err != nil {
		return
	}

	if err := removeUser(file, &tempSuperblock, &crrInode, usuario, buffer); err != nil {
		fmt.Fprintf(buffer, "Error RMUSR: %s.\n", err.Error())
		return
	}

	if err := ManejoArchivo.EscribirObjeto(file, crrInode, int64(tempSuperblock.SB_Inode_Start+indexInode*int32(binary.Size(EstructuraDisco.Inode{}))), buffer); err != nil {
		return
	}
	fmt.Fprintf(buffer, "Usuario eliminado con éxito: %s.\n", usuario)
}

func removeUser(file *os.File, superblock *EstructuraDisco.Superblock, inode *EstructuraDisco.Inode, usuario string, buffer *bytes.Buffer) error {
	var newContent strings.Builder
	userFound := false

	for _, block := range inode.IN_Block {
		if block != -1 {
			var fileBlock EstructuraDisco.FileBlock
			if err := ManejoArchivo.LeerObjeto(file, &fileBlock, int64(superblock.SB_Block_Start+block*int32(binary.Size(EstructuraDisco.FileBlock{}))), buffer); err != nil {
				return fmt.Errorf("no se pudo leer el FileBlock")
			}
			content := strings.TrimRight(string(fileBlock.B_Content[:]), "\x00")
			lines := strings.Split(content, "\n")
			for _, line := range lines {
				if strings.Contains(line, "U") {
					parts := strings.Split(line, ",")
					if len(parts) >= 4 && strings.TrimSpace(parts[3]) == usuario {
						userFound = true
						continue
					}
				}
				newContent.WriteString(line + "\n")
			}
		}
	}

	if !userFound {
		return fmt.Errorf("el usuario '%s' no existe", usuario)
	}
	return writeUpdatedContent(file, superblock, inode, newContent.String(), buffer)
}

func writeUpdatedContent(file *os.File, superblock *EstructuraDisco.Superblock, inode *EstructuraDisco.Inode, content string, buffer *bytes.Buffer) error {
	contentBytes := []byte(content)
	var currentBlock int32 = 0
	var currentOffset int = 0

	for currentOffset < len(contentBytes) {
		if currentBlock >= int32(len(inode.IN_Block)) {
			return fmt.Errorf("no hay suficientes bloques disponibles")
		}

		if inode.IN_Block[currentBlock] == -1 {
			newBlockIndex, err := createNewFileBlock(inode, superblock, file, buffer)
			if err != nil {
				return err
			}
			inode.IN_Block[currentBlock] = newBlockIndex
		}

		var fileBlock EstructuraDisco.FileBlock
		start := currentOffset
		end := currentOffset + 64
		if end > len(contentBytes) {
			end = len(contentBytes)
		}
		copy(fileBlock.B_Content[:], contentBytes[start:end])

		if err := ManejoArchivo.EscribirObjeto(file, fileBlock, int64(superblock.SB_Block_Start+inode.IN_Block[currentBlock]*int32(binary.Size(EstructuraDisco.FileBlock{}))), buffer); err != nil {
			return fmt.Errorf("error al escribir el bloque actualizado: %v", err)
		}

		currentOffset = end
		currentBlock++
	}

	inode.IN_Size = int32(len(contentBytes))
	return nil
}

func Chgrp(user string, newGroup string, buffer *bytes.Buffer) {
	fmt.Fprintf(buffer, "CHGRP---------------------------------------------------------------------\n")

	if Usuario.Dato.GetIDUsuario() == "" && Usuario.Dato.GetIDParticion() == "" {
		fmt.Fprint(buffer, "Error CHGRP: No hay un usuario logueado.\n")
		return
	}
	if Usuario.Dato.GetIDUsuario() != "root" {
		fmt.Fprint(buffer, "Error CHGRP: El usuario no tiene permiso de escritura.\n")
		return
	}

	ParticionesMount := AdminDisco.GetMountedPartitions()
	var filepath string
	var id string
	for _, partitions := range ParticionesMount {
		for _, partition := range partitions {
			if partition.LoggedIn {
				filepath = partition.Ruta
				id = partition.ID
				break
			}
		}
		if filepath != "" {
			break
		}
	}

	file, err := ManejoArchivo.AbrirArchivo(filepath, buffer)
	if err != nil {
		return
	}
	defer file.Close()

	var TempMBR EstructuraDisco.MRB
	if err := ManejoArchivo.LeerObjeto(file, &TempMBR, 0, buffer); err != nil {
		return
	}

	var index int = -1
	for i := 0; i < 4; i++ {
		if TempMBR.Partitions[i].PartSize != 0 && strings.Contains(string(TempMBR.Partitions[i].PartId[:]), id) {
			if TempMBR.Partitions[i].PartStatus[0] == '1' {
				index = i
				break
			}
		}
	}
	if index == -1 {
		fmt.Fprintf(buffer, "Error CHGRP: No se encontró la partición.\n")
		return
	}

	var tempSuperblock EstructuraDisco.Superblock
	if err := ManejoArchivo.LeerObjeto(file, &tempSuperblock, int64(TempMBR.Partitions[index].PartStart), buffer); err != nil {
		return
	}

	indexInode := SistemaDeArchivos.BuscarStart("/users.txt", file, tempSuperblock, buffer)
	if indexInode == -1 {
		return
	}

	var crrInode EstructuraDisco.Inode
	if err := ManejoArchivo.LeerObjeto(file, &crrInode, int64(tempSuperblock.SB_Inode_Start+indexInode*int32(binary.Size(EstructuraDisco.Inode{}))), buffer); err != nil {
		return
	}

	data := readAllFileBlocks(&crrInode, file, tempSuperblock, buffer)
	cleanedData := LimpiarNull(data)

	if !userExists(cleanedData, user) {
		fmt.Fprintf(buffer, "Error CHGRP: El usuario no existe. %s\n", user)
		return
	}

	if !grupExiste(cleanedData, newGroup) {
		fmt.Fprintf(buffer, "Error CHGRP: El grupo no existe. %s\n", newGroup)
		return
	}

	updatedData, changed := updateUserGroup(cleanedData, user, newGroup)
	if !changed {
		fmt.Fprintf(buffer, "Error CHGRP: No se pudo cambiar el grupo del usuario. %s\n", user)
		return
	}

	if err := writeUpdatedUserData(&crrInode, updatedData, file, tempSuperblock, buffer); err != nil {
		fmt.Fprintf(buffer, "Error CHGRP: No se pudo actualizar el archivo /users.txt: %s\n", err)
		return
	}
	fmt.Fprintf(buffer, "Grupo del usuario cambiado con éxito: %s al grupo %s\n", user, newGroup)
	//fmt.Println("Grupo del usuario cambiado con éxito:", user, "al grupo", newGroup)
}

func updateUserGroup(data string, user string, newGroup string) (string, bool) {
	lines := strings.Split(data, "\n")
	changed := false
	for i, line := range lines {
		fields := strings.Split(line, ",")
		if len(fields) >= 4 && fields[1] == "U" && fields[3] == user {
			fields[2] = newGroup
			lines[i] = strings.Join(fields, ",")
			changed = true
			break
		}
	}
	return strings.Join(lines, "\n"), changed
}

func writeUpdatedUserData(inode *EstructuraDisco.Inode, updatedData string, file *os.File, superblock EstructuraDisco.Superblock, buffer *bytes.Buffer) error {
	blockSize := int32(binary.Size(EstructuraDisco.FileBlock{}))
	dataBlocks := []string{}
	for i := 0; i < len(updatedData); i += int(blockSize) {
		end := i + int(blockSize)
		if end > len(updatedData) {
			end = len(updatedData)
		}
		dataBlocks = append(dataBlocks, updatedData[i:end])
	}
	for i, block := range dataBlocks {
		if int32(i) >= int32(len(inode.IN_Block)) {
			return fmt.Errorf("no hay suficientes bloques disponibles en el inodo")
		}
		if inode.IN_Block[i] == -1 {
			newBlockIndex, err := createNewFileBlock(inode, &superblock, file, buffer)
			if err != nil {
				return fmt.Errorf("error al crear un nuevo bloque: %v", err)
			}
			inode.IN_Block[i] = newBlockIndex
		}
		var fileBlock EstructuraDisco.FileBlock
		copy(fileBlock.B_Content[:], block)
		if err := ManejoArchivo.EscribirObjeto(file, fileBlock, int64(superblock.SB_Block_Start+inode.IN_Block[i]*blockSize), buffer); err != nil {
			return fmt.Errorf("error al escribir el bloque %d: %v", i, err)
		}
	}
	inode.IN_Size = int32(len(updatedData))
	if err := ManejoArchivo.EscribirObjeto(file, *inode, int64(superblock.SB_Inode_Start+inode.IN_Block[0]*int32(binary.Size(EstructuraDisco.Inode{}))), buffer); err != nil {
		return fmt.Errorf("error al actualizar el inodo: %v", err)
	}

	for i := len(dataBlocks); i < len(inode.IN_Block); i++ {
		if inode.IN_Block[i] != -1 {
			inode.IN_Block[i] = -1
		}
	}

	return nil
}
*/

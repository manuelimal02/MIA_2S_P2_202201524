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

	ParticionesMontadas := AdminDisco.GetMountedPartitions()
	var RutaArchivo string
	var id string
	for _, Particiones := range ParticionesMontadas {
		for _, Particion := range Particiones {
			if Particion.LoggedIn {
				RutaArchivo = Particion.Ruta
				id = Particion.ID
				break
			}
		}
		if RutaArchivo != "" {
			break
		}
	}

	archivo, err := ManejoArchivo.AbrirArchivo(RutaArchivo, buffer)
	if err != nil {
		return
	}
	defer archivo.Close()

	var MBRTemporal EstructuraDisco.MRB
	if err := ManejoArchivo.LeerObjeto(archivo, &MBRTemporal, 0, buffer); err != nil {
		return
	}

	var index int = -1
	for i := 0; i < 4; i++ {
		if MBRTemporal.Partitions[i].PartSize != 0 && strings.Contains(string(MBRTemporal.Partitions[i].PartId[:]), id) {
			if MBRTemporal.Partitions[i].PartStatus[0] == '1' {
				index = i
				break
			}
		}
	}
	if index == -1 {
		fmt.Fprint(buffer, "Error MKUSR: No se encontró la partición.\n")
		return
	}

	var SuperBloqueTemporal EstructuraDisco.Superblock
	if err := ManejoArchivo.LeerObjeto(archivo, &SuperBloqueTemporal, int64(MBRTemporal.Partitions[index].PartStart), buffer); err != nil {
		return
	}

	indexInode := SistemaDeArchivos.BuscarStart("/users.txt", archivo, SuperBloqueTemporal, buffer)
	if indexInode == -1 {
		fmt.Fprint(buffer, "Error MKUSR: No se encontró el archivo /users.txt\n")
		return
	}

	var crrInode EstructuraDisco.Inode
	if err := ManejoArchivo.LeerObjeto(archivo, &crrInode, int64(SuperBloqueTemporal.SB_Inode_Start+indexInode*int32(binary.Size(EstructuraDisco.Inode{}))), buffer); err != nil {
		return
	}

	cadena := LeerBloquesArchivos(&crrInode, archivo, SuperBloqueTemporal, buffer)
	CadenaProcesada := LimpiarNull(cadena)

	if ExisteUsuario(CadenaProcesada, user) {
		fmt.Fprintf(buffer, "Error MKUR: El usuario ya existe registrado. %s\n", user)
		return
	}

	if !ExisteGrupo(CadenaProcesada, grp) {
		fmt.Fprintf(buffer, "Error MKUR: El grupo no existe registrado. %s\n", grp)
		return
	}

	UltimoIDGrupo := ObtenerUltimoIDGrupo(CadenaProcesada) + 1
	DatosNuevoUsuario := fmt.Sprintf("%d,U,%s,%s,%s\n", UltimoIDGrupo, grp, user, pass)

	if err := EscribirNuevoUsuario(&crrInode, CadenaProcesada, DatosNuevoUsuario, archivo, SuperBloqueTemporal, buffer); err != nil {
		fmt.Fprintf(buffer, "Error MKUR: No se pudo escribir el nuevo usuario. %s\n", err)
		return
	}
	fmt.Fprintf(buffer, "Usuario creado con éxito: %s.\n", user)
}

func ExisteUsuario(cadena string, user string) bool {
	lines := strings.Split(cadena, "\n")
	for _, line := range lines {
		fields := strings.Split(line, ",")
		if len(fields) >= 4 && fields[1] == "U" && fields[3] == user {
			return true
		}
	}
	return false
}

func ExisteGrupo(cadena string, grupo string) bool {
	lines := strings.Split(cadena, "\n")
	for _, line := range lines {
		fields := strings.Split(line, ",")
		if len(fields) >= 2 && fields[1] == "G" && fields[2] == grupo {
			return true
		}
	}
	return false
}

func LimpiarNull(cadena string) string {
	return strings.TrimRight(cadena, "\x00")
}

func LeerBloquesArchivos(inode *EstructuraDisco.Inode, archivo *os.File, superblock EstructuraDisco.Superblock, buffer *bytes.Buffer) string {
	var cadena string
	for _, block := range inode.IN_Block {
		if block != -1 {
			var fileBlock EstructuraDisco.FileBlock
			if err := ManejoArchivo.LeerObjeto(archivo, &fileBlock, int64(superblock.SB_Block_Start+block*int32(binary.Size(EstructuraDisco.FileBlock{}))), buffer); err != nil {
				continue
			}
			cadena += string(fileBlock.B_Content[:])
		}
	}
	return cadena
}

func ObtenerUltimoIDGrupo(cadena string) int {
	lines := strings.Split(cadena, "\n")
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

func EscribirNuevoUsuario(inode *EstructuraDisco.Inode, existingData, DatosNuevoUsuario string, archivo *os.File, superblock EstructuraDisco.Superblock, buffer *bytes.Buffer) error {
	fullData := existingData + DatosNuevoUsuario
	var BloqueActual int32 = 0
	var OffsetActual int = 0

	for OffsetActual < len(fullData) {
		if BloqueActual >= int32(len(inode.IN_Block)) {
			fmt.Fprintf(buffer, "Error Grupo Y Usuario: No hay suficientes bloques disponibles.\n")
			return fmt.Errorf("no hay suficientes bloques disponibles")
		}

		if inode.IN_Block[BloqueActual] == -1 {
			newBlockIndex, err := CrearBloqueDeArchivos(inode, &superblock, archivo, buffer)
			if err != nil {
				return err
			}
			inode.IN_Block[BloqueActual] = newBlockIndex
		}

		var fileBlock EstructuraDisco.FileBlock
		start := OffsetActual
		end := OffsetActual + 64
		if end > len(fullData) {
			end = len(fullData)
		}
		copy(fileBlock.B_Content[:], fullData[start:end])

		if err := ManejoArchivo.EscribirObjeto(archivo, fileBlock, int64(superblock.SB_Block_Start+inode.IN_Block[BloqueActual]*int32(binary.Size(EstructuraDisco.FileBlock{}))), buffer); err != nil {
			fmt.Fprintf(buffer, "Error Grupo Y Usuario: No se pudo escribir el bloque actualizado. %s\n", err)
			return fmt.Errorf("error al escribir el bloque actualizado: %v", err)
		}

		OffsetActual = end
		BloqueActual++
	}

	inode.IN_Size = int32(len(fullData))
	if err := ManejoArchivo.EscribirObjeto(archivo, *inode, int64(superblock.SB_Inode_Start+inode.IN_Block[0]*int32(binary.Size(EstructuraDisco.Inode{}))), buffer); err != nil {
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

	ParticionesMontadas := AdminDisco.GetMountedPartitions()
	var RutaArchivo string
	var id string

	for _, particiones := range ParticionesMontadas {
		for _, particion := range particiones {
			if particion.LoggedIn {
				RutaArchivo = particion.Ruta
				id = particion.ID
				break
			}
		}
		if RutaArchivo != "" {
			break
		}
	}

	archivo, err := ManejoArchivo.AbrirArchivo(RutaArchivo, buffer)
	if err != nil {
		return
	}
	defer archivo.Close()

	var MBRTemporal EstructuraDisco.MRB
	if err := ManejoArchivo.LeerObjeto(archivo, &MBRTemporal, 0, buffer); err != nil {
		return
	}

	var index int = -1
	for i := 0; i < 4; i++ {
		if MBRTemporal.Partitions[i].PartSize != 0 && strings.Contains(string(MBRTemporal.Partitions[i].PartId[:]), id) {
			if MBRTemporal.Partitions[i].PartStatus[0] == '1' {
				index = i
				break
			}
		}
	}

	if index == -1 {
		fmt.Fprintf(buffer, "Error MKGRP: No se encontró la partición.\n")
		return
	}

	var SuperBloqueTemporal EstructuraDisco.Superblock
	if err := ManejoArchivo.LeerObjeto(archivo, &SuperBloqueTemporal, int64(MBRTemporal.Partitions[index].PartStart), buffer); err != nil {
		return
	}

	indexInode := SistemaDeArchivos.BuscarStart("/users.txt", archivo, SuperBloqueTemporal, buffer)
	if indexInode == -1 {
		fmt.Fprintf(buffer, "Error MKGRP: No se encontró el archivo usuarios.txt\n")
		return
	}

	var crrInode EstructuraDisco.Inode
	if err := ManejoArchivo.LeerObjeto(archivo, &crrInode, int64(SuperBloqueTemporal.SB_Inode_Start+indexInode*int32(binary.Size(EstructuraDisco.Inode{}))), buffer); err != nil {
		return
	}

	NuevoGrupoID, err := ObtenerNuevoGrupoID(archivo, &SuperBloqueTemporal, &crrInode, grupos, buffer)
	if err != nil {
		fmt.Fprintf(buffer, "Error MKGRP: %s\n", err.Error())
		return
	}

	NuevoGrupoEntrante := fmt.Sprintf("%d,G,%s\n", NuevoGrupoID, grupos)

	if err := EscribirNuevoGrupoEntrante(archivo, &SuperBloqueTemporal, &crrInode, NuevoGrupoEntrante, buffer); err != nil {
		fmt.Fprintf(buffer, "Error MKGRP: %s\n", err.Error())
		return
	}

	if err := ManejoArchivo.EscribirObjeto(archivo, crrInode, int64(SuperBloqueTemporal.SB_Inode_Start+indexInode*int32(binary.Size(EstructuraDisco.Inode{}))), buffer); err != nil {
		return
	}

	fmt.Fprintf(buffer, "Grupo creado exitosamente: %s.\n", grupos)
}

func ObtenerNuevoGrupoID(archivo *os.File, superblock *EstructuraDisco.Superblock, inode *EstructuraDisco.Inode, grupo string, buffer *bytes.Buffer) (int, error) {
	lastID := 0
	for _, block := range inode.IN_Block {
		if block != -1 {
			var fileBlock EstructuraDisco.FileBlock
			if err := ManejoArchivo.LeerObjeto(archivo, &fileBlock, int64(superblock.SB_Block_Start+block*int32(binary.Size(EstructuraDisco.FileBlock{}))), buffer); err != nil {
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

func EscribirNuevoGrupoEntrante(archivo *os.File, superblock *EstructuraDisco.Superblock, inode *EstructuraDisco.Inode, newEntry string, buffer *bytes.Buffer) error {
	for i, block := range inode.IN_Block {
		if block != -1 {
			var fileBlock EstructuraDisco.FileBlock
			if err := ManejoArchivo.LeerObjeto(archivo, &fileBlock, int64(superblock.SB_Block_Start+block*int32(binary.Size(EstructuraDisco.FileBlock{}))), buffer); err != nil {
				return fmt.Errorf("no se pudo leer el FileBlock")
			}

			content := strings.TrimRight(string(fileBlock.B_Content[:]), "\x00")
			remainingSpace := 64 - len(content)

			if len(newEntry) <= remainingSpace {
				// El nuevo grupo cabe en este bloque
				copy(fileBlock.B_Content[len(content):], []byte(newEntry))
				return ManejoArchivo.EscribirObjeto(archivo, fileBlock, int64(superblock.SB_Block_Start+block*int32(binary.Size(EstructuraDisco.FileBlock{}))), buffer)
			}
		} else {
			// Encontramos un bloque vacío, vamos a crear uno nuevo
			newBlockIndex, err := CrearBloqueDeArchivos(inode, superblock, archivo, buffer)
			if err != nil {
				return err
			}
			inode.IN_Block[i] = newBlockIndex

			var newFileBlock EstructuraDisco.FileBlock
			copy(newFileBlock.B_Content[:], []byte(newEntry))
			return ManejoArchivo.EscribirObjeto(archivo, newFileBlock, int64(superblock.SB_Block_Start+newBlockIndex*int32(binary.Size(EstructuraDisco.FileBlock{}))), buffer)
		}
	}

	return fmt.Errorf("no hay espacio disponible para crear el nuevo grupo")
}

func CrearBloqueDeArchivos(inode *EstructuraDisco.Inode, superblock *EstructuraDisco.Superblock, archivo *os.File, buffer *bytes.Buffer) (int32, error) {
	fmt.Println(inode.IN_Block)
	var newBlockIndex int32 = -1
	for i := 0; i < int(superblock.SB_Blocks_Count); i++ {
		var blockStatus byte
		if err := ManejoArchivo.LeerObjeto(archivo, &blockStatus, int64(superblock.SB_Bm_Block_Start+int32(i)), buffer); err != nil {
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

	if err := ManejoArchivo.EscribirObjeto(archivo, byte(1), int64(superblock.SB_Bm_Block_Start+newBlockIndex), buffer); err != nil {
		return -1, err
	}

	return newBlockIndex, nil
}

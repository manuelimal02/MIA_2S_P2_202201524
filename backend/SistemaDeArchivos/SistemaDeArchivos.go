package SistemaDeArchivos

import (
	"Proyecto1/AdminDisco"
	"Proyecto1/EstructuraDisco"
	"Proyecto1/ManejoArchivo"
	"Proyecto1/Usuario"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"time"
)

func MKFS(id string, type_ string, buffer *bytes.Buffer) {
	fmt.Fprintf(buffer, "MKFS---------------------------------------------------------------------\n")

	var ParticionesMontadas AdminDisco.ParticionMontada
	var ParticionEncontrada bool

	for _, Particiones := range AdminDisco.GetMountedPartitions() {
		for _, Particion := range Particiones {
			if Particion.ID == id {
				ParticionesMontadas = Particion
				ParticionEncontrada = true
				break
			}
		}
		if ParticionEncontrada {
			break
		}
	}

	if !ParticionEncontrada {
		fmt.Fprintf(buffer, "Error MFKS: La partición: %s no existe.\n", id)
		return
	}

	if ParticionesMontadas.Estado != '1' {
		fmt.Fprintf(buffer, "Error MFKS: La partición %s aún no está montada.\n", id)
		return
	}

	archivo, err := ManejoArchivo.AbrirArchivo(ParticionesMontadas.Ruta, buffer)
	if err != nil {
		return
	}

	var MBRTemporal EstructuraDisco.MRB
	if err := ManejoArchivo.LeerObjeto(archivo, &MBRTemporal, 0, buffer); err != nil {
		return
	}

	var IndiceParticion int = -1
	for i := 0; i < 4; i++ {
		if MBRTemporal.Partitions[i].PartSize != 0 {
			if strings.Contains(string(MBRTemporal.Partitions[i].PartId[:]), id) {
				IndiceParticion = i
				break
			}
		}
	}

	if IndiceParticion == -1 {
		fmt.Fprintf(buffer, "Error MFKS: La partición: %s no existe.\n", id)
		return
	}

	numerador := int32(MBRTemporal.Partitions[IndiceParticion].PartSize - int32(binary.Size(EstructuraDisco.Superblock{})))
	denrominador_base := int32(4 + int32(binary.Size(EstructuraDisco.Inode{})) + 3*int32(binary.Size(EstructuraDisco.FileBlock{})))
	denrominador := denrominador_base
	n := int32(numerador / denrominador)

	// Crear el Superbloque
	var NuevoSuperBloque EstructuraDisco.Superblock
	NuevoSuperBloque.SB_FileSystem_Type = 2
	NuevoSuperBloque.SB_Inodes_Count = n
	NuevoSuperBloque.SB_Blocks_Count = 3 * n
	NuevoSuperBloque.SB_Free_Blocks_Count = 3*n - 2
	NuevoSuperBloque.SB_Free_Inodes_Count = n - 2
	FechaActual := time.Now()
	FechaString := FechaActual.Format("2006-01-02 15:04:05")
	FechaBytes := []byte(FechaString)
	copy(NuevoSuperBloque.SB_Mtime[:], FechaBytes)
	copy(NuevoSuperBloque.SB_Umtime[:], FechaBytes)
	NuevoSuperBloque.SB_Mnt_Count = 1
	NuevoSuperBloque.SB_Magic = 0xEF53
	NuevoSuperBloque.SB_Inode_Size = int32(binary.Size(EstructuraDisco.Inode{}))
	NuevoSuperBloque.SB_Block_Size = int32(binary.Size(EstructuraDisco.FileBlock{}))
	// Calcular las posiciones de los bloques
	NuevoSuperBloque.SB_Bm_Inode_Start = MBRTemporal.Partitions[IndiceParticion].PartStart + int32(binary.Size(EstructuraDisco.Superblock{}))
	NuevoSuperBloque.SB_Bm_Block_Start = NuevoSuperBloque.SB_Bm_Inode_Start + n
	NuevoSuperBloque.SB_Inode_Start = NuevoSuperBloque.SB_Bm_Block_Start + 3*n
	NuevoSuperBloque.SB_Block_Start = NuevoSuperBloque.SB_Inode_Start + n*int32(binary.Size(EstructuraDisco.Inode{}))
	// Escribir el superbloque en el archivo
	SistemaEXT2(n, MBRTemporal.Partitions[IndiceParticion], NuevoSuperBloque, FechaString, archivo, buffer)
	defer archivo.Close()
}

func SistemaEXT2(n int32, Particion EstructuraDisco.Partition, NuevoSuperBloque EstructuraDisco.Superblock, Fecha string, archivo *os.File, buffer *bytes.Buffer) {
	for i := int32(0); i < n; i++ {
		err := ManejoArchivo.EscribirObjeto(archivo, byte(0), int64(NuevoSuperBloque.SB_Bm_Inode_Start+i), buffer)
		if err != nil {
			return
		}
	}
	for i := int32(0); i < 3*n; i++ {
		err := ManejoArchivo.EscribirObjeto(archivo, byte(0), int64(NuevoSuperBloque.SB_Bm_Block_Start+i), buffer)
		if err != nil {
			return
		}
	}
	// Inicializa inodos y bloques con valores predeterminados
	if err := initInodesAndBlocks(n, NuevoSuperBloque, archivo, buffer); err != nil {
		fmt.Println("Error: ", err)
		return
	}
	// Crea la carpeta raíz y el archivo users.txt
	if err := createRootAndUsersFile(NuevoSuperBloque, Fecha, archivo, buffer); err != nil {
		fmt.Println("Error: ", err)
		return
	}
	// Escribe el superbloque actualizado al archivo
	if err := ManejoArchivo.EscribirObjeto(archivo, NuevoSuperBloque, int64(Particion.PartStart), buffer); err != nil {
		fmt.Println("Error: ", err)
		return
	}
	// Marca los primeros inodos y bloques como usados
	if err := markUsedInodesAndBlocks(NuevoSuperBloque, archivo, buffer); err != nil {
		fmt.Println("Error: ", err)
		return
	}
	// Imprimir el Superblock final
	EstructuraDisco.PrintSuperblock(NuevoSuperBloque)
	fmt.Fprintf(buffer, "Partición: %s formateada exitosamente.\n", string(Particion.PartName[:]))

}

// Función auxiliar para inicializar inodos y bloques
func initInodesAndBlocks(n int32, newSuperblock EstructuraDisco.Superblock, file *os.File, buffer *bytes.Buffer) error {
	var newInode EstructuraDisco.Inode
	for i := int32(0); i < 15; i++ {
		newInode.IN_Block[i] = -1
	}

	for i := int32(0); i < n; i++ {
		if err := ManejoArchivo.EscribirObjeto(file, newInode, int64(newSuperblock.SB_Inode_Start+i*int32(binary.Size(EstructuraDisco.Inode{}))), buffer); err != nil {
			return err
		}
	}

	var newFileblock EstructuraDisco.FileBlock
	for i := int32(0); i < 3*n; i++ {
		if err := ManejoArchivo.EscribirObjeto(file, newFileblock, int64(newSuperblock.SB_Block_Start+i*int32(binary.Size(EstructuraDisco.FileBlock{}))), buffer); err != nil {
			return err
		}
	}

	return nil
}

// Función auxiliar para crear la carpeta raíz y el archivo users.txt
func createRootAndUsersFile(newSuperblock EstructuraDisco.Superblock, date string, file *os.File, buffer *bytes.Buffer) error {
	var Inode0, Inode1 EstructuraDisco.Inode
	initInode(&Inode0, date)
	initInode(&Inode1, date)

	Inode0.IN_Block[0] = 0
	Inode1.IN_Block[0] = 1

	// Asignar el tamaño real del contenido
	data := "1,G,root\n1,U,root,root,123\n"
	actualSize := int32(len(data))
	Inode1.IN_Size = actualSize // Esto ahora refleja el tamaño real del contenido

	var Fileblock1 EstructuraDisco.FileBlock
	copy(Fileblock1.B_Content[:], data) // Copia segura de datos a FileBlock

	var Folderblock0 EstructuraDisco.FolderBlock
	Folderblock0.B_Content[0].B_Inodo = 0
	copy(Folderblock0.B_Content[0].B_Name[:], ".")
	Folderblock0.B_Content[1].B_Inodo = 0
	copy(Folderblock0.B_Content[1].B_Name[:], "..")
	Folderblock0.B_Content[2].B_Inodo = 1
	copy(Folderblock0.B_Content[2].B_Name[:], "users.txt")

	// Escribir los inodos y bloques en las posiciones correctas
	if err := ManejoArchivo.EscribirObjeto(file, Inode0, int64(newSuperblock.SB_Inode_Start), buffer); err != nil {
		return err
	}
	if err := ManejoArchivo.EscribirObjeto(file, Inode1, int64(newSuperblock.SB_Inode_Start+int32(binary.Size(EstructuraDisco.Inode{}))), buffer); err != nil {
		return err
	}
	if err := ManejoArchivo.EscribirObjeto(file, Folderblock0, int64(newSuperblock.SB_Block_Start), buffer); err != nil {
		return err
	}
	if err := ManejoArchivo.EscribirObjeto(file, Fileblock1, int64(newSuperblock.SB_Block_Start+int32(binary.Size(EstructuraDisco.FolderBlock{}))), buffer); err != nil {
		return err
	}

	return nil
}

// Función auxiliar para inicializar un inodo
func initInode(inode *EstructuraDisco.Inode, date string) {
	inode.IN_Uid = 1
	inode.IN_Gid = 1
	inode.IN_Size = 0
	copy(inode.IN_Atime[:], date)
	copy(inode.IN_Ctime[:], date)
	copy(inode.IN_Mtime[:], date)
	copy(inode.IN_Perm[:], "664")

	for i := int32(0); i < 15; i++ {
		inode.IN_Block[i] = -1
	}
}

// Función auxiliar para marcar los inodos y bloques usados
func markUsedInodesAndBlocks(newSuperblock EstructuraDisco.Superblock, file *os.File, buffer *bytes.Buffer) error {
	if err := ManejoArchivo.EscribirObjeto(file, byte(1), int64(newSuperblock.SB_Bm_Inode_Start), buffer); err != nil {
		return err
	}
	if err := ManejoArchivo.EscribirObjeto(file, byte(1), int64(newSuperblock.SB_Bm_Inode_Start+1), buffer); err != nil {
		return err
	}
	if err := ManejoArchivo.EscribirObjeto(file, byte(1), int64(newSuperblock.SB_Bm_Block_Start), buffer); err != nil {
		return err
	}
	if err := ManejoArchivo.EscribirObjeto(file, byte(1), int64(newSuperblock.SB_Bm_Block_Start+1), buffer); err != nil {
		return err
	}
	return nil
}

func CAT(files []string, buffer *bytes.Buffer) {
	fmt.Fprintf(buffer, "CAT---------------------------------------------------------------------\n")
	if Usuario.Dato.GetIDParticion() == "" && Usuario.Dato.GetIDUsuario() == "" {
		fmt.Fprintf(buffer, "Error CAT: No hay un usuario logueado.\n")
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

	// Read the MBR
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
		fmt.Fprintf(buffer, "Error CAT: No se encontró la partición.\n")
		return
	}

	var tempSuperblock EstructuraDisco.Superblock
	if err := ManejoArchivo.LeerObjeto(file, &tempSuperblock, int64(TempMBR.Partitions[index].PartStart), buffer); err != nil {
		return
	}

	for _, filePath := range files {
		indexInode := BuscarStart(filePath, file, tempSuperblock, buffer)
		if indexInode == -1 {
			fmt.Fprintf(buffer, "Error: No se pudo encontrar el archivo %s\n", filePath)
			continue
		}

		var crrInode EstructuraDisco.Inode
		if err := ManejoArchivo.LeerObjeto(file, &crrInode, int64(tempSuperblock.SB_Inode_Start+indexInode*int32(binary.Size(EstructuraDisco.Inode{}))), buffer); err != nil {
			continue
		}
		for _, block := range crrInode.IN_Block {
			if block != -1 {
				var fileblock EstructuraDisco.FileBlock
				if err := ManejoArchivo.LeerObjeto(file, &fileblock, int64(tempSuperblock.SB_Block_Start+block*int32(binary.Size(EstructuraDisco.FileBlock{}))), buffer); err != nil {
					continue
				}
				EstructuraDisco.PrintFileblock(fileblock, buffer)
			}
		}
		fmt.Fprintf(buffer, "CAT: Archivo %s Impreso Exitosamente.\n", filePath)
	}
}

func BuscarStart(path string, file *os.File, tempSuperblock EstructuraDisco.Superblock, buffer *bytes.Buffer) int32 {
	TempStepsPath := strings.Split(path, "/")
	RutaPasada := TempStepsPath[1:]
	var Inode0 EstructuraDisco.Inode
	if err := ManejoArchivo.LeerObjeto(file, &Inode0, int64(tempSuperblock.SB_Inode_Start), buffer); err != nil {
		return -1
	}
	return BuscarInodoRuta(RutaPasada, Inode0, file, tempSuperblock, buffer)
}

func BuscarInodoRuta(RutaPasada []string, Inode EstructuraDisco.Inode, file *os.File, tempSuperblock EstructuraDisco.Superblock, buffer *bytes.Buffer) int32 {
	SearchedName := strings.Replace(pop(&RutaPasada), " ", "", -1)
	for _, block := range Inode.IN_Block {
		if block != -1 {
			if len(RutaPasada) == 0 {
				var fileblock EstructuraDisco.FileBlock
				if err := ManejoArchivo.LeerObjeto(file, &fileblock, int64(tempSuperblock.SB_Block_Start+block*int32(binary.Size(EstructuraDisco.FileBlock{}))), buffer); err != nil {
					return -1
				}
				return 1
			} else {
				var crrFolderBlock EstructuraDisco.FolderBlock
				if err := ManejoArchivo.LeerObjeto(file, &crrFolderBlock, int64(tempSuperblock.SB_Block_Start+block*int32(binary.Size(EstructuraDisco.FolderBlock{}))), buffer); err != nil {
					return -1
				}
				for _, folder := range crrFolderBlock.B_Content {
					if strings.Contains(string(folder.B_Name[:]), SearchedName) {
						var NextInode EstructuraDisco.Inode
						if err := ManejoArchivo.LeerObjeto(file, &NextInode, int64(tempSuperblock.SB_Inode_Start+folder.B_Inodo*int32(binary.Size(EstructuraDisco.Inode{}))), buffer); err != nil {
							return -1
						}
						return BuscarInodoRuta(RutaPasada, NextInode, file, tempSuperblock, buffer)
					}
				}
			}
		}
	}
	return -1
}

func pop(s *[]string) string {
	lastIndex := len(*s) - 1
	last := (*s)[lastIndex]
	*s = (*s)[:lastIndex]
	return last
}

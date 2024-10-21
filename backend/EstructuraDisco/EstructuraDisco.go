package EstructuraDisco

import (
	"bytes"
	"fmt"
)

// Master Boot Record (MBR)
type MRB struct {
	MbrTamano        int32
	MbrFechaCreacion [10]byte
	MbrDskSignature  int32
	MbrDskFit        [1]byte
	Partitions       [4]Partition
}

func ImprimirMBR(datos MRB) {
	fmt.Printf("Fecha de Creación: %s, Ajuste: %s, Tamaño: %d, Identificador: %d\n",
		string(datos.MbrFechaCreacion[:]), string(datos.MbrDskFit[:]), datos.MbrTamano, datos.MbrDskSignature)
	for i := 0; i < 4; i++ {
		ImprimirParticion(datos.Partitions[i])
	}
}

// Partition
type Partition struct {
	PartStatus      [1]byte
	PartType        [1]byte
	PartFit         [1]byte
	PartStart       int32
	PartSize        int32
	PartName        [16]byte
	PartCorrelative int32
	PartId          [4]byte
}

func ImprimirParticion(datos Partition) {
	fmt.Printf("Nombre: %s, Tipo: %s, Inicio: %d, Tamaño: %d, Estado: %s, ID: %s, Ajuste: %s, Correlativo: %d\n",
		string(datos.PartName[:]), string(datos.PartType[:]), datos.PartStart, datos.PartSize, string(datos.PartStatus[:]),
		string(datos.PartId[:]), string(datos.PartFit[:]), datos.PartCorrelative)
}

// Extended Boot Record (EBR)
type EBR struct {
	PartMount [1]byte
	PartFit   [1]byte
	PartStart int32
	PartSize  int32
	PartNext  int32
	PartName  [16]byte
}

func PrintEBR(data EBR) {
	fmt.Printf("Name: %s, fit: %c, start: %d, size: %d, next: %d, mount: %c\n",
		string(data.PartName[:]),
		data.PartFit,
		data.PartStart,
		data.PartSize,
		data.PartNext,
		data.PartMount)
}

// Superblock
type Superblock struct {
	SB_FileSystem_Type   int32
	SB_Inodes_Count      int32
	SB_Blocks_Count      int32
	SB_Free_Blocks_Count int32
	SB_Free_Inodes_Count int32
	SB_Mtime             [17]byte
	SB_Umtime            [17]byte
	SB_Mnt_Count         int32
	SB_Magic             int32
	SB_Inode_Size        int32
	SB_Block_Size        int32
	SB_Fist_Ino          int32
	SB_First_Blo         int32
	SB_Bm_Inode_Start    int32
	SB_Bm_Block_Start    int32
	SB_Inode_Start       int32
	SB_Block_Start       int32
}

func PrintSuperblock(sb Superblock) {
	fmt.Println("====== Superblock ======")
	fmt.Printf("S_filesystem_type: %d\n", sb.SB_FileSystem_Type)
	fmt.Printf("S_inodes_count: %d\n", sb.SB_Inodes_Count)
	fmt.Printf("S_blocks_count: %d\n", sb.SB_Blocks_Count)
	fmt.Printf("S_free_blocks_count: %d\n", sb.SB_Free_Blocks_Count)
	fmt.Printf("S_free_inodes_count: %d\n", sb.SB_Free_Inodes_Count)
	fmt.Printf("S_mtime: %s\n", string(sb.SB_Mtime[:]))
	fmt.Printf("S_umtime: %s\n", string(sb.SB_Umtime[:]))
	fmt.Printf("S_mnt_count: %d\n", sb.SB_Mnt_Count)
	fmt.Printf("S_magic: 0x%X\n", sb.SB_Magic)
	fmt.Printf("S_inode_size: %d\n", sb.SB_Inode_Size)
	fmt.Printf("S_block_size: %d\n", sb.SB_Block_Size)
	fmt.Printf("S_fist_ino: %d\n", sb.SB_Fist_Ino)
	fmt.Printf("S_first_blo: %d\n", sb.SB_First_Blo)
	fmt.Printf("S_bm_inode_start: %d\n", sb.SB_Bm_Inode_Start)
	fmt.Printf("S_bm_block_start: %d\n", sb.SB_Bm_Block_Start)
	fmt.Printf("S_inode_start: %d\n", sb.SB_Inode_Start)
	fmt.Printf("S_block_start: %d\n", sb.SB_Block_Start)
	fmt.Println("========================")
}

// Inode
type Inode struct {
	IN_Uid   int32
	IN_Gid   int32
	IN_Size  int32
	IN_Atime [17]byte
	IN_Ctime [17]byte
	IN_Mtime [17]byte
	IN_Block [15]int32
	IN_Type  [1]byte
	IN_Perm  [3]byte
}

func PrintInode(inode Inode) {
	fmt.Println("====== Inode ======")
	fmt.Printf("I_uid: %d\n", inode.IN_Uid)
	fmt.Printf("I_gid: %d\n", inode.IN_Gid)
	fmt.Printf("I_size: %d\n", inode.IN_Size)
	fmt.Printf("I_atime: %s\n", string(inode.IN_Atime[:]))
	fmt.Printf("I_ctime: %s\n", string(inode.IN_Ctime[:]))
	fmt.Printf("I_mtime: %s\n", string(inode.IN_Mtime[:]))
	fmt.Printf("I_type: %s\n", string(inode.IN_Type[:]))
	fmt.Printf("I_perm: %s\n", string(inode.IN_Perm[:]))
	fmt.Printf("I_block: %v\n", inode.IN_Block)
	fmt.Println("===================")
}

// Bloque De Carpetas
type FolderBlock struct {
	B_Content [4]Content
}

type Content struct {
	B_Name  [12]byte
	B_Inodo int32
}

func PrintFolderblock(folderblock FolderBlock) {
	fmt.Println("====== Folderblock ======")
	for i, content := range folderblock.B_Content {
		fmt.Printf("Content %d: Name: %s, Inodo: %d\n", i, string(content.B_Name[:]), content.B_Inodo)
	}
	fmt.Println("=========================")
}

// Bloque De Archivos
type FileBlock struct {
	B_Content [64]byte
}

func PrintFileblock(fileblock FileBlock, buffer *bytes.Buffer) {
	fmt.Fprintf(buffer, "%s\n", string(fileblock.B_Content[:]))
}

type PointerBlock struct {
	B_Pointers [16]int32
}

func PrintPointerblock(pointerblock PointerBlock) {
	fmt.Println("====== Pointerblock ======")
	for i, pointer := range pointerblock.B_Pointers {
		fmt.Printf("Pointer %d: %d\n", i, pointer)
	}
	fmt.Println("=========================")
}

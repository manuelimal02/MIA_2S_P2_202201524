package Usuario

import (
	"Proyecto1/AdminDisco"
	"Proyecto1/EstructuraDisco"
	"Proyecto1/ManejoArchivo"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

type ParticionUsuario struct {
	IDParticion string
	IDUsuario   string
}

func (Dato *ParticionUsuario) GetIDParticion() string {
	return Dato.IDParticion
}

func (Dato *ParticionUsuario) GetIDUsuario() string {
	return Dato.IDUsuario
}

func (Dato *ParticionUsuario) SetIDParticion(idParticion string) {
	Dato.IDParticion = idParticion
}

func (Dato *ParticionUsuario) SetIDUsuario(idUsuario string) {
	Dato.IDUsuario = idUsuario
}

var Dato ParticionUsuario

func LOGIN(user string, pass string, id string, buffer *bytes.Buffer) {
	fmt.Fprint(buffer, "LOGIN\n")
	ParticionesMontadas := AdminDisco.GetMountedPartitions()
	var RutaArchivo string
	var ParticionEncontrada bool
	var Login bool = false

	for _, Particiones := range ParticionesMontadas {
		for _, Particion := range Particiones {
			if Particion.ID == id && Particion.LoggedIn {
				fmt.Fprintf(buffer, "Error LOGIN: Ya existe un usuario logueado en la partición:%s\n", id)
				return
			}
			if Particion.ID == id {
				RutaArchivo = Particion.Ruta
				ParticionEncontrada = true
				break
			}
		}
		if ParticionEncontrada {
			break
		}
	}

	if !ParticionEncontrada {
		fmt.Fprintf(buffer, "Error LOGIN: No se encontró ninguna partición montada con el ID: %s\n", id)
		return
	}

	file, err := ManejoArchivo.AbrirArchivo(RutaArchivo, buffer)
	if err != nil {
		return
	}
	defer file.Close()

	var MBRTemporal EstructuraDisco.MRB

	if err := ManejoArchivo.LeerObjeto(file, &MBRTemporal, 0, buffer); err != nil {
		return
	}

	var index int = -1
	for i := 0; i < 4; i++ {
		if MBRTemporal.Partitions[i].PartSize != 0 {
			if strings.Contains(string(MBRTemporal.Partitions[i].PartId[:]), id) {
				if MBRTemporal.Partitions[i].PartStatus[0] == '1' {
					index = i
				} else {
					return
				}
				break
			}
		}
	}

	if index == -1 {
		fmt.Fprintf(buffer, "Error LOGIN: No se encontró ninguna partición con el ID: %s\n", id)
		return
	}

	var SuperBloqueTemporal EstructuraDisco.Superblock
	if err := ManejoArchivo.LeerObjeto(file, &SuperBloqueTemporal, int64(MBRTemporal.Partitions[index].PartStart), buffer); err != nil {
		return
	}

	IndexInode := InitSearch("/users.txt", file, SuperBloqueTemporal, buffer)

	var crrInode EstructuraDisco.Inode
	// Leer el Inodo desde el archivo binario
	if err := ManejoArchivo.LeerObjeto(file, &crrInode, int64(SuperBloqueTemporal.SB_Inode_Start+IndexInode*int32(binary.Size(EstructuraDisco.Inode{}))), buffer); err != nil {
		return
	}
	data := GetInodeFileData(crrInode, file, SuperBloqueTemporal, buffer)

	lines := strings.Split(data, "\n")
	for _, line := range lines {
		words := strings.Split(line, ",")
		if len(words) == 5 {
			if (strings.Contains(words[3], user)) && (strings.Contains(words[4], pass)) {
				Login = true
				break
			}
		}
	}

	if Login {
		fmt.Fprintf(buffer, "Usuario logueado con éxito en la partición:%s\n", id)
		AdminDisco.MarkPartitionAsLoggedIn(id)
	}
	Dato.SetIDParticion(id)
	Dato.SetIDUsuario(user)

}

func LOGOUT(buffer *bytes.Buffer) {
	fmt.Fprint(buffer, "LOGOUT\n")
	ParticionesMontadas := AdminDisco.GetMountedPartitions()
	var SesionActiva bool

	if len(ParticionesMontadas) == 0 {
		fmt.Fprintf(buffer, "Error LOGOUT: No hay ninguna partición montada.\n")
		return
	}

	for _, Particiones := range ParticionesMontadas {
		for _, Particion := range Particiones {
			if Particion.LoggedIn {
				SesionActiva = true
				break
			}
		}
		if SesionActiva {
			break
		}
	}
	if !SesionActiva {
		fmt.Fprintf(buffer, "Error LOGOUT: No hay ninguna sesión activa.\n")
		return
	} else {
		AdminDisco.MarkPartitionAsLoggedOut(Dato.GetIDParticion())
		fmt.Fprintf(buffer, "Sesión cerrada con éxito de la partición:%s\n", Dato.GetIDParticion())
	}
	Dato.SetIDParticion("")
	Dato.SetIDUsuario("")
}

func InitSearch(path string, file *os.File, SuperBloqueTemporal EstructuraDisco.Superblock, buffer *bytes.Buffer) int32 {
	TempStepsPath := strings.Split(path, "/")
	StepsPath := TempStepsPath[1:]
	var Inode0 EstructuraDisco.Inode
	if err := ManejoArchivo.LeerObjeto(file, &Inode0, int64(SuperBloqueTemporal.SB_Inode_Start), buffer); err != nil {
		return -1
	}
	return SarchInodeByPath(StepsPath, Inode0, file, SuperBloqueTemporal, buffer)
}

func SarchInodeByPath(StepsPath []string, Inode EstructuraDisco.Inode, file *os.File, SuperBloqueTemporal EstructuraDisco.Superblock, buffer *bytes.Buffer) int32 {
	index := int32(0)
	SearchedName := strings.Replace(pop(&StepsPath), " ", "", -1)
	for _, block := range Inode.IN_Block {
		if block != -1 {
			if index < 13 {
				var crrFolderBlock EstructuraDisco.FolderBlock
				if err := ManejoArchivo.LeerObjeto(file, &crrFolderBlock, int64(SuperBloqueTemporal.SB_Block_Start+block*int32(binary.Size(EstructuraDisco.FolderBlock{}))), buffer); err != nil {
					return -1
				}
				for _, folder := range crrFolderBlock.B_Content {
					if strings.Contains(string(folder.B_Name[:]), SearchedName) {
						if len(StepsPath) == 0 {
							return folder.B_Inodo
						} else {
							var NextInode EstructuraDisco.Inode
							if err := ManejoArchivo.LeerObjeto(file, &NextInode, int64(SuperBloqueTemporal.SB_Inode_Start+folder.B_Inodo*int32(binary.Size(EstructuraDisco.Inode{}))), buffer); err != nil {
								return -1
							}
							return SarchInodeByPath(StepsPath, NextInode, file, SuperBloqueTemporal, buffer)
						}
					}
				}
			}
		}
		index++
	}
	return 0
}

func GetInodeFileData(Inode EstructuraDisco.Inode, file *os.File, SuperBloqueTemporal EstructuraDisco.Superblock, buffer *bytes.Buffer) string {
	index := int32(0)
	var content string
	for _, block := range Inode.IN_Block {
		if block != -1 {
			if index < 13 {
				var crrFileBlock EstructuraDisco.FileBlock
				if err := ManejoArchivo.LeerObjeto(file, &crrFileBlock, int64(SuperBloqueTemporal.SB_Block_Start+block*int32(binary.Size(EstructuraDisco.FileBlock{}))), buffer); err != nil {
					return ""
				}
				content += string(crrFileBlock.B_Content[:])
			}
		}
		index++
	}
	return content
}

func pop(s *[]string) string {
	lastIndex := len(*s) - 1
	last := (*s)[lastIndex]
	*s = (*s)[:lastIndex]
	return last
}

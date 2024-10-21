package AdminDisco

import (
	"Proyecto1/EstructuraDisco"
	"Proyecto1/ManejoArchivo"
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type ParticionMontada struct {
	Ruta     string
	Nombre   string
	ID       string
	Estado   byte
	LoggedIn bool
}

var ListaParticionesMontadas = make(map[string][]ParticionMontada)

func PrintMountedPartitions(ruta string, buffer *bytes.Buffer) {
	if len(ListaParticionesMontadas) == 0 {
		fmt.Println("No hay particiones montadas.")
		return
	}
	for DiscoID, partitions := range ListaParticionesMontadas {
		if DiscoID == ruta {
			fmt.Println("Disco:", DiscoID)
			fmt.Println("---------------------------")
			for _, Particion := range partitions {
				loginStatus := "No"
				if Particion.LoggedIn {
					loginStatus = "Sí"
				}
				fmt.Printf("Nombre: %v, ID: %v, Ruta: %v, Estado: %c, LoggedIn: %v\n",
					Particion.Nombre, Particion.ID, Particion.Ruta, Particion.Estado, loginStatus)
			}
		}
		fmt.Println("---------------------------")
	}
}

func GetMountedPartitions() map[string][]ParticionMontada {
	return ListaParticionesMontadas
}

func MarkPartitionAsLoggedIn(id string) {
	for DiscoID, partitions := range ListaParticionesMontadas {
		for i, Particion := range partitions {
			if Particion.ID == id {
				ListaParticionesMontadas[DiscoID][i].LoggedIn = true
				return
			}
		}
	}
}

func MarkPartitionAsLoggedOut(id string) {
	for DiscoID, partitions := range ListaParticionesMontadas {
		for i, Particion := range partitions {
			if Particion.ID == id {
				ListaParticionesMontadas[DiscoID][i].LoggedIn = false
				return
			}
		}
	}
}

func getLastDiskID() string {
	var UltimoDiscoID string
	for DiscoID := range ListaParticionesMontadas {
		UltimoDiscoID = DiscoID
	}
	return UltimoDiscoID
}

func EliminarDiscoPorRuta(ruta string, buffer *bytes.Buffer) {
	discoID := GenerarDiscoID(ruta)
	if _, existe := ListaParticionesMontadas[discoID]; existe {
		delete(ListaParticionesMontadas, discoID)
		fmt.Fprintf(buffer, "El disco con ruta '%s' y sus particiones asociadas han sido eliminados.\n", ruta)
	}
}

func GenerarDiscoID(path string) string {
	return strings.ToLower(path)
}

func MKDISK(tamano int, ajuste string, unidad string, ruta string, buffer *bytes.Buffer) {
	fmt.Fprintf(buffer, "MKDISK---------------------------------------------------------------------\n")
	// Validar el tamaño (size)
	if tamano <= 0 {
		fmt.Fprintf(buffer, "Error MKDISK: El tamaño del disco debe ser mayor a 0.\n")
		return
	}
	// Validar el ajuste (fit)
	if ajuste != "bf" && ajuste != "ff" && ajuste != "wf" {
		fmt.Fprintf(buffer, "Error MKDISK: El ajuste del disco debe ser BF, FF, o WF.\n")
		return
	}
	// Validar la unidad (unit)
	if unidad != "k" && unidad != "m" {
		fmt.Fprintf(buffer, "Error MKDISK: La unidad de tamaño debe ser Kilobytes o Megabytes.\n")
		return
	}
	// Validar la ruta (path)
	if ruta == "" {
		fmt.Fprintf(buffer, "Error MKDISK: La ruta del disco es obligatoria.\n")
		return
	}

	// Crear el archivo en la ruta especificada
	err := ManejoArchivo.CrearArchivo(ruta, buffer)
	if err != nil {
		return
	}
	// Convertir el tamaño a bytes
	if unidad == "k" {
		tamano = tamano * 1024
	} else {
		tamano = tamano * 1024 * 1024
	}
	// Abrir el archivo para escritura
	archivo, err := ManejoArchivo.AbrirArchivo(ruta, buffer)
	if err != nil {
		return
	}
	// Inicializar el archivo con ceros
	for i := 0; i < tamano; i++ {
		err := ManejoArchivo.EscribirObjeto(archivo, byte(0), int64(i), buffer)
		if err != nil {
			return
		}
	}
	// Inicializar el MBR
	var nuevo_mbr EstructuraDisco.MRB
	nuevo_mbr.MbrTamano = int32(tamano)
	nuevo_mbr.MbrDskSignature = rand.Int31()
	FechaActual := time.Now()
	FechaCreacion := FechaActual.Format("2006-01-02")
	copy(nuevo_mbr.MbrFechaCreacion[:], FechaCreacion)
	copy(nuevo_mbr.MbrDskFit[:], ajuste)
	// Escribir el MBR en el archivo
	if err := ManejoArchivo.EscribirObjeto(archivo, nuevo_mbr, 0, buffer); err != nil {
		return
	}
	var TempMRB EstructuraDisco.MRB
	if err := ManejoArchivo.LeerObjeto(archivo, &TempMRB, 0, buffer); err != nil {
		return
	}
	fmt.Println("---------------------------------------------")
	EstructuraDisco.ImprimirMBR(TempMRB)
	fmt.Println("---------------------------------------------")
	defer archivo.Close()
	fmt.Fprintf(buffer, "Disco creado con éxito en la ruta: %s.\n", ruta)

}

func RMDISK(ruta string, buffer *bytes.Buffer) {
	fmt.Fprintf(buffer, "RMDISK---------------------------------------------------------------------\n")
	// Validar la ruta (path)
	if ruta == "" {
		fmt.Fprintf(buffer, "Error RMDISK: La ruta del disco es obligatoria.\n")
		return
	}
	err := ManejoArchivo.EliminarArchivo(ruta, buffer)
	if err != nil {
		return
	}
	EliminarDiscoPorRuta(ruta, buffer)
	fmt.Fprintf(buffer, "Disco eliminado con éxito en la ruta: %s.\n", ruta)
}

func FDISK(tamano int, unidad string, ruta string, tipo string, ajuste string, nombre string, buffer *bytes.Buffer) {
	fmt.Fprintf(buffer, "FDISK---------------------------------------------------------------------\n")
	// Validar el tamaño (size)
	if tamano <= 0 {
		fmt.Fprintf(buffer, "Error FDISK: EL tamaño de la partición debe ser mayor que 0.\n")
		return
	}
	// Validar la unidad (unit)
	if unidad != "b" && unidad != "k" && unidad != "m" {
		fmt.Fprintf(buffer, "Error FDISK: La unidad de tamaño debe ser Bytes, Kilobytes, Megabytes.\n")
		return
	}
	// Validar la ruta (path)
	if ruta == "" {
		fmt.Fprintf(buffer, "Error FDISK: La ruta del disco es obligatoria.\n")
		return
	}
	// Validar el tipo (type)
	if tipo != "p" && tipo != "e" && tipo != "l" {
		fmt.Fprintf(buffer, "Error FDISK: El tipo de partición debe ser Primaria, Extendida, Lógica.\n")
		return
	}
	// Validar el ajuste (fit)
	if ajuste != "bf" && ajuste != "ff" && ajuste != "wf" {
		fmt.Fprintf(buffer, "Error FDISK: El ajuste de la partición debe ser BF, WF o FF.\n")
		return
	}
	// Validar el nombre (name)
	if nombre == "" {
		fmt.Fprintf(buffer, "Error FDISK: El nombre de la partición es obligatorio.\n")
		return
	}

	// Convertir el tamaño a bytes
	if unidad == "k" {
		tamano = tamano * 1024
	} else if unidad == "m" {
		tamano = tamano * 1024 * 1024
	}

	// Abrir archivo binario
	archivo, err := ManejoArchivo.AbrirArchivo(ruta, buffer)
	if err != nil {
		return
	}

	var MBRTemporal EstructuraDisco.MRB
	if err := ManejoArchivo.LeerObjeto(archivo, &MBRTemporal, 0, buffer); err != nil {
		return
	}

	for i := 0; i < 4; i++ {
		if strings.Contains(string(MBRTemporal.Partitions[i].PartName[:]), nombre) {
			fmt.Fprintf(buffer, "Error FDISK: El nombre: %s ya está en uso en las particiones.\n", nombre)
			return
		}
	}

	var ContadorPrimaria, ContadorExtendida, TotalParticiones int
	var EspacioUtilizado int32 = 0

	for i := 0; i < 4; i++ {
		if MBRTemporal.Partitions[i].PartSize != 0 {
			TotalParticiones++
			EspacioUtilizado += MBRTemporal.Partitions[i].PartSize

			if MBRTemporal.Partitions[i].PartType[0] == 'p' {
				ContadorPrimaria++
			} else if MBRTemporal.Partitions[i].PartType[0] == 'e' {
				ContadorExtendida++
			}
		}
	}

	if TotalParticiones >= 4 && tipo != "l" {
		fmt.Fprintf(buffer, "Error FDISK: No se pueden crear más de 4 particiones primarias o extendidas en total.\n")
		return
	}
	if tipo == "e" && ContadorExtendida > 0 {
		fmt.Fprintf(buffer, "Error FDISK: Solo se permite una partición extendida por disco.\n")
		return
	}
	if tipo == "l" && ContadorExtendida == 0 {
		fmt.Fprintf(buffer, "Error FDISK: No se puede crear una partición lógica sin una partición extendida.\n")
		return
	}
	if EspacioUtilizado+int32(tamano) > MBRTemporal.MbrTamano {
		fmt.Fprintf(buffer, "Error FDISK: No hay suficiente espacio en el disco para crear esta partición.\n")
		return
	}

	var vacio int32 = int32(binary.Size(MBRTemporal))
	if TotalParticiones > 0 {
		vacio = MBRTemporal.Partitions[TotalParticiones-1].PartStart + MBRTemporal.Partitions[TotalParticiones-1].PartSize
	}

	for i := 0; i < 4; i++ {
		if MBRTemporal.Partitions[i].PartSize == 0 {
			if tipo == "p" || tipo == "e" {
				MBRTemporal.Partitions[i].PartSize = int32(tamano)
				MBRTemporal.Partitions[i].PartStart = vacio
				copy(MBRTemporal.Partitions[i].PartName[:], nombre)
				copy(MBRTemporal.Partitions[i].PartFit[:], ajuste)
				copy(MBRTemporal.Partitions[i].PartStatus[:], "0")
				copy(MBRTemporal.Partitions[i].PartType[:], tipo)
				MBRTemporal.Partitions[i].PartCorrelative = int32(TotalParticiones + 1)
				if tipo == "e" {
					EBRInicio := vacio
					EBRNuevo := EstructuraDisco.EBR{
						PartFit:   [1]byte{ajuste[0]},
						PartStart: EBRInicio,
						PartSize:  0,
						PartNext:  -1,
					}
					copy(EBRNuevo.PartName[:], "")
					if err := ManejoArchivo.EscribirObjeto(archivo, EBRNuevo, int64(EBRInicio), buffer); err != nil {
						return
					}
				}
				fmt.Fprintf(buffer, "Partición creada tipo: %s exitosamente en la ruta: %s con el nombre: %s.\n", tipo, ruta, nombre)
				break
			}
		}
	}

	if tipo == "l" {
		var ParticionExtendida *EstructuraDisco.Partition
		for i := 0; i < 4; i++ {
			if MBRTemporal.Partitions[i].PartType[0] == 'e' {
				ParticionExtendida = &MBRTemporal.Partitions[i]
				break
			}
		}
		if ParticionExtendida == nil {
			fmt.Fprintf(buffer, "Error FDISK: No se encontró una partición extendida para crear la partición lógica.\n")
			return
		}

		EBRPosterior := ParticionExtendida.PartStart
		var EBRUltimo EstructuraDisco.EBR
		for {
			if err := ManejoArchivo.LeerObjeto(archivo, &EBRUltimo, int64(EBRPosterior), buffer); err != nil {
				return
			}
			if strings.Contains(string(EBRUltimo.PartName[:]), nombre) {
				fmt.Fprintf(buffer, "Error FDISK: El nombre: %s ya está en uso en las particiones.\n", nombre)
				return
			}
			if EBRUltimo.PartNext == -1 {
				break
			}
			EBRPosterior = EBRUltimo.PartNext
		}

		var EBRNuevoPosterior int32
		if EBRUltimo.PartSize == 0 {
			EBRNuevoPosterior = EBRPosterior
		} else {
			EBRNuevoPosterior = EBRUltimo.PartStart + EBRUltimo.PartSize
		}

		if EBRNuevoPosterior+int32(tamano)+int32(binary.Size(EstructuraDisco.EBR{})) > ParticionExtendida.PartStart+ParticionExtendida.PartSize {
			fmt.Fprintf(buffer, "Error FDISK: No hay suficiente espacio en la partición extendida para esta partición lógica.\n")
			return
		}

		if EBRUltimo.PartSize != 0 {
			EBRUltimo.PartNext = EBRNuevoPosterior
			if err := ManejoArchivo.EscribirObjeto(archivo, EBRUltimo, int64(EBRPosterior), buffer); err != nil {
				return
			}
		}

		newEBR := EstructuraDisco.EBR{
			PartFit:   [1]byte{ajuste[0]},
			PartStart: EBRNuevoPosterior + int32(binary.Size(EstructuraDisco.EBR{})),
			PartSize:  int32(tamano),
			PartNext:  -1,
		}
		copy(newEBR.PartName[:], nombre)
		if err := ManejoArchivo.EscribirObjeto(archivo, newEBR, int64(EBRNuevoPosterior), buffer); err != nil {
			return
		}
		fmt.Fprintf(buffer, "Partición lógica creada exitosamente en la ruta: %s con el nombre: %s.\n", ruta, nombre)
		fmt.Println("---------------------------------------------")
		EBRActual := ParticionExtendida.PartStart
		for {
			var EBRTemp EstructuraDisco.EBR
			if err := ManejoArchivo.LeerObjeto(archivo, &EBRTemp, int64(EBRActual), buffer); err != nil {
				fmt.Fprintf(buffer, "Error leyendo EBR: %v\n", err)
				return
			}
			EstructuraDisco.PrintEBR(EBRTemp)
			if EBRTemp.PartNext == -1 {
				break
			}
			EBRActual = EBRTemp.PartNext
		}
		fmt.Println("---------------------------------------------")
	}
	if err := ManejoArchivo.EscribirObjeto(archivo, MBRTemporal, 0, buffer); err != nil {
		return
	}
	var TempMRB EstructuraDisco.MRB
	if err := ManejoArchivo.LeerObjeto(archivo, &TempMRB, 0, buffer); err != nil {
		return
	}
	fmt.Println("---------------------------------------------")
	EstructuraDisco.ImprimirMBR(TempMRB)
	fmt.Println("---------------------------------------------")
	defer archivo.Close()
}

func MOUNT(ruta string, nombre string, buffer *bytes.Buffer) {
	fmt.Fprintf(buffer, "MOUNT---------------------------------------------------------------------\n")
	fmt.Print(ruta)
	// Validar la ruta (path)
	if ruta == "" {
		fmt.Fprintf(buffer, "Error MOUNT: La ruta del disco es obligatoria.\n")
		return
	}
	// Validar el nombre (name)
	if nombre == "" {
		fmt.Fprintf(buffer, "Error MOUNT: El nombre de la partición es obligatorio.\n")
		return
	}
	// Abrir archivo binario
	archivo, err := ManejoArchivo.AbrirArchivo(ruta, buffer)
	if err != nil {
		return
	}
	var MBRTemporal EstructuraDisco.MRB
	if err := ManejoArchivo.LeerObjeto(archivo, &MBRTemporal, 0, buffer); err != nil {
		return
	}

	var ParticionExiste = false
	var Particion EstructuraDisco.Partition
	var IndiceParticion int

	NombreBytes := [16]byte{}
	copy(NombreBytes[:], []byte(nombre))

	for i := 0; i < 4; i++ {
		if MBRTemporal.Partitions[i].PartType[0] == 'e' && bytes.Equal(MBRTemporal.Partitions[i].PartName[:], NombreBytes[:]) {
			fmt.Fprintf(buffer, "Error MOUNT: No se puede montar una partición extendida.\n")
			return
		}
	}

	for i := 0; i < 4; i++ {
		if MBRTemporal.Partitions[i].PartType[0] == 'p' && bytes.Equal(MBRTemporal.Partitions[i].PartName[:], NombreBytes[:]) {
			Particion = MBRTemporal.Partitions[i]
			IndiceParticion = i
			ParticionExiste = true
			break
		}
	}

	if !ParticionExiste {
		fmt.Fprintf(buffer, "Error MOUNT: No se encontró la partición con el nombre especificado. Solo se pueden montar particiones primarias.\n")
		return
	}

	if Particion.PartStatus[0] == '1' {
		fmt.Fprintf(buffer, "Error MOUNT: La partición ya está montada.\n")
		return
	}

	DiscoID := GenerarDiscoID(ruta)
	ListaParticionesMontadasEnDisco := ListaParticionesMontadas[DiscoID]
	var Letra byte

	if len(ListaParticionesMontadasEnDisco) == 0 {
		if len(ListaParticionesMontadas) == 0 {
			Letra = 'a'
		} else {
			UltimoDiscoID := getLastDiskID()
			UltimaLetra := ListaParticionesMontadas[UltimoDiscoID][0].ID[len(ListaParticionesMontadas[UltimoDiscoID][0].ID)-1]
			Letra = UltimaLetra + 1
		}
	} else {
		Letra = ListaParticionesMontadasEnDisco[0].ID[len(ListaParticionesMontadasEnDisco[0].ID)-1]
	}
	var indice int

	carnet := "202201524"
	UltimosDigitos := carnet[len(carnet)-2:]
	indice = len(ListaParticionesMontadasEnDisco)
	IDParticion := fmt.Sprintf("%s%d%c", UltimosDigitos, indice+1, Letra)

	Particion.PartStatus[0] = '1'
	copy(Particion.PartId[:], IDParticion)
	MBRTemporal.Partitions[IndiceParticion] = Particion
	ListaParticionesMontadas[DiscoID] = append(ListaParticionesMontadas[DiscoID], ParticionMontada{
		Ruta:   ruta,
		Nombre: nombre,
		ID:     IDParticion,
		Estado: '1',
	})
	fmt.Fprintf(buffer, "Partición montada con éxito en la ruta: %s con el nombre: %s y ID: %s.\n", ruta, nombre, IDParticion)

	if err := ManejoArchivo.EscribirObjeto(archivo, MBRTemporal, 0, buffer); err != nil {
		return
	}
	fmt.Println("---------------------------------------------")
	PrintMountedPartitions(ruta, buffer)
	fmt.Println("---------------------------------------------")
	var TempMRB EstructuraDisco.MRB
	if err := ManejoArchivo.LeerObjeto(archivo, &TempMRB, 0, buffer); err != nil {
		return
	}
	EstructuraDisco.ImprimirMBR(TempMRB)
	fmt.Println("---------------------------------------------")
	defer archivo.Close()
}

func LIST(buffer *bytes.Buffer) {
	fmt.Fprintf(buffer, "LIST---------------------------------------------------------------------\n")
	if len(ListaParticionesMontadas) == 0 {
		fmt.Fprintf(buffer, "No hay particiones montadas.")
		return
	}
	for DiscoID, partitions := range ListaParticionesMontadas {
		fmt.Fprintf(buffer, "Disco: %s\n", DiscoID)
		fmt.Fprintf(buffer, "---------------------------\n")
		for _, Particion := range partitions {
			loginStatus := "No"
			if Particion.LoggedIn {
				loginStatus = "Sí"
			}
			fmt.Fprintf(buffer, "Nombre: %s, ID: %s, Ruta: %s, Estado: %c, LoggedIn: %s\n",
				Particion.Nombre, Particion.ID, Particion.Ruta, Particion.Estado, loginStatus)
		}
		fmt.Fprintf(buffer, "---------------------------\n")
	}
}

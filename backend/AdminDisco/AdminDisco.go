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

type PathDisk struct {
	Path string
}

var ListaParticionesMontadas = make(map[string][]ParticionMontada)

var ListaRutasDiscos = make(map[string][]PathDisk)

func AddDiskPath(path string) {
	ListaRutasDiscos[path] = append(ListaRutasDiscos[path], PathDisk{Path: path})
}

func DeleteDiskPath(path string) {
	delete(ListaRutasDiscos, path)
}

func ObtenerRutaDiscos(buffer *bytes.Buffer) {
	for _, rutas := range ListaRutasDiscos {
		for _, ruta := range rutas {
			fmt.Fprintf(buffer, "%s\n", ruta.Path)
		}
	}
}

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

// ------------------------------------------------------------------------------------------------------------------------------------
// Función para leer el MBR desde un archivo binario
func ReadMBR(path string, buffer *bytes.Buffer) {
	file, err := ManejoArchivo.AbrirArchivo(path, buffer)
	if err != nil {
		return
	}
	defer file.Close()

	// Crear una variable para almacenar el MBR
	var mbr EstructuraDisco.MRB

	// Leer el MBR desde el archivo
	err = ManejoArchivo.LeerObjeto(file, &mbr, 0, buffer) // Leer desde la posición 0
	if err != nil {
		return
	}
}

type PartitionInfo struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Start  int32  `json:"start"`
	Size   int32  `json:"size"`
	Status string `json:"status"`
}

func ListPartitions(path string) ([]PartitionInfo, error) {
	// Abrir el archivo binario
	var buffer *bytes.Buffer
	file, err := ManejoArchivo.AbrirArchivo(path, buffer)
	if err != nil {
		return nil, fmt.Errorf("error al abrir el archivo: %v", err)
	}
	defer file.Close()

	// Crear una variable para almacenar el MBR
	var mbr EstructuraDisco.MRB

	// Leer el MBR desde el archivo
	err = ManejoArchivo.LeerObjeto(file, &mbr, 0, buffer) // Leer desde la posición 0
	if err != nil {
		return nil, fmt.Errorf("error al leer el MBR: %v", err)
	}

	// Crear una lista de particiones basada en el MBR
	var partitions []PartitionInfo
	for _, partition := range mbr.Partitions {
		if partition.PartSize > 0 { // Solo agregar si la partición tiene un tamaño
			// Limpiar el nombre para eliminar caracteres nulos
			partitionName := strings.TrimRight(string(partition.PartName[:]), "\x00")

			partitions = append(partitions, PartitionInfo{
				Name:   partitionName,
				Type:   strings.TrimRight(string(partition.PartType[:]), "\x00"),
				Start:  partition.PartStart,
				Size:   partition.PartSize,
				Status: strings.TrimRight(string(partition.PartStatus[:]), "\x00"),
			})
		}
	}

	return partitions, nil
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
	TamanioBloque := 1024 * 1024
	BloqueCero := make([]byte, TamanioBloque)
	TamanioRestante := tamano
	for TamanioRestante > 0 {
		if TamanioRestante < TamanioBloque {
			BloqueCero = make([]byte, TamanioRestante)
		}
		_, err := archivo.Write(BloqueCero)
		if err != nil {
			fmt.Fprintf(buffer, "Error MKDISK: Error escribiendo ceros: %v.\n", err)
			return
		}
		TamanioRestante -= TamanioBloque
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
	// Agregar la ruta del disco a la lista de rutas
	AddDiskPath(ruta)
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
	DeleteDiskPath(ruta)
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

func UNMOUNT(id string, buffer *bytes.Buffer) {
	fmt.Fprintf(buffer, "UNMOUNT---------------------------------------------------------------------\n")

	if id == "" {
		fmt.Fprintf(buffer, "Error UNMOUNT: El ID de la partición es obligatorio.\n")
		return
	}

	var PariticionEncontrada *ParticionMontada
	var ID_Disco string
	var IndiceParticion int
	var Ruta string

	for disco, Particiones := range ListaParticionesMontadas {
		for i, partition := range Particiones {
			if partition.ID == id {
				PariticionEncontrada = &Particiones[i]
				ID_Disco = disco
				IndiceParticion = i
				Ruta = PariticionEncontrada.Ruta
				break
			}
		}
		if PariticionEncontrada != nil {
			break
		}
	}

	if PariticionEncontrada == nil {
		fmt.Fprintf(buffer, "Error: No se encontró una partición montada con el ID proporcionado: %s.\n", id)
		return
	}

	// Abrir el archivo del disco correspondiente
	file, err := ManejoArchivo.AbrirArchivo(PariticionEncontrada.Ruta, buffer)
	if err != nil {
		return
	}
	defer file.Close()

	// Leer el MBR
	var MBRTemporal EstructuraDisco.MRB
	if err := ManejoArchivo.LeerObjeto(file, &MBRTemporal, 0, buffer); err != nil {
		fmt.Println("Error: No se pudo leer el MBR desde el archivo")
		return
	}

	// Buscar la partición en el MBR utilizando el nombre
	NombreBytes := [16]byte{}
	copy(NombreBytes[:], []byte(PariticionEncontrada.Nombre))
	ParticionActualizada := false

	for i := 0; i < 4; i++ {
		if bytes.Equal(MBRTemporal.Partitions[i].PartName[:], NombreBytes[:]) {
			// Cambiar el estado de la partición de montada ('1') a desmontada ('0')
			MBRTemporal.Partitions[i].PartStatus[0] = '0'
			// Borrar el ID de la partición
			copy(MBRTemporal.Partitions[i].PartId[:], "")
			ParticionActualizada = true
			break
		}
	}

	if !ParticionActualizada {
		fmt.Fprintf(buffer, "Error: No se pudo encontrar la partición en el MBR para desmontar.\n")
		return
	}
	// Sobrescribir el MBR actualizado al archivo
	if err := ManejoArchivo.EscribirObjeto(file, MBRTemporal, 0, buffer); err != nil {
		return
	}
	// Eliminar la partición de la lista de particiones montadas
	ListaParticionesMontadas[ID_Disco] = append(ListaParticionesMontadas[ID_Disco][:IndiceParticion], ListaParticionesMontadas[ID_Disco][IndiceParticion+1:]...)

	// Si ya no hay particiones montadas en este disco, eliminar el disco de la lista
	if len(ListaParticionesMontadas[ID_Disco]) == 0 {
		delete(ListaParticionesMontadas, ID_Disco)
	}

	fmt.Fprintf(buffer, "Partición desmontada con éxito en la ruta: %s con el nombre: %s y ID: %s.\n", Ruta, PariticionEncontrada.Nombre, id)
	fmt.Println("---------------------------------------------")
	PrintMountedPartitions(Ruta, buffer)
	fmt.Println("---------------------------------------------")
}

func LIST(buffer *bytes.Buffer) {

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
	}
}

// Función para eliminar particiones
func ELIMINAR_PARTICION(path string, name string, delete string, buffer *bytes.Buffer) {
	fmt.Fprint(buffer, "FDISK DELETE---------------------------------------------------------------------\n")

	if delete == "" {
		fmt.Println("Error FDISK DELETE: Se debe establecer la configuración 'fast' o 'full'.")
		return
	}

	file, err := ManejoArchivo.AbrirArchivo(path, buffer)
	if err != nil {
		return
	}

	var MBRTemporal EstructuraDisco.MRB
	if err := ManejoArchivo.LeerObjeto(file, &MBRTemporal, 0, buffer); err != nil {
		return
	}

	ExisteParticion := false
	for i := 0; i < 4; i++ {
		NombreParticion := strings.TrimRight(string(MBRTemporal.Partitions[i].PartName[:]), "\x00")
		if NombreParticion == name {
			ExisteParticion = true
			// Si es una partición extendida, eliminar las particiones lógicas dentro de ella
			if MBRTemporal.Partitions[i].PartType[0] == 'e' {
				fmt.Println("Eliminando Particiones Lógicas Dentro De La Partición Extendida.")
				EBRPosterior := MBRTemporal.Partitions[i].PartStart
				var EBRActual EstructuraDisco.EBR
				for {
					err := ManejoArchivo.LeerObjeto(file, &EBRActual, int64(EBRPosterior), buffer)
					if err != nil {
						break
					}
					// Detener el bucle si el EBR está vacío
					if EBRActual.PartStart == 0 && EBRActual.PartSize == 0 {
						break
					}
					// Eliminar partición lógica
					if delete == "fast" {
						EBRActual = EstructuraDisco.EBR{}                                          // Resetear el EBR manualmente
						ManejoArchivo.EscribirObjeto(file, EBRActual, int64(EBRPosterior), buffer) // Sobrescribir el EBR reseteado
					} else if delete == "full" {
						ManejoArchivo.LlenarEspacioConCeros(file, EBRActual.PartStart, EBRActual.PartSize, buffer)
						EBRActual = EstructuraDisco.EBR{}                                          // Resetear el EBR manualmente
						ManejoArchivo.EscribirObjeto(file, EBRActual, int64(EBRPosterior), buffer) // Sobrescribir el EBR reseteado
					}
					if EBRActual.PartNext == -1 {
						break
					}
					EBRPosterior = EBRActual.PartNext
				}
			}
			// Proceder a eliminar la partición (extendida o primaria)
			if delete == "fast" {
				MBRTemporal.Partitions[i] = EstructuraDisco.Partition{} // Resetear la partición manualmente
				fmt.Fprintf(buffer, "Partición eliminada correctamente en modo Fast.\n")
			} else if delete == "full" {
				start := MBRTemporal.Partitions[i].PartStart
				size := MBRTemporal.Partitions[i].PartSize
				MBRTemporal.Partitions[i] = EstructuraDisco.Partition{} // Resetear la partición manualmente
				ManejoArchivo.LlenarEspacioConCeros(file, start, size, buffer)
				ManejoArchivo.VerificarCeros(file, start, size, buffer)
				fmt.Fprintf(buffer, "Partición eliminada correctamente en modo Full.\n")
			}
			break
		}
	}

	if !ExisteParticion {
		fmt.Println("Buscando En Particiones Lógicas Dentro De Las Extendidas.")
		for i := 0; i < 4; i++ {
			if MBRTemporal.Partitions[i].PartType[0] == 'e' {
				EBRPosterior := MBRTemporal.Partitions[i].PartStart
				var EBRTemporalA EstructuraDisco.EBR
				for {
					err := ManejoArchivo.LeerObjeto(file, &EBRTemporalA, int64(EBRPosterior), buffer)
					if err != nil {
						break
					}
					NombreParticionLogica := strings.TrimRight(string(EBRTemporalA.PartName[:]), "\x00")
					if NombreParticionLogica == name {
						ExisteParticion = true
						if delete == "fast" {
							EBRTemporalA = EstructuraDisco.EBR{}                                          // Resetear el EBR manualmente
							ManejoArchivo.EscribirObjeto(file, EBRTemporalA, int64(EBRPosterior), buffer) // Sobrescribir el EBR reseteado
							fmt.Fprintf(buffer, "Partición lógica eliminada correctamente en modo Fast.\n")
						} else if delete == "full" {
							ManejoArchivo.LlenarEspacioConCeros(file, EBRTemporalA.PartStart, EBRTemporalA.PartSize, buffer)
							EBRTemporalA = EstructuraDisco.EBR{}                                          // Resetear el EBR manualmente
							ManejoArchivo.EscribirObjeto(file, EBRTemporalA, int64(EBRPosterior), buffer) // Sobrescribir el EBR reseteado
							ManejoArchivo.VerificarCeros(file, EBRTemporalA.PartStart, EBRTemporalA.PartSize, buffer)
							fmt.Fprintf(buffer, "Partición lógica eliminada correctamente en modo Full.\n")
						}
						break
					}
					if EBRTemporalA.PartNext == -1 {
						break
					}
					EBRPosterior = EBRTemporalA.PartNext
				}
			}
			if ExisteParticion {
				break
			}
		}
	}

	if !ExisteParticion {
		fmt.Fprintf(buffer, "Error FDISK DELETE: No se encontró la partición con el nombre: %s\n", name)
		return
	}
	if err := ManejoArchivo.EscribirObjeto(file, MBRTemporal, 0, buffer); err != nil {
		return
	}
	fmt.Println("--------------------------------------------------------------------")
	fmt.Println("MBR Y EBR Actualizado Después De La Eliminación:")
	EstructuraDisco.ImprimirMBR(MBRTemporal)
	fmt.Println("--------------------------------------------------------------------")
	for i := 0; i < 4; i++ {
		if MBRTemporal.Partitions[i].PartType[0] == 'e' {
			EBRPosterior := MBRTemporal.Partitions[i].PartStart
			var EBRTemporalA EstructuraDisco.EBR
			for {
				err := ManejoArchivo.LeerObjeto(file, &EBRTemporalA, int64(EBRPosterior), buffer)
				if err != nil {
					break
				}
				if EBRTemporalA.PartStart == 0 && EBRTemporalA.PartSize == 0 {
					break
				}
				EstructuraDisco.PrintEBR(EBRTemporalA)
				if EBRTemporalA.PartNext == -1 {
					break
				}
				EBRPosterior = EBRTemporalA.PartNext
			}
		}
	}
	fmt.Println("--------------------------------------------------------------------")
	defer file.Close()
}

func ADD_PARTICION(path string, name string, add int, unit string, buffer *bytes.Buffer) error {
	fmt.Fprint(buffer, "FDISK ADD---------------------------------------------------------------------\n")

	if add == 0 {
		fmt.Fprintf(buffer, "Error FDISK ADD: El tamaño a agregar debe ser distinto que 0.\n")
		return nil
	}

	file, err := ManejoArchivo.AbrirArchivo(path, buffer)
	if err != nil {
		return err
	}
	defer file.Close()

	var MbrTemporal EstructuraDisco.MRB
	if err := ManejoArchivo.LeerObjeto(file, &MbrTemporal, 0, buffer); err != nil {
		return err
	}

	var ParticionEncontrada *EstructuraDisco.Partition
	var TipoParticion byte

	// Revisar si la partición es primaria o extendida
	for i := 0; i < 4; i++ {
		NombreParticion := strings.TrimRight(string(MbrTemporal.Partitions[i].PartName[:]), "\x00")
		if NombreParticion == name {
			ParticionEncontrada = &MbrTemporal.Partitions[i]
			TipoParticion = MbrTemporal.Partitions[i].PartType[0]
			break
		}
	}

	// Si no se encuentra en las primarias/extendidas, buscar en las particiones lógicas
	if ParticionEncontrada == nil {
		for i := 0; i < 4; i++ {
			if MbrTemporal.Partitions[i].PartType[0] == 'e' {
				EBRPosterior := MbrTemporal.Partitions[i].PartStart
				var EBRTemporal1 EstructuraDisco.EBR
				for {
					if err := ManejoArchivo.LeerObjeto(file, &EBRTemporal1, int64(EBRPosterior), buffer); err != nil {
						return err
					}
					EBRNombreParticion := strings.TrimRight(string(EBRTemporal1.PartName[:]), "\x00")
					if EBRNombreParticion == name {
						TipoParticion = 'l'
						ParticionEncontrada = &EstructuraDisco.Partition{
							PartStart: EBRTemporal1.PartStart,
							PartSize:  EBRTemporal1.PartSize,
						}
						break
					}
					if EBRTemporal1.PartNext == -1 {
						break
					}
					EBRPosterior = EBRTemporal1.PartNext
				}
				if ParticionEncontrada != nil {
					break
				}
			}
		}
	}

	if ParticionEncontrada == nil {
		fmt.Fprintf(buffer, "Error FDISK ADD: No se encontró la partición con el nombre: %s.\n", name)
		return nil
	}

	var BytesAgregarELiminar int
	if unit == "k" {
		BytesAgregarELiminar = add * 1024
	} else if unit == "m" {
		BytesAgregarELiminar = add * 1024 * 1024
	} else {
		fmt.Fprintf(buffer, "Error FDISK ADD: Unidad desconocida, debe ser 'K' o 'M'.\n")
		return nil
	}

	var DeberiaModificar = true

	// Comprobar si es posible agregar o quitar espacio
	if add > 0 {
		// Agregar espacio: verificar si hay suficiente espacio libre después de la partición
		nextPartitionStart := ParticionEncontrada.PartStart + ParticionEncontrada.PartSize
		if TipoParticion == 'l' {
			// Para particiones lógicas, verificar con el siguiente EBR o el final de la partición extendida
			for i := 0; i < 4; i++ {
				if MbrTemporal.Partitions[i].PartType[0] == 'e' {
					extendedPartitionEnd := MbrTemporal.Partitions[i].PartStart + MbrTemporal.Partitions[i].PartSize
					if nextPartitionStart+int32(BytesAgregarELiminar) > extendedPartitionEnd {
						fmt.Fprintf(buffer, "Error FDISK ADD: No hay suficiente espacio libre dentro de la partición extendida.\n")
						DeberiaModificar = false
					}
					break
				}
			}
		} else {
			if nextPartitionStart+int32(BytesAgregarELiminar) > MbrTemporal.MbrTamano {
				fmt.Fprintf(buffer, "Error FDISK ADD: No hay suficiente espacio libre después de la partición.\n")
				DeberiaModificar = false
			}
		}
	} else {
		// Quitar espacio: verificar que no se reduzca el tamaño por debajo de 0
		if ParticionEncontrada.PartSize+int32(BytesAgregarELiminar) < 0 {
			fmt.Fprintf(buffer, "Error FDISK ADD: No es posible reducir la partición por debajo de %d.\n", add)
			DeberiaModificar = false
		}
	}

	// Solo modificar si no hay errores
	if DeberiaModificar {
		ParticionEncontrada.PartSize += int32(BytesAgregarELiminar)
		fmt.Fprintf(buffer, "Tamaño de la partición modificado con éxito en la ruta: %s con el nombre: %s.\n", path, name)
	} else {
		fmt.Fprintf(buffer, "Error FDISK ADD: No se realizaron modificaciones debido a un error.\n")
		return nil
	}

	// Si es una partición lógica, sobrescribir el EBR
	if TipoParticion == 'l' {
		EBRPosterior := ParticionEncontrada.PartStart
		var EBRTemporal2 EstructuraDisco.EBR
		if err := ManejoArchivo.LeerObjeto(file, &EBRTemporal2, int64(EBRPosterior), buffer); err != nil {
			return err
		}
		EBRTemporal2.PartSize = ParticionEncontrada.PartSize
		if err := ManejoArchivo.EscribirObjeto(file, EBRTemporal2, int64(EBRPosterior), buffer); err != nil {
			return err
		}

	}

	if err := ManejoArchivo.EscribirObjeto(file, MbrTemporal, 0, buffer); err != nil {
		fmt.Println("Error al escribir el MBR actualizado:", err)
		return err
	}

	// Imprimir el MBR modificado
	fmt.Println("--------------------------------------------------------------------")
	fmt.Println("MBR Y EBR Actualizado Después De La Eliminación:")
	EstructuraDisco.ImprimirMBR(MbrTemporal)
	fmt.Println("--------------------------------------------------------------------")
	for i := 0; i < 4; i++ {
		if MbrTemporal.Partitions[i].PartType[0] == 'e' {
			EBRPosterior := MbrTemporal.Partitions[i].PartStart
			var EBRTemporalA EstructuraDisco.EBR
			for {
				err := ManejoArchivo.LeerObjeto(file, &EBRTemporalA, int64(EBRPosterior), buffer)
				if err != nil {
					break
				}
				if EBRTemporalA.PartStart == 0 && EBRTemporalA.PartSize == 0 {
					break
				}
				EstructuraDisco.PrintEBR(EBRTemporalA)
				if EBRTemporalA.PartNext == -1 {
					break
				}
				EBRPosterior = EBRTemporalA.PartNext
			}
		}
	}
	fmt.Println("--------------------------------------------------------------------")

	return nil
}

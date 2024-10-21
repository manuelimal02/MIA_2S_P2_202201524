package ManejoArchivo

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
)

// Función para crear un archivo binario
func CrearArchivo(nombre string, buffer *bytes.Buffer) error {
	directorio := filepath.Dir(nombre)
	if err := os.MkdirAll(directorio, os.ModePerm); err != nil {
		fmt.Fprintf(buffer, "Error en CrearArchivo (directorio): %v.\n", err)
		return err
	}

	if _, err := os.Stat(nombre); os.IsNotExist(err) {
		archivo, err := os.Create(nombre)
		if err != nil {
			fmt.Fprintf(buffer, "Error en CrearArchivo (creación): %v.\n", err)
			return err
		}
		defer archivo.Close()
	}
	return nil
}

func EliminarArchivo(nombre string, buffer *bytes.Buffer) error {
	if _, err := os.Stat(nombre); os.IsNotExist(err) {
		fmt.Fprintf(buffer, "Error: El archivo no existe: %v.\n", err)
		return err
	}
	err := os.Remove(nombre)
	if err != nil {
		fmt.Fprintf(buffer, "Error al eliminar el archivo: %v.\n", err)
		return err
	}
	return nil
}

// Función para abrir un archivo binario en modo lectura/escritura
func AbrirArchivo(nombre string, buffer *bytes.Buffer) (*os.File, error) {
	archivo, err := os.OpenFile(nombre, os.O_RDWR, 0644)
	if err != nil {
		fmt.Fprintf(buffer, "Error en AbrirArchivo: %v.\n", err)
		return nil, err
	}
	return archivo, nil
}

// Función para escribir un objeto en un archivo binario
func EscribirObjeto(archivo *os.File, datos interface{}, posicion int64, buffer *bytes.Buffer) error {
	archivo.Seek(posicion, 0)
	err := binary.Write(archivo, binary.LittleEndian, datos)
	if err != nil {
		fmt.Fprintf(buffer, "Error en EscribirObjeto: %v.\n", err)
		return err
	}
	return nil
}

// Función para leer un objeto de un archivo binario
func LeerObjeto(archivo *os.File, datos interface{}, posicion int64, buffer *bytes.Buffer) error {
	archivo.Seek(posicion, 0)
	err := binary.Read(archivo, binary.LittleEndian, datos)
	if err != nil {
		fmt.Fprintf(buffer, "Error en LeerObjeto: %v.\n", err)
		return err
	}
	return nil
}

// Función para llenar el espacio con ceros (\0)
func LlenarEspacioConCeros(file *os.File, start int32, size int32, bufferError *bytes.Buffer) error {
	// Posiciona el archivo al inicio del área que debe ser llenada
	file.Seek(int64(start), 0)
	// Crear un buffer lleno de ceros
	buffer := make([]byte, size)
	// Escribir los ceros en el archivo
	_, err := file.Write(buffer)
	if err != nil {
		fmt.Fprintf(bufferError, "Error al llenar el espacio con ceros: %v.\n", err)
		return err
	}
	fmt.Fprintf(bufferError, "Espacio llenado con ceros desde el byte %d por %d bytes.\n", start, size)
	return nil
}

// Función para verificar que un bloque del archivo esté lleno de ceros
func VerificarCeros(file *os.File, start int32, size int32, buffer *bytes.Buffer) {
	Cero := make([]byte, size)
	_, err := file.ReadAt(Cero, int64(start))
	if err != nil {
		fmt.Fprintf(buffer, "Error al leer la sección eliminada: %v.\n", err)
		return
	}
	// Verificar si todos los bytes leídos son ceros
	LlenoDeCeros := true
	for _, b := range Cero {
		if b != 0 {
			LlenoDeCeros = false
			break
		}
	}
	if LlenoDeCeros {
		fmt.Fprintf(buffer, "La partición eliminada está completamente llena de ceros.\n")
	} else {
		fmt.Fprintf(buffer, "Advertencia: La partición eliminada no está completamente llena de ceros.\n")
	}
}

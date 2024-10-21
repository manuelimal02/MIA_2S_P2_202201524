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

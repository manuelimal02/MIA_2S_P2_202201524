package Analizador

import (
	"Proyecto1/AdminDisco"
	"Proyecto1/AdminRoot"
	"Proyecto1/Reporte"
	"Proyecto1/SistemaDeArchivos"
	"Proyecto1/Usuario"
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var re = regexp.MustCompile(`-(\w+)=("[^"]+"|\S+)`)

func Analizar(texto string) string {
	var buffer bytes.Buffer
	scanner := bufio.NewScanner(strings.NewReader(texto))
	for scanner.Scan() {
		entrada := scanner.Text()
		if len(entrada) == 0 || entrada[0] == '#' {
			fmt.Fprintf(&buffer, "%s\n", entrada)
			continue
		}
		entrada = strings.TrimSpace(entrada)
		comando, parametros := comando_parametro(entrada)
		AnalizarComando(comando, parametros, &buffer)
	}
	return buffer.String()
}

func comando_parametro(entrada string) (string, string) {
	partes := strings.Fields(entrada)
	if len(partes) > 0 {
		comando := strings.ToLower(partes[0])
		for i := 1; i < len(partes); i++ {
			partes[i] = strings.ToLower(partes[i])
		}
		parametros := strings.Join(partes[1:], " ")
		return comando, parametros
	}
	return "", entrada
}

func AnalizarComando(comando string, parametros string, buffer io.Writer) {
	if strings.Contains(comando, "mkdisk") {
		comando_mkdisk(parametros, buffer)
	} else if strings.Contains(comando, "rmdisk") {
		comando_rmdisk(parametros, buffer)
	} else if strings.Contains(comando, "fdisk") {
		comando_fdisk(parametros, buffer)
	} else if strings.Contains(comando, "unmount") {
		comando_unmount(parametros, buffer)
	} else if strings.Contains(comando, "mount") {
		comando_mount(parametros, buffer)
	} else if strings.Contains(comando, "mkfs") {
		comando_mkfs(parametros, buffer)
	} else if strings.Contains(comando, "login") {
		comando_login(parametros, buffer)
	} else if strings.Contains(comando, "logout") {
		comando_logout(parametros, buffer)
	} else if strings.Contains(comando, "mkgrp") {
		comando_mkgrp(parametros, buffer)
	} else if strings.Contains(comando, "rmgrp") {
		//comando_rmgrp(parametros, buffer)
	} else if strings.Contains(comando, "mkusr") {
		comando_mkusr(parametros, buffer)
	} else if strings.Contains(comando, "rmusr") {
		//comando_rmusr(parametros, buffer)
	} else if strings.Contains(comando, "chgrp") {
		//comando_chgrp(parametros, buffer)
	} else if strings.Contains(comando, "cat") {
		comando_cat(parametros, buffer)
	} else if strings.Contains(comando, "rep") {
		comando_rep(parametros, buffer)
	} else if strings.Contains(comando, "list") {
		comando_list(parametros, buffer)
	} else {
		fmt.Fprintf(buffer, "Error: Comando No Encontrado.\n")
	}
}

// Función para ejecutar el comando MKDISK
func comando_mkdisk(params string, buffer io.Writer) {
	fs := flag.NewFlagSet("mkdisk", flag.ExitOnError)
	tamano := fs.Int("size", 0, "size")
	ajuste := fs.String("fit", "ff", "fit")
	unidad := fs.String("unit", "m", "unit")
	ruta := fs.String("path", "", "path")

	fs.Parse(os.Args[1:])
	matches := re.FindAllStringSubmatch(params, -1)

	for _, match := range matches {
		nombreFlag := match[1]
		valorFlag := strings.ToLower(match[2])
		valorFlag = strings.Trim(valorFlag, "\"")
		switch nombreFlag {
		case "size", "fit", "unit", "path":
			fs.Set(nombreFlag, valorFlag)
		default:
			fmt.Fprintf(buffer, "Error: El comando 'MKDISK' incluye parámetros no asociados.\n")
			return
		}
	}
	fs.Parse([]string{})
	AdminDisco.MKDISK(*tamano, *ajuste, *unidad, *ruta, buffer.(*bytes.Buffer))
}

// Función para ejecutar el comando RMDISK
func comando_rmdisk(params string, buffer io.Writer) {
	fs := flag.NewFlagSet("rmdisk", flag.ExitOnError)
	ruta := fs.String("path", "", "Ruta")

	fs.Parse(os.Args[1:])
	matches := re.FindAllStringSubmatch(params, -1)

	for _, match := range matches {
		nombreFlag := match[1]
		valorFlag := strings.ToLower(match[2])
		valorFlag = strings.Trim(valorFlag, "\"")
		switch nombreFlag {
		case "path":
			fs.Set(nombreFlag, valorFlag)
		default:
			fmt.Fprintf(buffer, "Error: El comando 'RMDISK' incluye parámetros no asociados.\n")
			return
		}
	}
	AdminDisco.RMDISK(*ruta, buffer.(*bytes.Buffer))
}

// Función para ejecutar el comando FDISK
func comando_fdisk(entrada string, buffer io.Writer) {
	fs := flag.NewFlagSet("fdisk", flag.ExitOnError)
	tamano := fs.Int("size", 0, "Tamaño")
	unidad := fs.String("unit", "k", "Unidad")
	ruta := fs.String("path", "", "Ruta")
	tipo := fs.String("type", "p", "Tipo")
	ajuste := fs.String("fit", "wf", "Ajuste")
	nombre := fs.String("name", "", "Nombre")
	eliminar := fs.String("delete", "", "Eliminar")
	agregar := fs.String("add", "", "Tamaño")

	fs.Parse(os.Args[1:])
	matches := re.FindAllStringSubmatch(entrada, -1)

	for _, match := range matches {
		nombreFlag := match[1]
		valorFlag := strings.ToLower(match[2])

		valorFlag = strings.Trim(valorFlag, "\"")

		switch nombreFlag {
		case "size", "unit", "path", "type", "fit", "name", "delete", "add":
			fs.Set(nombreFlag, valorFlag)
		default:
			fmt.Fprintf(buffer, "Error: El comando 'FDISK' incluye parámetros no asociados.\n")
			return
		}
	}

	if *agregar != "" {
		var Add int
		Add, err := strconv.Atoi(*agregar)
		if err != nil {
			fmt.Fprintf(buffer, "Error: No se pudo convertir el valor de 'add' a entero.\n")
			return
		}
		if *ruta == "" || *nombre == "" {
			fmt.Fprint(buffer, "Error FDISK ADD: Para agregar una partición, se requiere 'unit', 'ruta' y 'nombre de la partición'.\n")
			return
		}
		fmt.Fprintf(buffer, "Tamano A Agregar O Eliminar: %d\n", Add)
		AdminDisco.ADD_PARTICION(*ruta, *nombre, Add, *unidad, buffer.(*bytes.Buffer))
		return
	}

	if *eliminar != "" {
		if *ruta == "" || *nombre == "" {
			fmt.Println("Error FDISK DELETE: Para eliminar una partición, se requiere 'ruta' y 'nombre de la partición'.")
			return
		}
		AdminDisco.ELIMINAR_PARTICION(*ruta, *nombre, *eliminar, buffer.(*bytes.Buffer))
		return
	}
	AdminDisco.FDISK(*tamano, *unidad, *ruta, *tipo, *ajuste, *nombre, buffer.(*bytes.Buffer))
}

// Función para ejecutar el comando MOUNT
func comando_mount(entrada string, buffer io.Writer) {
	fs := flag.NewFlagSet("mount", flag.ExitOnError)
	ruta := fs.String("path", "", "Ruta")
	nombre := fs.String("name", "", "Nombre")

	fs.Parse(os.Args[1:])
	matches := re.FindAllStringSubmatch(entrada, -1)

	for _, match := range matches {
		nombreFlag := match[1]
		valorFlag := strings.ToLower(match[2])

		valorFlag = strings.Trim(valorFlag, "\"")

		switch nombreFlag {
		case "path", "name":
			fs.Set(nombreFlag, valorFlag)
		default:
			fmt.Fprintf(buffer, "Error: El comando 'MOUNT' incluye parámetros no asociados.\n")
			return
		}
	}
	AdminDisco.MOUNT(*ruta, *nombre, buffer.(*bytes.Buffer))
}

// Función para ejecutar el comando LIST
func comando_list(entrada string, buffer io.Writer) {
	entrada = strings.TrimSpace(entrada)
	if len(entrada) > 0 {
		fmt.Fprintf(buffer, "Error: El comando 'LIST' incluye parámetros no asociados.\n")
		return
	}
	AdminDisco.LIST(buffer.(*bytes.Buffer))
}

// Función para ejecutar el comando REP
func comando_rep(entrada string, buffer io.Writer) {
	fs := flag.NewFlagSet("rep", flag.ExitOnError)
	nombre := fs.String("name", "", "Nombre")
	ruta := fs.String("path", "full", "Ruta")
	ID := fs.String("id", "", "IDParticion")
	path_file_ls := fs.String("path_file_l", "", "PathFile")

	fs.Parse(os.Args[1:])
	matches := re.FindAllStringSubmatch(entrada, -1)

	for _, match := range matches {
		nombreFlag := match[1]
		valorFlag := strings.ToLower(match[2])

		valorFlag = strings.Trim(valorFlag, "\"")

		switch nombreFlag {
		case "name", "path", "id", "path_file_l":
			fs.Set(nombreFlag, valorFlag)
		default:
			fmt.Fprintf(buffer, "Error: El comando 'REP' incluye parámetros no asociados.\n")
			return
		}
	}
	Reporte.Rep(*nombre, *ruta, *ID, *path_file_ls, buffer.(*bytes.Buffer))
}

func comando_mkfs(entrada string, buffer io.Writer) {
	fs := flag.NewFlagSet("mkfs", flag.ExitOnError)
	id := fs.String("id", "", "id")
	tipo := fs.String("type", "full", "tipo")
	sistema := fs.String("fs", "", "sistema")

	fs.Parse(os.Args[1:])
	matches := re.FindAllStringSubmatch(entrada, -1)

	for _, match := range matches {
		nombreFlag := match[1]
		valorFlag := strings.ToLower(match[2])

		valorFlag = strings.Trim(valorFlag, "\"")

		switch nombreFlag {
		case "type", "id", "fs":
			fs.Set(nombreFlag, valorFlag)
		default:
			fmt.Fprintf(buffer, "Error: El comando 'MKFS' incluye parámetros no asociados.\n")
			return
		}
	}
	SistemaDeArchivos.MKFS(*id, *tipo, *sistema, buffer.(*bytes.Buffer))
}

func comando_cat(params string, buffer io.Writer) {
	files := make(map[int]string)
	matches := re.FindAllStringSubmatch(params, -1)

	for _, match := range matches {
		flagName := match[1]
		flagValue := strings.ToLower(match[2])
		flagValue = strings.Trim(flagValue, "\"")

		if strings.HasPrefix(flagName, "file") {
			fileNumber, err := strconv.Atoi(strings.TrimPrefix(flagName, "file"))
			if err != nil {
				fmt.Fprint(buffer, "Error CAT: Nombre de archivo inválido.\n")
				return
			}
			if flagValue == "" {
				fmt.Fprintf(buffer, "Error CAT: EL Parametro: %d no contiene ninguna ruta.\n", fileNumber)
				return
			}

			files[fileNumber] = flagValue
		} else {
			fmt.Fprintf(buffer, "Error CAT: El comando 'CAT' incluye parámetros no asociados.\n")
			return
		}
	}

	var orderedFiles []string
	for i := 1; i <= len(files); i++ {
		if file, exists := files[i]; exists {
			orderedFiles = append(orderedFiles, file)
		} else {
			fmt.Fprintf(buffer, "Error CAT: Falta un archivo en la secuencia.\n")
			return
		}
	}

	if len(orderedFiles) == 0 {
		fmt.Fprintf(buffer, "Error CAT: No se encontraron archivos\n")
		return
	}

	SistemaDeArchivos.CAT(orderedFiles, buffer.(*bytes.Buffer))
}

// Función para ejecutar el comando LOGIN
func comando_login(input string, buffer io.Writer) {
	fs := flag.NewFlagSet("login", flag.ExitOnError)
	user := fs.String("user", "", "Usuario")
	pass := fs.String("pass", "", "Contraseña")
	id := fs.String("id", "", "Id")

	fs.Parse(os.Args[1:])
	matches := re.FindAllStringSubmatch(input, -1)

	for _, match := range matches {
		flagName := match[1]
		flagValue := match[2]

		flagValue = strings.Trim(flagValue, "\"")

		switch flagName {
		case "user", "pass", "id":
			fs.Set(flagName, flagValue)
		default:
			fmt.Println("Error: Flag not found")
		}
	}

	if *user == "" || *pass == "" || *id == "" {
		fmt.Fprintf(buffer, "Error: Faltan parámetros obligatorios para el comando 'LOGIN'.\n")
		return
	}

	Usuario.LOGIN(*user, *pass, *id, buffer.(*bytes.Buffer))

}

// Función para ejecutar el comando LOGOUT
func comando_logout(entrada string, buffer io.Writer) {
	entrada = strings.TrimSpace(entrada)
	if len(entrada) > 0 {
		fmt.Fprintf(buffer, "Error: El comando 'LOGOUT' incluye parámetros no asociados.\n")
		return
	}
	Usuario.LOGOUT(buffer.(*bytes.Buffer))
}

func comando_mkgrp(entrada string, buffer io.Writer) {
	fs := flag.NewFlagSet("mkgrp ", flag.ExitOnError)
	nombre := fs.String("name", "", "Nombre")

	fs.Parse(os.Args[1:])
	matches := re.FindAllStringSubmatch(entrada, -1)

	for _, match := range matches {
		nombreFlag := match[1]
		valorFlag := strings.ToLower(match[2])

		valorFlag = strings.Trim(valorFlag, "\"")

		switch nombreFlag {
		case "name":
			fs.Set(nombreFlag, valorFlag)
		default:
			fmt.Fprintf(buffer, "Error: El comando 'MKGRP' incluye parámetros no asociados.\n")
			return
		}
	}
	if *nombre == "" {
		fmt.Fprintf(buffer, "Error: Faltan parámetros obligatorios para el comando 'MKGRP'.\n")
		return
	}
	AdminRoot.Mkgrp(*nombre, buffer.(*bytes.Buffer))
}

/*
func comando_rmgrp(entrada string, buffer io.Writer) {
	fs := flag.NewFlagSet("rmgrp ", flag.ExitOnError)
	nombre := fs.String("name", "", "Nombre")

	fs.Parse(os.Args[1:])
	matches := re.FindAllStringSubmatch(entrada, -1)

	for _, match := range matches {
		nombreFlag := match[1]
		valorFlag := strings.ToLower(match[2])

		valorFlag = strings.Trim(valorFlag, "\"")

		switch nombreFlag {
		case "name":
			fs.Set(nombreFlag, valorFlag)
		default:
			fmt.Fprintf(buffer, "Error: El comando 'RMGRP' incluye parámetros no asociados.\n")
			return
		}
	}
	if *nombre == "" {
		fmt.Fprintf(buffer, "Error: Faltan parámetros obligatorios para el comando 'RMGRP'.\n")
		return
	}
	AdminRoot.Rmgrp(*nombre, buffer.(*bytes.Buffer))
}*/

func comando_mkusr(entrada string, buffer io.Writer) {
	fs := flag.NewFlagSet("mkusr ", flag.ExitOnError)
	nombre := fs.String("user", "", "Usuario")
	pass := fs.String("pass", "", "Contrasena")
	grp := fs.String("grp", "", "Grupo")

	fs.Parse(os.Args[1:])
	matches := re.FindAllStringSubmatch(entrada, -1)

	for _, match := range matches {
		nombreFlag := match[1]
		valorFlag := strings.ToLower(match[2])

		valorFlag = strings.Trim(valorFlag, "\"")

		switch nombreFlag {
		case "user", "pass", "grp":
			fs.Set(nombreFlag, valorFlag)
		default:
			fmt.Fprintf(buffer, "Error: El comando 'MKUSR' incluye parámetros no asociados.\n")
			return
		}
	}
	if *nombre == "" || *pass == "" || *grp == "" {
		fmt.Fprintf(buffer, "Error: Faltan parámetros obligatorios para el comando 'MKUSR'.\n")
		return
	}
	AdminRoot.Mkusr(*nombre, *pass, *grp, buffer.(*bytes.Buffer))
}

/*
func comando_rmusr(entrada string, buffer io.Writer) {
	fs := flag.NewFlagSet("rmusr ", flag.ExitOnError)
	nombre := fs.String("user", "", "usuario")

	fs.Parse(os.Args[1:])
	matches := re.FindAllStringSubmatch(entrada, -1)

	for _, match := range matches {
		nombreFlag := match[1]
		valorFlag := strings.ToLower(match[2])

		valorFlag = strings.Trim(valorFlag, "\"")

		switch nombreFlag {
		case "user":
			fs.Set(nombreFlag, valorFlag)
		default:
			fmt.Fprintf(buffer, "Error: El comando 'RMUSR' incluye parámetros no asociados.\n")
			return
		}
	}
	if *nombre == "" {
		fmt.Fprintf(buffer, "Error: Faltan parámetros obligatorios para el comando 'RMUSR'.\n")
		return
	}
	AdminRoot.Rmusr(*nombre, buffer.(*bytes.Buffer))
}*/
/*
func comando_chgrp(entrada string, buffer io.Writer) {
	fs := flag.NewFlagSet("chgrp ", flag.ExitOnError)
	usr := fs.String("user", "", "Usuario")
	grp := fs.String("grp", "", "Grupo")

	fs.Parse(os.Args[1:])
	matches := re.FindAllStringSubmatch(entrada, -1)

	for _, match := range matches {
		nombreFlag := match[1]
		valorFlag := strings.ToLower(match[2])

		valorFlag = strings.Trim(valorFlag, "\"")

		switch nombreFlag {
		case "user", "grp":
			fs.Set(nombreFlag, valorFlag)
		default:
			fmt.Fprintf(buffer, "Error: El comando 'CHGRP' incluye parámetros no asociados.\n")
			return
		}
	}
	if *usr == "" || *grp == "" {
		fmt.Fprintf(buffer, "Error: Faltan parámetros obligatorios para el comando 'CHGRP'.\n")
		return
	}
	AdminRoot.Chgrp(*usr, *grp, buffer.(*bytes.Buffer))
}
*/

func comando_unmount(entrada string, buffer io.Writer) {
	fs := flag.NewFlagSet("unmount", flag.ExitOnError)
	id := fs.String("id", "", "ID")

	fs.Parse(os.Args[1:])
	matches := re.FindAllStringSubmatch(entrada, -1)

	for _, match := range matches {
		nombreFlag := match[1]
		valorFlag := strings.ToLower(match[2])

		valorFlag = strings.Trim(valorFlag, "\"")

		switch nombreFlag {
		case "id":
			fs.Set(nombreFlag, valorFlag)
		default:
			fmt.Fprintf(buffer, "Error: El comando 'UNMOUNT' incluye parámetros no asociados.\n")
			return
		}
	}
	AdminDisco.UNMOUNT(*id, buffer.(*bytes.Buffer))
}

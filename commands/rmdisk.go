package comandos

import (
	"fmt"
	"os"
	"strings"
	"regexp"
	"github.com/melgxrga/proyecto1Archivos/consola"
)

type ParametrosRmdisk struct {
	Path string
}

type Rmdisk struct {
	Params ParametrosRmdisk
}

func (r *Rmdisk) Exe(parametros []string) {
	r.Params = r.SaveParams(parametros)

	if r.Rmdisk(r.Params.Path) {
		consola.AddToConsole(fmt.Sprintf("\nrmdisk realizado con éxito para la ruta: %s\n\n", r.Params.Path))
	} else {
		consola.AddToConsole(fmt.Sprintf("\n[ERROR!] No se logró realizar el comando rmdisk para la ruta: %s\n\n", r.Params.Path))
	}
}


func (r *Rmdisk) SaveParams(parametros []string) ParametrosRmdisk {
	var params ParametrosRmdisk
	// Unir todos los parámetros en una sola cadena
	args := strings.Join(parametros, " ")

	// Expresión regular para capturar el parámetro -path
	re := regexp.MustCompile(`-path="[^"]+"|-path=[^\s]+`)
	matches := re.FindAllString(args, -1)

	// Iterar sobre cada coincidencia
	for _, match := range matches {
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			fmt.Printf("Formato de parámetro inválido: %s\n", match)
			continue
		}
		key, value := strings.ToLower(kv[0]), kv[1]

		// Quitar comillas si las tiene
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		// Procesar el parámetro encontrado
		switch key {
		case "-path":
			if value == "" {
				fmt.Println("Error: el path no puede estar vacío")
				continue
			}
			params.Path = value
		default:
			fmt.Printf("Parámetro desconocido: %s\n", key)
		}
	}

	// Validación final de los parámetros obligatorios
	if params.Path == "" {
		fmt.Println("Error: Falta el parámetro obligatorio -path")
	}

	return params
}


func (r *Rmdisk) Rmdisk(path string) bool {
	// Comprobando si existe una ruta valida para la creacion del disco
	if path == "" {
		consola.AddToConsole("no se encontro una ruta\n")
		return false
	}
	err := os.Remove(path)
	if err != nil {
		consola.AddToConsole(fmt.Sprintf("no se pudo eliminar el archivo %s\n", err.Error()))
		return false
	}
	return true
}

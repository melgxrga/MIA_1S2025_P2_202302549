package usuariosygrupos

import (
	"fmt"
	"os"
	"io/ioutil"
	"strings"
	"regexp"
	"github.com/melgxrga/proyecto1Archivos/consola"
	"github.com/melgxrga/proyecto1Archivos/logger"
)

type Edit struct {
	Params struct {
		Path      string
		Contenido string
	}
}

func (e *Edit) Exe(params []string) {
	// Parsear los parámetros recibidos
	editCmd, err := ParseEditParams(strings.Join(params, " "))
	if err != nil {
		consola.AddToConsole(fmt.Sprintf("ERROR: %s\n", err))
		return
	}
	e.Params = editCmd.Params

	if !logger.Log.IsLoggedIn() {
		consola.AddToConsole("ERROR: Debe estar logueado para editar archivos.\n")
		return
	}
	// Verificar permisos de lectura y escritura sobre el archivo
	fileInfo, err := os.Stat(e.Params.Path)
	if err != nil {
		consola.AddToConsole(fmt.Sprintf("ERROR: El archivo no existe: %s\n", e.Params.Path))
		return
	}
	if fileInfo.IsDir() {
		consola.AddToConsole(fmt.Sprintf("ERROR: No se puede editar una carpeta: %s\n", e.Params.Path))
		return
	}
	// Leer el contenido a insertar
	contenidoBytes, err := ioutil.ReadFile(e.Params.Contenido)
	if err != nil {
		consola.AddToConsole(fmt.Sprintf("ERROR: No se pudo leer el archivo de contenido: %s\n", e.Params.Contenido))
		return
	}
	// Escribir el contenido en el archivo destino
	err = ioutil.WriteFile(e.Params.Path, contenidoBytes, 0644)
	if err != nil {
		consola.AddToConsole(fmt.Sprintf("ERROR: No se pudo editar el archivo: %s\n", err))
		return
	}
	consola.AddToConsole(fmt.Sprintf("Archivo editado correctamente: %s\n", e.Params.Path))
}

// Parseo de parámetros para el comando edit
func ParseEditParams(paramStr string) (Edit, error) {
	var edit Edit
	args := paramStr
	// Permite -path="valor con espacios" o -path=valor
	re := regexp.MustCompile(`-path="[^"]+"|-path=[^\s]+|-contenido="[^"]+"|-contenido=[^\s]+`)
	matches := re.FindAllString(args, -1)
	for _, match := range matches {
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key, value := strings.ToLower(kv[0]), kv[1]
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}
		if key == "-path" {
			edit.Params.Path = value
		} else if key == "-contenido" {
			edit.Params.Contenido = value
		}
	}
	if edit.Params.Path == "" || edit.Params.Contenido == "" {
		return edit, fmt.Errorf("Faltan parámetros obligatorios -path o -contenido")
	}
	return edit, nil
}

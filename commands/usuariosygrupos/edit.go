package usuariosygrupos

import (
	"fmt"
	"os"
	"io/ioutil"
	"strings"
	"MIA_1S2025_P2_202302549/consola"
	"MIA_1S2025_P2_202302549/logger"
)

type Edit struct {
	Params struct {
		Path      string
		Contenido string
	}
}

func (e *Edit) Exe() {
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
	params := strings.Fields(paramStr)
	for _, param := range params {
		if strings.HasPrefix(param, "-path=") {
			edit.Params.Path = strings.TrimPrefix(param, "-path=")
		} else if strings.HasPrefix(param, "-contenido=") {
			edit.Params.Contenido = strings.TrimPrefix(param, "-contenido=")
		}
	}
	if edit.Params.Path == "" || edit.Params.Contenido == "" {
		return edit, fmt.Errorf("Faltan parámetros obligatorios -path o -contenido")
	}
	return edit, nil
}

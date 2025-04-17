package usuariosygrupos

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"github.com/melgxrga/proyecto1Archivos/consola"
	"github.com/melgxrga/proyecto1Archivos/logger"
)

type Rename struct {
	Params struct {
		Path string
		Name string
	}
}

// Parseo de parámetros para el comando rename
func ParseRenameParams(paramStr string) (Rename, error) {
	var rename Rename
	args := paramStr
	re := regexp.MustCompile(`-path="[^"]+"|-path=[^\s]+|-name="[^"]+"|-name=[^\s]+`)
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
			rename.Params.Path = value
		} else if key == "-name" {
			rename.Params.Name = value
		}
	}
	if rename.Params.Path == "" || rename.Params.Name == "" {
		return rename, fmt.Errorf("Faltan parámetros obligatorios -path o -name")
	}
	return rename, nil
}

func (r *Rename) Exe(params []string) {
	// Parsear los parámetros recibidos
	renameCmd, err := ParseRenameParams(strings.Join(params, " "))
	if err != nil {
		consola.AddToConsole(fmt.Sprintf("ERROR: %s\n", err))
		return
	}
	r.Params = renameCmd.Params

	if !logger.Log.IsLoggedIn() {
		consola.AddToConsole("ERROR: Debe estar logueado para renombrar archivos o carpetas.\n")
		return
	}
	// Verificar existencia y permisos de escritura
	fileInfo, err := os.Stat(r.Params.Path)
	if err != nil {
		consola.AddToConsole(fmt.Sprintf("ERROR: El archivo o carpeta no existe: %s\n", r.Params.Path))
		return
	}
	if fileInfo.Mode().Perm()&(1<<(uint(7))) == 0 {
		consola.AddToConsole("ERROR: No tiene permisos de escritura sobre el archivo o carpeta.\n")
		return
	}
	// Verificar que no exista el nuevo nombre en el mismo directorio
	dir := filepath.Dir(r.Params.Path)
	nuevoPath := filepath.Join(dir, r.Params.Name)
	if _, err := os.Stat(nuevoPath); err == nil {
		consola.AddToConsole(fmt.Sprintf("ERROR: Ya existe un archivo o carpeta con el nombre %s en el directorio.\n", r.Params.Name))
		return
	}
	// Renombrar
	err = os.Rename(r.Params.Path, nuevoPath)
	if err != nil {
		consola.AddToConsole(fmt.Sprintf("ERROR: No se pudo renombrar: %s\n", err))
		return
	}
	consola.AddToConsole(fmt.Sprintf("Renombrado correctamente a: %s\n", nuevoPath))
}

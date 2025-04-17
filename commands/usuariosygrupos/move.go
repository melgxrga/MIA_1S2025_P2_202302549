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

type Move struct {
	Params struct {
		Path    string
		Destino string
	}
}

// Parseo de parámetros para el comando move
func ParseMoveParams(paramStr string) (Move, error) {
	var moveCmd Move
	args := paramStr
	re := regexp.MustCompile(`-path="[^"]+"|-path=[^\s]+|-destino="[^"]+"|-destino=[^\s]+`)
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
			moveCmd.Params.Path = value
		} else if key == "-destino" {
			moveCmd.Params.Destino = value
		}
	}
	if moveCmd.Params.Path == "" || moveCmd.Params.Destino == "" {
		return moveCmd, fmt.Errorf("Faltan parámetros obligatorios -path o -destino")
	}
	return moveCmd, nil
}

func (m *Move) Exe(params []string) {
	moveCmd, err := ParseMoveParams(strings.Join(params, " "))
	if err != nil {
		consola.AddToConsole(fmt.Sprintf("ERROR: %s\n", err))
		return
	}
	m.Params = moveCmd.Params

	if !logger.Log.IsLoggedIn() {
		consola.AddToConsole("ERROR: Debe estar logueado para mover archivos o carpetas.\n")
		return
	}

	info, err := os.Stat(m.Params.Path)
	if err != nil {
		consola.AddToConsole(fmt.Sprintf("ERROR: La ruta de origen no existe: %s\n", m.Params.Path))
		return
	}
	// Verificar permisos de escritura
	if info.Mode().Perm()&(1<<(uint(7))) == 0 {
		consola.AddToConsole("ERROR: No tiene permisos de escritura sobre el archivo o carpeta de origen.\n")
		return
	}

	destInfo, err := os.Stat(m.Params.Destino)
	if err != nil || !destInfo.IsDir() {
		consola.AddToConsole("ERROR: La carpeta destino no existe o no es un directorio.\n")
		return
	}
	if destInfo.Mode().Perm()&(1<<(uint(7))) == 0 {
		consola.AddToConsole("ERROR: No tiene permisos de escritura sobre la carpeta destino.\n")
		return
	}

	newPath := filepath.Join(m.Params.Destino, filepath.Base(m.Params.Path))
	fmt.Printf("[move] Moviendo %s a %s\n", m.Params.Path, newPath)
	err = os.Rename(m.Params.Path, newPath)
	if err != nil {
		consola.AddToConsole(fmt.Sprintf("ERROR: No se pudo mover: %s\n", err))
		fmt.Printf("[move] ERROR al mover: %s\n", err)
		return
	}
	consola.AddToConsole(fmt.Sprintf("Movimiento realizado correctamente a: %s\n", newPath))
}

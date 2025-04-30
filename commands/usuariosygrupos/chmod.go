package usuariosygrupos

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"os/user"
	"github.com/melgxrga/proyecto1Archivos/consola"
	"github.com/melgxrga/proyecto1Archivos/logger"
	datos "github.com/melgxrga/proyecto1Archivos/structures"
	comandos "github.com/melgxrga/proyecto1Archivos/commands"
)

type Chmod struct {
	Params struct {
		Path      string
		Ugo       string
		Recursivo bool
	}
}

func ParseChmodParams(paramStr string) (Chmod, error) {
	var chmodCmd Chmod
	args := paramStr
	re := regexp.MustCompile(`-path="[^"]+"|-path=[^\s]+|-ugo="[^"]+"|-ugo=[^\s]+|-r`)
	matches := re.FindAllString(args, -1)
	for _, match := range matches {
		if match == "-r" {
			chmodCmd.Params.Recursivo = true
			continue
		}
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key, value := strings.ToLower(kv[0]), kv[1]
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}
		if key == "-path" {
			chmodCmd.Params.Path = value
		} else if key == "-ugo" {
			chmodCmd.Params.Ugo = value
		}
	}
	if chmodCmd.Params.Path == "" || chmodCmd.Params.Ugo == "" {
		return chmodCmd, fmt.Errorf("Faltan parámetros obligatorios -path o -ugo")
	}
	return chmodCmd, nil
}

func (c *Chmod) Exe(params []string) {
	chmodCmd, err := ParseChmodParams(strings.Join(params, " "))
	if err != nil {
		consola.AddToConsole(fmt.Sprintf("ERROR: %s\n", err))
		return
	}
	c.Params = chmodCmd.Params

	if !logger.Log.IsLoggedIn() {
		consola.AddToConsole("ERROR: Debe haber una sesión activa para ejecutar chmod.\n")
		return
	}
	// Validar formato de ugo
	if len(c.Params.Ugo) != 3 {
		consola.AddToConsole("ERROR: El parámetro -ugo debe tener exactamente 3 dígitos.\n")
		return
	}
	ugo := make([]int, 3)
	for i, ch := range c.Params.Ugo {
		val, err := strconv.Atoi(string(ch))
		if err != nil || val < 0 || val > 7 {
			consola.AddToConsole("ERROR: Cada dígito de -ugo debe estar entre 0 y 7.\n")
			return
		}
		ugo[i] = val
	}

	// Obtener usuario actual
	userNameArr := logger.Log.GetUserName()
	userName := strings.Trim(string(userNameArr[:]), "\x00")

	// Validar permisos: solo root o dueño puede cambiar permisos
	if !logger.Log.UserIsRoot() && !usuarioEsDueno(c.Params.Path, userName) {
		consola.AddToConsole("ERROR: Solo el usuario root o el dueño pueden cambiar permisos.\n")
		return
	}

	mode, err := strconv.ParseUint(c.Params.Ugo, 8, 32)
	if err != nil {
		consola.AddToConsole("ERROR: El parámetro -ugo debe ser un número octal válido.\n")
		return
	}

	if c.Params.Recursivo {
		err := filepath.WalkDir(c.Params.Path, func(path string, d os.DirEntry, err error) error {
			if err == nil {
				errChmod := os.Chmod(path, os.FileMode(mode))
				if errChmod != nil {
					consola.AddToConsole(fmt.Sprintf("ERROR: No se pudo cambiar permisos en '%s': %v\n", path, errChmod))
				}
			}
			return nil
		})
		if err != nil {
			consola.AddToConsole(fmt.Sprintf("ERROR: Falló el recorrido recursivo: %v\n", err))
			return
		}
	} else {
		err := os.Chmod(c.Params.Path, os.FileMode(mode))
		if err != nil {
			consola.AddToConsole(fmt.Sprintf("ERROR: No se pudo cambiar permisos en '%s': %v\n", c.Params.Path, err))
			return
		}
	}
	consola.AddToConsole(fmt.Sprintf("Permisos cambiados a '%s' en '%s'\n", c.Params.Ugo, c.Params.Path))
}

// usuarioEsDueno verifica si el usuario activo es dueño del archivo/carpeta en el SO real
func usuarioEsDueno(path string, userName string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return false
	}
	uid := stat.Uid
	// Obtener UID real del usuario activo
	usr, err := user.Lookup(userName)
	if err != nil {
		return false
	}
	uidActivo, err := strconv.ParseUint(usr.Uid, 10, 32)
	if err != nil {
		return false
	}
	return uid == uint32(uidActivo)
}

func chmodArchivo(path string, ugo []int, diskPath string, superbloque *datos.SuperBloque, tablaInodo *datos.TablaInodo, usuarioId int64) bool {
	numInodo, inodo, _, ok := BuscarInodoPorRuta(path, tablaInodo, superbloque, diskPath)
	if !ok {
		consola.AddToConsole(fmt.Sprintf("ERROR: No se encontró el archivo/carpeta '%s'\n", path))
		return false
	}
	// Solo root puede cambiar permisos de cualquier archivo, otros solo si son dueños
	if !logger.Log.UserIsRoot() && inodo.I_uid != usuarioId {
		consola.AddToConsole("ERROR: Solo el usuario root o el dueño pueden cambiar permisos.\n")
		return false
	}
	// Cambiar permisos
	inodo.I_perm = int64(ugo[0])*64 + int64(ugo[1])*8 + int64(ugo[2])
	comandos.Fwrite(inodo, diskPath, superbloque.S_inode_start+numInodo*superbloque.S_inode_size)
	return true
}

func chmodRecursivo(dir string, ugo []int, diskPath string, superbloque *datos.SuperBloque, tablaInodo *datos.TablaInodo, usuarioId int64) {
	_, inodo, _, ok := BuscarInodoPorRuta(dir, tablaInodo, superbloque, diskPath)
	if !ok {
		consola.AddToConsole(fmt.Sprintf("ERROR: No se encontró el directorio '%s'\n", dir))
		return
	}
	if !chmodArchivo(dir, ugo, diskPath, superbloque, tablaInodo, usuarioId) {
		return
	}
	// Si es carpeta, recorre hijos
	for _, ptr := range inodo.I_block {
		if ptr == -1 {
			continue
		}
		var bloque datos.BloqueDeCarpetas
		comandos.Fread(&bloque, diskPath, superbloque.S_block_start+ptr*superbloque.S_block_size)
		for _, content := range bloque.B_content {
			name := strings.Trim(string(content.B_name[:]), "\x00")
			if name == "." || name == ".." || content.B_inodo == -1 {
				continue
			}
			var hijo datos.TablaInodo
			comandos.Fread(&hijo, diskPath, superbloque.S_inode_start+int64(content.B_inodo)*superbloque.S_inode_size)
			if hijo.I_type == 0 {
				chmodRecursivo(filepath.Join(dir, name), ugo, diskPath, superbloque, tablaInodo, usuarioId)
			} else {
				chmodArchivo(filepath.Join(dir, name), ugo, diskPath, superbloque, tablaInodo, usuarioId)
			}
		}
	}
}

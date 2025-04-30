package usuariosygrupos

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
	"github.com/melgxrga/proyecto1Archivos/consola"
	"github.com/melgxrga/proyecto1Archivos/logger"
	datos "github.com/melgxrga/proyecto1Archivos/structures"
	comandos "github.com/melgxrga/proyecto1Archivos/commands"
)

type Chown struct {
	Params struct {
		Path    string
		Usuario string
		Recursivo bool
	}
}

func ParseChownParams(paramStr string) (Chown, error) {
	var chownCmd Chown
	args := paramStr
	args = strings.ReplaceAll(args, "-r ", "-r=true ") // Soporta -r como flag
	params := strings.Fields(args)
	for _, param := range params {
		if strings.HasPrefix(param, "-path=") {
			chownCmd.Params.Path = strings.TrimPrefix(param, "-path=")
		} else if strings.HasPrefix(param, "-usuario=") {
			chownCmd.Params.Usuario = strings.TrimPrefix(param, "-usuario=")
		} else if strings.HasPrefix(param, "-r") {
			chownCmd.Params.Recursivo = true
		}
	}
	if chownCmd.Params.Path == "" || chownCmd.Params.Usuario == "" {
		return chownCmd, fmt.Errorf("Faltan parámetros obligatorios -path o -usuario")
	}
	return chownCmd, nil
}

// Exe ejecuta el comando chown
func (c *Chown) Exe(params []string) {
	chownCmd, err := ParseChownParams(strings.Join(params, " "))
	if err != nil {
		consola.AddToConsole(fmt.Sprintf("ERROR: %s\n", err))
		return
	}
	c.Params = chownCmd.Params

	if !logger.Log.IsLoggedIn() {
		consola.AddToConsole("ERROR: Debe haber una sesión activa para ejecutar chown.\n")
		return
	}
	// Buscar disco y superbloque
	diskPath, start, rutaInterna := GetMountInfoByPath(c.Params.Path)
	if diskPath == "" {
		consola.AddToConsole("ERROR: No se encontró el disco montado para la ruta dada.\n")
		return
	}
	var superbloque datos.SuperBloque
	comandos.Fread(&superbloque, diskPath, start)
	var rootInodo datos.TablaInodo
	comandos.Fread(&rootInodo, diskPath, superbloque.S_inode_start)
	// Validar usuario destino
	if !usuarioExiste(c.Params.Usuario, diskPath, &superbloque, &rootInodo) {
		consola.AddToConsole(fmt.Sprintf("ERROR: El usuario '%s' no existe\n", c.Params.Usuario))
		return
	}
	// Obtener usuario actual
	userNameArr := logger.Log.GetUserName()
	userName := strings.Trim(string(userNameArr[:]), "\x00")
	contenido := ReadFile(&rootInodo, diskPath, &superbloque)
	usuarioId := GetUserId(contenido, userName)
	if usuarioId == -1 {
		consola.AddToConsole("ERROR: Usuario actual no válido.\n")
		return
	}
	if c.Params.Recursivo {
		chownRecursivo(rutaInterna, c.Params.Usuario, diskPath, &superbloque, &rootInodo, usuarioId)
	} else {
		if !cambiarPropietario(rutaInterna, c.Params.Usuario, diskPath, &superbloque, &rootInodo, usuarioId) {
			return
		}
	}
	consola.AddToConsole(fmt.Sprintf("Propietario cambiado a '%s' en '%s'\n", c.Params.Usuario, c.Params.Path))
}

// --- Funciones auxiliares (mock, debes adaptar a tu sistema real) ---

func usuarioExiste(usuario string, path string, superbloque *datos.SuperBloque, tablaInodo *datos.TablaInodo) bool {
	contenido := ReadFile(tablaInodo, path, superbloque)
	mk := Mkusr{}
	return mk.ExisteUsuario(contenido, usuario)
}

func tienePermisoCambiarPropietario(inodo *datos.TablaInodo, usuarioId int64) bool {
	// Solo root (uid=1) o el dueño pueden cambiar propietario
	return usuarioId == 1 || inodo.I_uid == usuarioId
}

func cambiarPropietario(path, usuario string, diskPath string, superbloque *datos.SuperBloque, tablaInodo *datos.TablaInodo, usuarioId int64) bool {
	// Buscar inodo
	numInodo, inodo, _, ok := BuscarInodoPorRuta(path, tablaInodo, superbloque, diskPath)
	if !ok {
		consola.AddToConsole(fmt.Sprintf("ERROR: No se encontró el archivo/carpeta '%s'\n", path))
		return false
	}
	mk := Mkusr{}
	contenido := ReadFile(tablaInodo, diskPath, superbloque)
	if !mk.ExisteUsuario(contenido, usuario) {
		consola.AddToConsole(fmt.Sprintf("ERROR: El usuario '%s' no existe\n", usuario))
		return false
	}
	nuevoUid := GetUserId(contenido, usuario)
	if !tienePermisoCambiarPropietario(inodo, usuarioId) {
		consola.AddToConsole(fmt.Sprintf("ERROR: No tienes permisos para cambiar el propietario de '%s'\n", path))
		return false
	}
	inodo.I_uid = nuevoUid
	// Actualiza mtime
	t := time.Now().String()
	copy(inodo.I_mtime[:], []byte(t)[:len(inodo.I_mtime)])
	comandos.Fwrite(inodo, diskPath, superbloque.S_inode_start+numInodo*superbloque.S_inode_size)
	return true
}

func chownRecursivo(dir, usuario string, diskPath string, superbloque *datos.SuperBloque, tablaInodo *datos.TablaInodo, usuarioId int64) {
	_, inodo, _, ok := BuscarInodoPorRuta(dir, tablaInodo, superbloque, diskPath)
	if !ok {
		consola.AddToConsole(fmt.Sprintf("ERROR: No se encontró el directorio '%s'\n", dir))
		return
	}
	if !cambiarPropietario(dir, usuario, diskPath, superbloque, tablaInodo, usuarioId) {
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
				// Carpeta
				chownRecursivo(filepath.Join(dir, name), usuario, diskPath, superbloque, tablaInodo, usuarioId)
			} else {
				cambiarPropietario(filepath.Join(dir, name), usuario, diskPath, superbloque, tablaInodo, usuarioId)
			}
		}
	}
}

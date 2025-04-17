package usuariosygrupos

import (
	"fmt"
	"strings"
	"github.com/melgxrga/proyecto1Archivos/consola"
	"github.com/melgxrga/proyecto1Archivos/logger"
	datos "github.com/melgxrga/proyecto1Archivos/structures"
	"regexp"
	"os"
	"unsafe"
	comandos "github.com/melgxrga/proyecto1Archivos/commands"
	lista "github.com/melgxrga/proyecto1Archivos/list"
	"strconv"
)



type ParametrosRemove struct {
	Path string
}

type Remove struct {
	Params ParametrosRemove
}

func (r *Remove) SaveParams(parametros []string) ParametrosRemove {
	var params ParametrosRemove
	args := strings.Join(parametros, " ")
	re := regexp.MustCompile(`(?i)-path="[^"]+"|(?i)-path=[^\s]+`)
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
			params.Path = value
		}
	}
	return params
}

func (r *Remove) Exe(parametros []string) {
	r.Params = r.SaveParams(parametros)
	if r.Params.Path == "" {
		consola.AddToConsole("ERROR: El parámetro -path es obligatorio.\n")
		return
	}
	if !logger.Log.IsLoggedIn() {
		consola.AddToConsole("ERROR: Debe estar logueado para eliminar archivos o carpetas.\n")
		return
	}
	// Aquí asume que existe una función RemoveNodeVirtual que elimina en el FS virtual
	err := RemoveNodeReal(r.Params.Path)
	if err == nil {
		consola.AddToConsole(fmt.Sprintf("\nSe eliminó correctamente del sistema real: %s\n\n", r.Params.Path))
	} else {
		consola.AddToConsole(fmt.Sprintf("\nERROR: No se encontró el archivo en el sistema real: %s\n\n", err))
	}
}

// GetMountedDiskAndStart busca el nodo montado por ID y retorna la ruta del disco y el inicio de la partición
func GetMountedDiskAndStart(id string) (string, int64) {
	nodo := lista.ListaMount.GetNodeById(id)
	if nodo == nil {
		return "", 0
	}
	if nodo.Value != nil {
		return nodo.Ruta, nodo.Value.Part_start
	} else if nodo.ValueL != nil {
		return nodo.Ruta, nodo.ValueL.Part_start + int64(unsafe.Sizeof(datos.EBR{}))
	}
	return "", 0
}

// GetMountInfoByPath busca el disco montado correspondiente al path y obtiene la ruta interna al FS
// Busca el nodo de montaje cuyo "Ruta" es prefijo del path absoluto del archivo
func GetMountInfoByPath(path string) (string, int64, string) {
	var best *lista.MountNode
	maxLen := -1
	temp := lista.ListaMount.First
	for temp != nil {
		if strings.HasPrefix(path, temp.Ruta) && len(temp.Ruta) > maxLen {
			best = temp
			maxLen = len(temp.Ruta)
		}
		temp = temp.Next
	}
	if best == nil {
		return "", 0, ""
	}
	// La ruta interna es el path relativo al punto de montaje
	rutaInterna := strings.TrimPrefix(path, best.Ruta)
	if strings.HasPrefix(rutaInterna, "/") {
		rutaInterna = rutaInterna[1:]
	}
	var start int64
	if best.Value != nil {
		start = best.Value.Part_start
	} else if best.ValueL != nil {
		start = best.ValueL.Part_start + int64(unsafe.Sizeof(datos.EBR{}))
	}
	return best.Ruta, start, rutaInterna
}

// RemoveNodeVirtual elimina el archivo/carpeta en el FS virtual, verificando permisos recursivamente
func RemoveNodeVirtual(path string, username string) (bool, string) {
	// Buscar en todos los discos montados
	for nodo := lista.ListaMount.First; nodo != nil; nodo = nodo.Next {
		var start int64
		if nodo.Value != nil {
			start = nodo.Value.Part_start
		} else if nodo.ValueL != nil {
			start = nodo.ValueL.Part_start + int64(unsafe.Sizeof(datos.EBR{}))
		}
		var superbloque datos.SuperBloque
		var rootInodo datos.TablaInodo
		comandos.Fread(&superbloque, nodo.Ruta, start)
		comandos.Fread(&rootInodo, nodo.Ruta, superbloque.S_inode_start)
		// Buscar el inodo del archivo/carpeta en este disco
		numInodo, inodo, esCarpeta, ok := BuscarInodoPorRuta(path, &rootInodo, &superbloque, nodo.Ruta)
		if ok {
			// 3. Verificar permisos de escritura
			if !TienePermisoEscritura(inodo, username) {
				return false, "No tiene permisos de escritura sobre el archivo/carpeta."
			}
			// 4. Si es archivo, eliminar
			if !esCarpeta {
				EliminarInodoYBloques(numInodo, nodo.Ruta, &superbloque)
				return true, "Archivo eliminado correctamente."
			}
			// 5. Si es carpeta, verificar recursivamente todos los hijos
			if !PuedeEliminarCarpetaRecursivo(inodo, username, nodo.Ruta, &superbloque) {
				return false, "No se puede eliminar la carpeta porque algún archivo o subcarpeta no tiene permisos de escritura."
			}
			EliminarCarpetaRecursivo(numInodo, nodo.Ruta, &superbloque)
			return true, "Carpeta eliminada correctamente."
		}
	}
	return false, "No existe el archivo o carpeta a eliminar en ningún disco montado."
}

// --- STUBS Y UTILIDADES ---
// Busca el inodo por ruta absoluta. Retorna numInodo, puntero a inodo, esCarpeta, ok
func BuscarInodoPorRuta(path string, rootInodo *datos.TablaInodo, superbloque *datos.SuperBloque, diskPath string) (int64, *datos.TablaInodo, bool, bool) {
	// Quitar '/' inicial y dividir la ruta
	path = strings.TrimPrefix(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || (len(parts) == 1 && parts[0] == "") {
		return 0, rootInodo, true, true // raíz
	}
	return buscarInodoRec(parts, rootInodo, superbloque, diskPath, 0)
}

// Búsqueda recursiva de inodos
func buscarInodoRec(parts []string, inodo *datos.TablaInodo, superbloque *datos.SuperBloque, diskPath string, nivel int) (int64, *datos.TablaInodo, bool, bool) {
	if nivel >= len(parts) {
		return int64(nivel), inodo, inodo.I_type == 0, true
	}
	nombre := parts[nivel]
	// Buscar en los bloques de carpetas
	for _, ptr := range inodo.I_block {
		if ptr == -1 {
			continue
		}
		var bloque datos.BloqueDeCarpetas
		comandos.Fread(&bloque, diskPath, superbloque.S_block_start+ptr*int64(unsafe.Sizeof(datos.BloqueDeCarpetas{})))
		for _, content := range bloque.B_content {
			cname := strings.Trim(string(content.B_name[:]), "\x00")
			if cname == nombre {
				// Leer el inodo hijo
				var nextInodo datos.TablaInodo
				comandos.Fread(&nextInodo, diskPath, superbloque.S_inode_start+int64(content.B_inodo)*int64(unsafe.Sizeof(datos.TablaInodo{})))
				return buscarInodoRec(parts, &nextInodo, superbloque, diskPath, nivel+1)
			}
		}
	}
	return -1, nil, false, false
}

func TienePermisoEscritura(inodo *datos.TablaInodo, username string) bool {
	// Simulación de modelo UGO: usuario, grupo, otros
	userUIDstr := logger.Log.GetUserId() // string
	userUID, _ := strconv.ParseInt(userUIDstr, 10, 64)
	// userGIDstr := logger.Log.GetGroupId() // string (descomenta si tienes esta función)
	// userGID, _ := strconv.ParseInt(userGIDstr, 10, 64)
	perm := inodo.I_perm

	if username == "root" || userUID == 1 {
		return (perm>>1)&1 == 1 // bit de escritura para propietario
	}
	if inodo.I_uid == userUID {
		return (perm>>1)&1 == 1 // bit de escritura para propietario
	}

	return (perm>>7)&1 == 1 // bit de escritura para otros
}
// Elimina inodo y sus bloques
// EliminarInodoYBloques libera el inodo y todos sus bloques asociados (archivos)
func EliminarInodoYBloques(numInodo int64, diskPath string, superbloque *datos.SuperBloque) {
	// Leer el inodo
	var inodo datos.TablaInodo
	comandos.Fread(&inodo, diskPath, superbloque.S_inode_start+numInodo*int64(unsafe.Sizeof(datos.TablaInodo{})))
	// Liberar bloques
	for _, ptr := range inodo.I_block {
		if ptr == -1 {
			continue
		}
		// Limpiar bloque (opcional, para depuración)
		var bloque datos.BloqueDeArchivos
		copy(bloque.B_content[:], []byte{})
		comandos.Fwrite(&bloque, diskPath, superbloque.S_block_start+ptr*int64(unsafe.Sizeof(datos.BloqueDeArchivos{})))
		// Liberar en bitmap (no implementado aquí)
	}
	// Limpiar inodo
	var empty datos.TablaInodo
	comandos.Fwrite(&empty, diskPath, superbloque.S_inode_start+numInodo*int64(unsafe.Sizeof(datos.TablaInodo{})))
	// Liberar en bitmap (no implementado aquí)
}
// PuedeEliminarCarpetaRecursivo verifica recursivamente si se puede eliminar toda la carpeta (todos los hijos tienen permisos de escritura)
func PuedeEliminarCarpetaRecursivo(inodo *datos.TablaInodo, username string, diskPath string, superbloque *datos.SuperBloque) bool {
	if inodo.I_type == 1 {
		return TienePermisoEscritura(inodo, username)
	}
	for _, ptr := range inodo.I_block {
		if ptr == -1 {
			continue
		}
		var bloque datos.BloqueDeCarpetas
		comandos.Fread(&bloque, diskPath, superbloque.S_block_start+ptr*int64(unsafe.Sizeof(datos.BloqueDeCarpetas{})))
		for _, content := range bloque.B_content {
			cname := strings.Trim(string(content.B_name[:]), "\x00")
			if cname == "." || cname == ".." || cname == "" {
				continue
			}
			var hijo datos.TablaInodo
			comandos.Fread(&hijo, diskPath, superbloque.S_inode_start+int64(content.B_inodo)*int64(unsafe.Sizeof(datos.TablaInodo{})))
			if !PuedeEliminarCarpetaRecursivo(&hijo, username, diskPath, superbloque) {
				return false
			}
		}
	}
	return TienePermisoEscritura(inodo, username)
}
// Elimina recursivamente la carpeta y su contenido
// EliminarCarpetaRecursivo elimina recursivamente una carpeta y todo su contenido
func EliminarCarpetaRecursivo(numInodo int64, diskPath string, superbloque *datos.SuperBloque) {
	var inodo datos.TablaInodo
	comandos.Fread(&inodo, diskPath, superbloque.S_inode_start+numInodo*int64(unsafe.Sizeof(datos.TablaInodo{})))
	if inodo.I_type == 1 {
		EliminarInodoYBloques(numInodo, diskPath, superbloque)
		return
	}
	// Recorrer bloques de carpetas
	for _, ptr := range inodo.I_block {
		if ptr == -1 {
			continue
		}
		var bloque datos.BloqueDeCarpetas
		comandos.Fread(&bloque, diskPath, superbloque.S_block_start+ptr*int64(unsafe.Sizeof(datos.BloqueDeCarpetas{})))
		for _, content := range bloque.B_content {
			cname := strings.Trim(string(content.B_name[:]), "\x00")
			if cname == "." || cname == ".." || cname == "" {
				continue
			}
			EliminarCarpetaRecursivo(int64(content.B_inodo), diskPath, superbloque)
		}
	}
	// Finalmente elimina el inodo de la carpeta
	EliminarInodoYBloques(numInodo, diskPath, superbloque)
}

// RemoveNodeReal elimina del sistema real (SO) si existe
func RemoveNodeReal(path string) error {
	return os.RemoveAll(path)
}

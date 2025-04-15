package usuariosygrupos

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
	"unsafe"
	"regexp"
	"github.com/melgxrga/proyecto1Archivos/bitmap"
	"github.com/melgxrga/proyecto1Archivos/commands"
	"github.com/melgxrga/proyecto1Archivos/consola"
	"github.com/melgxrga/proyecto1Archivos/structures"
	"github.com/melgxrga/proyecto1Archivos/list"
	"github.com/melgxrga/proyecto1Archivos/logger"
)

type ParametrosMkdir struct {
	Path string
	R    bool
}

type Mkdir struct {
	Params ParametrosMkdir
}

func (m *Mkdir) Exe(parametros []string) {
	m.Params = m.SaveParams(parametros)

	// Intentamos crear el directorio y verificamos si realmente fue creado
	if creado := m.Mkdir(m.Params.Path, m.Params.R); creado {
		if _, err := os.Stat(m.Params.Path); os.IsNotExist(err) {
			// Si Mkdir devolvió true pero el directorio no existe, corregimos el mensaje
			consola.AddToConsole(fmt.Sprintf("\nERROR: la carpeta con ruta %s no se creó correctamente\n\n", m.Params.Path))
		} else {
			// Si el directorio realmente existe, se confirma su creación
			consola.AddToConsole(fmt.Sprintf("\nla carpeta con ruta %s se creó correctamente\n\n", m.Params.Path))
		}
	} else {
		consola.AddToConsole(fmt.Sprintf("\nla carpeta con ruta %s no se pudo crear\n\n", m.Params.Path))
	}
}


func (m *Mkdir) SaveParams(parametros []string) ParametrosMkdir {
	var params ParametrosMkdir
	// Unir todos los parámetros en una sola cadena
	args := strings.Join(parametros, " ")

	// Expresión regular para capturar los parámetros
	re := regexp.MustCompile(`-path="[^"]+"|-path=[^\s]+|-r`)
	matches := re.FindAllString(args, -1)

	// Iterar sobre cada coincidencia
	for _, match := range matches {
		kv := strings.SplitN(match, "=", 2)

		// Si es la opción `-r`, solo se marca como `true`
		if match == "-r" {
			params.R = true
			continue
		}

		if len(kv) != 2 {
			fmt.Printf("Formato de parámetro inválido: %s\n", match)
			continue
		}
		key, value := strings.ToLower(kv[0]), kv[1]

		// Quitar comillas si las tiene
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		// Procesar el parámetro
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

	// Validación final del parámetro obligatorio
	if params.Path == "" {
		fmt.Println("Error: Falta el parámetro obligatorio -path")
	}

	return params
}


func (m *Mkdir) Mkdir(path string, r bool) bool {
	if path == "" {
		consola.AddToConsole("no se encontro una ruta\n")
		return false
	}
	path = strings.Replace(path, "/", "", 1)
	if !logger.Log.IsLoggedIn() {
		consola.AddToConsole("no se encuentra un usuario loggeado para crear un archivo\n")
		return false
	}

	if lista.ListaMount.GetNodeById(logger.Log.GetUserId()).Value != nil {
		return createDirectory(logger.Log.GetUserName(), lista.ListaMount.GetNodeById(logger.Log.GetUserId()).Ruta, path, lista.ListaMount.GetNodeById(logger.Log.GetUserId()).Value.Part_start, r)
	} else if lista.ListaMount.GetNodeById(logger.Log.GetUserId()).ValueL != nil {
		return createDirectory(logger.Log.GetUserName(), lista.ListaMount.GetNodeById(logger.Log.GetUserId()).Ruta, path, lista.ListaMount.GetNodeById(logger.Log.GetUserId()).ValueL.Part_start+int64(unsafe.Sizeof(datos.EBR{})), r)
	}
	return false
}

func createDirectory(name [10]byte, path, ruta string, whereToStart int64, r bool) bool {
	fmt.Println("Iniciando createDirectory con name:", string(TrimArray(name[:])), "path:", path, "ruta:", ruta, "whereToStart:", whereToStart, "r:", r)

	// superbloque de la partición
	var superbloque datos.SuperBloque
	comandos.Fread(&superbloque, path, whereToStart)
	fmt.Println("SuperBloque leído: S_inode_start =", superbloque.S_inode_start, "S_inode_size =", superbloque.S_inode_size)

	// leyendo la tabla de inodos raíz
	var tablaInodoRoot datos.TablaInodo
	comandos.Fread(&tablaInodoRoot, path, superbloque.S_inode_start)
	fmt.Println("Tabla de inodos raíz leída desde:", superbloque.S_inode_start)

	// obteniendo el contenido de Users.txt
	var tablaInodoUsers datos.TablaInodo
	comandos.Fread(&tablaInodoUsers, path, superbloque.S_inode_start+superbloque.S_inode_size)
	fmt.Println("Tabla de inodos de Users.txt leída desde:", superbloque.S_inode_start+superbloque.S_inode_size)

	contenido := ReadFile(&tablaInodoUsers, path, &superbloque)
	fmt.Println("Contenido de Users.txt:", contenido)

	// obteniendo el user ID y el group ID
	userId := GetUserId(contenido, string(TrimArray(name[:])))
	groupId := GetGroupId(contenido, string(TrimArray(name[:])))
	fmt.Println("UserID obtenido:", userId, "GroupID obtenido:", groupId)

	if r {
		fmt.Println("Modo recursivo activado, creando directorios intermedios si es necesario.")
		FindAndCreateDirectories(&tablaInodoRoot, path, ruta, &superbloque, 0, userId, groupId)
	}

	num := NewInodeDirectory(&superbloque, path, userId, groupId)
	fmt.Println("Nuevo inodo directorio creado en posición:", num)

	FindDirs(num, &tablaInodoRoot, path, ruta, &superbloque, 0)

	fmt.Println("Escribiendo tabla de inodos actualizada...")
	comandos.Fwrite(&tablaInodoRoot, path, superbloque.S_inode_start)
	fmt.Println("Escribiendo SuperBloque actualizado...")
	comandos.Fwrite(&superbloque, path, whereToStart)

	fmt.Println("Finalización exitosa de createDirectory")
	return true
}

func NewInodeDirectory(superbloque *datos.SuperBloque, path string, userId, groupId int64) int64 {
	fmt.Println("Iniciando NewInodeDirectory con userID:", userId, "groupID:", groupId)

	var nuevaTabla datos.TablaInodo
	posicionActual := bitmap.WriteInBitmapInode(path, superbloque)
	fmt.Println("Nueva posición de inodo asignada:", posicionActual)

	// Llenado de datos del inodo
	nuevaTabla.I_uid = userId
	nuevaTabla.I_gid = groupId
	nuevaTabla.I_size = 0
	nuevaTabla.I_type = '0'
	nuevaTabla.I_perm = 664

	// Llenando fechas
	atime := time.Now().String()
	ctime := time.Now().String()
	mtime := time.Now().String()

	fmt.Println("Tiempos generados: atime:", atime, "ctime:", ctime, "mtime:", mtime)

	for i := 0; i < len(nuevaTabla.I_atime)-1; i++ {
		nuevaTabla.I_atime[i] = atime[i]
		nuevaTabla.I_ctime[i] = ctime[i]
		nuevaTabla.I_mtime[i] = mtime[i]
	}

	// Inicializando bloques
	for i := 0; i < len(nuevaTabla.I_block); i++ {
		nuevaTabla.I_block[i] = -1
	}
	fmt.Println("Bloques del inodo inicializados a -1")

	// Crear nuevo bloque de carpetas
	posicionNuevoBloqueCarpetas := bitmap.WriteInBitmapBlock(path, superbloque)
	fmt.Println("Nueva posición de bloque de carpetas asignada:", posicionNuevoBloqueCarpetas)

	nuevaTabla.I_block[0] = posicionNuevoBloqueCarpetas

	nuevoBloqueCarpetas := datos.BloqueDeCarpetas{}

	// Configuración de los directorios especiales
	copy(nuevoBloqueCarpetas.B_content[0].B_name[:], ".")
	nuevoBloqueCarpetas.B_content[0].B_inodo = int32(posicionActual)

	copy(nuevoBloqueCarpetas.B_content[1].B_name[:], "..")
	nuevoBloqueCarpetas.B_content[1].B_inodo = -1

	copy(nuevoBloqueCarpetas.B_content[2].B_name[:], "")
	nuevoBloqueCarpetas.B_content[2].B_inodo = -1

	copy(nuevoBloqueCarpetas.B_content[3].B_name[:], "")
	nuevoBloqueCarpetas.B_content[3].B_inodo = -1

	fmt.Println("Bloque de carpetas configurado correctamente.")

	// Escribiendo la nueva tabla de inodos
	inodoWritePos := superbloque.S_inode_start + posicionActual*superbloque.S_inode_size
	fmt.Println("Escribiendo nuevo inodo en posición:", inodoWritePos)
	comandos.Fwrite(&nuevaTabla, path, inodoWritePos)

	// Escribiendo el nuevo bloque de carpetas
	bloqueWritePos := superbloque.S_block_start + posicionNuevoBloqueCarpetas*superbloque.S_block_size
	fmt.Println("Escribiendo nuevo bloque de carpetas en posición:", bloqueWritePos)
	comandos.Fwrite(&nuevoBloqueCarpetas, path, bloqueWritePos)

	fmt.Println("Finalización exitosa de NewInodeDirectory")
	return posicionActual
}


func GetContent(cont string) string {
	// aqui hay que leer el archivo y ejecutarlo
	file, err := os.Open(cont)
	if err != nil {
		consola.AddToConsole(fmt.Sprintf("Error al intentar abrir el archivo: %s\n", cont))
		return ""
	}

	defer file.Close()

	// Crear un scanner para luego leer linea por linea el archivo de entrada
	scanner := bufio.NewScanner(file)
	content := ""
	// Leyendo linea por linea
	for scanner.Scan() {
		// obteniendo la linea actual
		content += scanner.Text()
	}

	// comprobar que no hubo error al leer el archivo
	if err := scanner.Err(); err != nil {
		consola.AddToConsole(fmt.Sprintf("Error al leer el archivo: %s\n", err))
		return ""
	}
	return content
}

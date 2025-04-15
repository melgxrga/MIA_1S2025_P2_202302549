package usuariosygrupos

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unsafe"
	"regexp"
	"github.com/melgxrga/proyecto1Archivos/bitmap"
	"github.com/melgxrga/proyecto1Archivos/commands"
	"github.com/melgxrga/proyecto1Archivos/consola"
	"github.com/melgxrga/proyecto1Archivos/structures"
	"github.com/melgxrga/proyecto1Archivos/list"
	"github.com/melgxrga/proyecto1Archivos/logger"
)

type ParametrosMkfile struct {
	Path string
	R    bool
	Size int
	Cont string
}

type Mkfile struct {
	Params ParametrosMkfile
}


func (m *Mkfile) SaveParams(parametros []string) ParametrosMkfile {
	var params ParametrosMkfile
	// Unir todos los parámetros en una sola cadena
	args := strings.Join(parametros, " ")

	// Expresión regular para capturar los parámetros
	re := regexp.MustCompile(`-path="[^"]+"|-path=[^\s]+|-size=\d+|-cont="[^"]+"|-cont=[^\s]+|-r`)
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
		case "-size":
			size, err := strconv.Atoi(value)
			if err != nil || size < 0 {
				fmt.Println("Error: el tamaño debe ser un número entero positivo")
				continue
			}
			params.Size = size
		case "-cont":
			if value == "" {
				fmt.Println("Error: el contenido no puede estar vacío")
				continue
			}
			params.Cont = value
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


func (m *Mkfile) Exe(parametros []string) {
	fmt.Println("Depuración: Ejecutando Mkfile con parámetros:", parametros)
	m.Params = m.SaveParams(parametros)

	fmt.Println("Depuración: Parámetros guardados ->", m.Params)

	if creado := m.Mkfile(m.Params.Path, m.Params.R, m.Params.Size, m.Params.Cont); creado {
		if _, err := os.Stat(m.Params.Path); os.IsNotExist(err) {
			consola.AddToConsole(fmt.Sprintf("\nERROR: El archivo %s no se creó correctamente\n\n", m.Params.Path))
		} else {
			consola.AddToConsole(fmt.Sprintf("\nEl archivo %s se creó correctamente\n\n", m.Params.Path))
		}
	} else {
		consola.AddToConsole(fmt.Sprintf("\nNo se pudo crear el archivo %s\n\n", m.Params.Path))
	}
}

func (m *Mkfile) Mkfile(path string, r bool, size int, cont string) bool {
	fmt.Println("Depuración: Iniciando Mkfile con path:", path, "r:", r, "size:", size, "cont:", cont)

	if path == "" {
		consola.AddToConsole("ERROR: No se especificó una ruta.\n")
		return false
	}

	path = strings.Replace(path, "/", "", 1)
	fmt.Println("Depuración: Ruta procesada ->", path)

	if !logger.Log.IsLoggedIn() {
		consola.AddToConsole("ERROR: No hay un usuario logueado para crear archivos.\n")
		return false
	}

	fmt.Println("Depuración: Usuario logueado ->", logger.Log.GetUserName())

	if size < 0 {
		consola.AddToConsole("ERROR: El tamaño del archivo no puede ser negativo.\n")
		return false
	}

	fmt.Println("Depuración: Buscando partición montada...")

	montaje := lista.ListaMount.GetNodeById(logger.Log.GetUserId())

	if montaje == nil {
		fmt.Println("Depuración: No se encontró una partición montada.")
		return false
	}

	if montaje.Value != nil {
		fmt.Println("Depuración: Partición primaria encontrada en", montaje.Ruta)
		return createFile(logger.Log.GetUserName(), montaje.Ruta, path, montaje.Value.Part_start, r, size, cont)
	} else if montaje.ValueL != nil {
		fmt.Println("Depuración: Partición extendida encontrada en", montaje.Ruta)
		return createFile(logger.Log.GetUserName(), montaje.Ruta, path, montaje.ValueL.Part_start+int64(unsafe.Sizeof(datos.EBR{})), r, size, cont)
	}

	fmt.Println("Depuración: No se pudo determinar la partición.")
	return false
}
func createFile(name [10]byte, path, ruta string, whereToStart int64, r bool, size int, cont string) bool {
	fmt.Println("Depuración: Creando archivo en", ruta, "Inicio en:", whereToStart)

	var superbloque datos.SuperBloque
	comandos.Fread(&superbloque, path, whereToStart)
	fmt.Println("Depuración: Superbloque leído correctamente.")

	var tablaInodoRoot datos.TablaInodo
	comandos.Fread(&tablaInodoRoot, path, superbloque.S_inode_start)
	fmt.Println("Depuración: Inodo raíz leído correctamente.")

	var tablaInodoUsers datos.TablaInodo
	comandos.Fread(&tablaInodoUsers, path, superbloque.S_inode_start+superbloque.S_inode_size)
	contenido := ReadFile(&tablaInodoUsers, path, &superbloque)
	fmt.Println("Depuración: Contenido de Users.txt obtenido.")

	userId := GetUserId(contenido, string(TrimArray(name[:])))
	groupId := GetGroupId(contenido, string(TrimArray(name[:])))
	fmt.Println("Depuración: Usuario ID ->", userId, "Grupo ID ->", groupId)

	if r {
		fmt.Println("Depuración: Creando directorios si no existen...")
		FindAndCreateDirectories(&tablaInodoRoot, path, ruta, &superbloque, 0, userId, groupId)
	}

	content := ""
	if cont != "" {
		fmt.Println("Depuración: Cargando contenido desde", cont)
		content = getContent(cont)
		fmt.Println("Depuración: Contenido obtenido ->", content)
	}

	fmt.Println("Depuración: Preparando contenido con tamaño", size)

	if size > 0 {
		for i := len(content); i < size; i++ {
			content += strconv.Itoa(i % 10)
		}
	}

	fmt.Println("Depuración: Contenido final:", content)

	num := NewInodeFile(&superbloque, path, userId, groupId, content)
	fmt.Println("Depuración: Inodo asignado en posición", num)

	FindDirectories(num, &tablaInodoRoot, path, ruta, &superbloque, 0)
	comandos.Fwrite(&tablaInodoRoot, path, superbloque.S_inode_start)
	comandos.Fwrite(&superbloque, path, whereToStart)
	fmt.Println("Depuración: Superbloque actualizado correctamente.")

	PrintTree(&tablaInodoRoot, &superbloque, path)
	return true
}

func GetGroupId(contenido, name string) int64 {
	lineas := strings.Split(contenido, "\n")
	lineas = lineas[:len(lineas)-1]
	groupName := ""
	for _, linea := range lineas {
		linea = strings.ReplaceAll(linea, "\x00", "")
		parametros := strings.Split(linea, ",")
		if parametros[1] != "U" {
			continue
		}
		if parametros[3] == name {
			groupName = parametros[2]
		}
	}
	if groupName == "" {
		return -1
	}
	for _, linea := range lineas {
		linea = strings.ReplaceAll(linea, "\x00", "")
		parametros := strings.Split(linea, ",")
		if parametros[1] != "G" {
			continue
		}
		if parametros[2] == groupName {
			num, _ := strconv.Atoi(parametros[0])
			return int64(num)
		}
	}
	return -1
}

func GetUserId(contenido, name string) int64 {
	lineas := strings.Split(contenido, "\n")
	lineas = lineas[:len(lineas)-1]
	for _, linea := range lineas {
		linea = strings.ReplaceAll(linea, "\x00", "")
		parametros := strings.Split(linea, ",")
		if parametros[1] != "U" {
			continue
		}
		if parametros[3] == name {
			num, _ := strconv.Atoi(parametros[0])
			return int64(num)
		}
	}
	return -1
}

func NewInodeFile(superbloque *datos.SuperBloque, path string, userId, groupId int64, contenido string) int64 {
    fmt.Println("Depuración: Creando nuevo inodo.")

    var nuevaTabla datos.TablaInodo
    nuevaPosicion := bitmap.WriteInBitmapInode(path, superbloque)

    if nuevaPosicion == -1 {
        fmt.Println("❌ ERROR: No se pudo asignar un nuevo inodo.")
        return -1
    }

    // Mostrar la posición asignada del inodo
    fmt.Println("Depuración: Inodo asignado en posición", nuevaPosicion)

    // Asignar valores básicos al inodo
    nuevaTabla.I_uid = userId
    nuevaTabla.I_gid = groupId
    nuevaTabla.I_size = int64(len(contenido))
    nuevaTabla.I_type = '1'  // tipo archivo
    nuevaTabla.I_perm = 664  // permisos: lectura y escritura para usuario y grupo, solo lectura para otros

    // Inicializar bloques del inodo a -1
    for i := 0; i < len(nuevaTabla.I_block); i++ {
        nuevaTabla.I_block[i] = -1
    }

    // Imprimir el estado del inodo antes de asignar contenido
    fmt.Println("Depuración: Inodo antes de asignar contenido:")
    fmt.Printf("I_uid: %d\n", nuevaTabla.I_uid)
    fmt.Printf("I_gid: %d\n", nuevaTabla.I_gid)
    fmt.Printf("I_size: %d\n", nuevaTabla.I_size)
    fmt.Printf("I_type: %d\n", nuevaTabla.I_type)
    fmt.Printf("I_perm: %d\n", nuevaTabla.I_perm)
    fmt.Printf("I_block: %+v\n", nuevaTabla.I_block)

    // Asignar contenido a los bloques
    fmt.Println("Depuración: Asignando contenido a los bloques.")
    llenarTablaDeInodoDeArchivos(&nuevaTabla, superbloque, path, contenido)

    // Verificar el estado de los bloques después de la asignación
    fmt.Println("Depuración: Estado de los bloques después de llenar con contenido:")
    fmt.Printf("I_block: %+v\n", nuevaTabla.I_block)

    // Guardar el inodo en disco
    fmt.Println("Depuración: Guardando inodo en posición", superbloque.S_inode_start+nuevaPosicion*superbloque.S_inode_size)
    comandos.Fwrite(&nuevaTabla, path, superbloque.S_inode_start+nuevaPosicion*superbloque.S_inode_size)

    return nuevaPosicion
}



func getContent(cont string) string {
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
		content += scanner.Text() + "\n"
	}

	// comprobar que no hubo error al leer el archivo
	if err := scanner.Err(); err != nil {
		consola.AddToConsole(fmt.Sprintf("Error al leer el archivo: %s\n", err))
		return ""
	}
	return content
}

func llenarTablaDeInodoDeArchivos(tablaInodo *datos.TablaInodo, superbloque *datos.SuperBloque, path, contenido string) {
	for i := 0; i < len(tablaInodo.I_block); i++ {
		if tablaInodo.I_block[i] == -1 {
			var bloqueArchivo datos.BloqueDeArchivos
			posicionBloqueDeArchivo := bitmap.WriteInBitmapBlock(path, superbloque)

			if posicionBloqueDeArchivo == -1 {
				fmt.Println("Error: No se pudo asignar un nuevo bloque de datos.")
				return
			}

			fmt.Printf(" Depuración: Bloque de archivo asignado en posición %d\n", posicionBloqueDeArchivo)

			tablaInodo.I_block[i] = posicionBloqueDeArchivo
			if StrlenBytes([]byte(contenido)) > 63 {
				copy(bloqueArchivo.B_content[:], []byte(contenido[:63]))
				comandos.Fwrite(&bloqueArchivo, path, superbloque.S_block_start+posicionBloqueDeArchivo*superbloque.S_block_size)
				llenarTablaDeInodoDeArchivos(tablaInodo, superbloque, path, contenido[63:])
			} else {
				copy(bloqueArchivo.B_content[:], []byte(contenido[:]))
				comandos.Fwrite(&bloqueArchivo, path, superbloque.S_block_start+posicionBloqueDeArchivo*superbloque.S_block_size)
			}
			return
		}
	}
}



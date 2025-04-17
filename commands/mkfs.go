package comandos

import (
	"fmt"
	"math"
	"strings"
	"time"
	"unsafe"
	"regexp"
	"github.com/melgxrga/proyecto1Archivos/consola"
	lista "github.com/melgxrga/proyecto1Archivos/list"
	datos "github.com/melgxrga/proyecto1Archivos/structures"
)

type ParametrosMkfs struct {
	Id string
	T  string
	Fs string // 2fs o 3fs
}

type Mkfs struct {
	Params ParametrosMkfs
}

func (m *Mkfs) Exe(parametros []string) {
	m.Params = m.SaveParams(parametros)
	if m.Mkfs(m.Params.Id, m.Params.T) {
		fsMsg := "EXT2"
		if m.Params.Fs == "3fs" {
			fsMsg = "EXT3"
		}
		consola.AddToConsole(fmt.Sprintf("\nel formateo con %s de la particion con id %s fue exitoso\n\n", fsMsg, m.Params.Id))
	} else {
		consola.AddToConsole(fmt.Sprintf("no se logro formatear la particion con id %s\n", m.Params.Id))
	}
}

func (m *Mkfs) SaveParams(parametros []string) ParametrosMkfs {
	var params ParametrosMkfs
	args := strings.Join(parametros, " ")
	re := regexp.MustCompile(`-id=[^\s]+|-type=[fF]{4}|-fs=[23]fs`)  // Soporta -fs=2fs o -fs=3fs
	matches := re.FindAllString(args, -1)
	for _, match := range matches {
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			fmt.Printf("Formato de parámetro inválido: %s\n", match)
			continue
		}
		key, value := strings.ToLower(kv[0]), kv[1]
		switch key {
		case "-id":
			if value == "" {
				fmt.Println("Error: el ID no puede estar vacío")
				continue
			}
			params.Id = value
		case "-type":
			value = strings.ToUpper(value)
			if value != "FULL" {
				fmt.Println("Error: el tipo debe ser FULL")
				continue
			}
			params.T = value
		case "-fs":
			value = strings.ToLower(value)
			if value != "2fs" && value != "3fs" {
				fmt.Println("Error: el valor de -fs debe ser 2fs o 3fs")
				continue
			}
			params.Fs = value
		}
	}

	// Validación final de los parámetros obligatorios
	if params.Id == "" {
		fmt.Println("Error: Falta el parámetro obligatorio -id")
	}

	// Valor por defecto para type si no se proporcionó
	if params.T == "" {
		params.T = "FULL"
	}
	// Valor por defecto para fs si no se proporcionó
	if params.Fs == "" {
		params.Fs = "2fs"
	}

	return params
}


func (m *Mkfs) Mkfs(id string, t string) bool {
	fs := m.Params.Fs
	// Depuración: Mostrar el ID recibido
	fmt.Printf("Depuración: ID recibido -> '%s'\n", id)

	// Comprobando que id no esté vacío
	if id == "" {
		consola.AddToConsole("Error: No se encontró el ID entre los comandos\n")
		return false
	}

	// Comprobando que type tenga un valor correcto
	if t != "full" && t != "FULL" && t != "" {
		consola.AddToConsole("Error: El valor del comando type no es permitido\n")
		return false
	}
	if t == "" || t == "full" {
		t = "FULL"
	}

	// Depuración: Mostrar todas las particiones montadas antes de buscar el nodo
	fmt.Println("Depuración: Lista de particiones montadas antes de buscar el nodo:")
	lista.ListaMount.PrintList() // Asegúrate de tener un método PrintAll en ListaMount para imprimir todos los montajes

	// Creando nuestro nodo auxiliar
	nodo := lista.ListaMount.GetNodeById(id)

	// Depuración: Verificar si encontró el nodo
	if nodo == nil {
		fmt.Printf("Error: El ID '%s' no coincide con ninguna partición montada\n", id)
		consola.AddToConsole(fmt.Sprintf("El ID '%s' no coincide con ninguna partición montada\n", id))
		return false
	} else {
		fmt.Printf("Depuración: Nodo encontrado -> ID: %s, Ruta: %s\n", id, nodo.Ruta)
	}

	// Lógica para elegir EXT2 o EXT3 según el parámetro -fs
	if fs == "3fs" {
		m.Ext3(nodo)
	} else {
		m.Ext2(nodo)
	}
	return true
}

func (m *Mkfs) Ext3(nodo *lista.MountNode) {
	whereToStart := 0
	partSize := 0
	if nodo.Value != nil {
		whereToStart = int(nodo.Value.Part_start)
		partSize = int(nodo.Value.Part_size)
	} else if nodo.ValueL != nil {
		whereToStart = int(nodo.ValueL.Part_start) + int(unsafe.Sizeof(datos.EBR{}))
		partSize = int(nodo.ValueL.Part_size)
	}
	const JOURNAL_ENTRIES = 50
	journalSize := JOURNAL_ENTRIES * int(unsafe.Sizeof(datos.Journal{}))
	n := float64(float64(
		partSize - int(unsafe.Sizeof(datos.SuperBloque{})) - journalSize,
	) / float64(
		4 + int(unsafe.Sizeof(datos.TablaInodo{})) + 3*int(unsafe.Sizeof(datos.BloqueDeArchivos{})),
	))
	if math.Floor(n) < 1 {
		consola.AddToConsole("el tamano de la particion es mas pequeno que el sistema de archivos\n")
		return
	}
	inodesQuantity := int64(math.Floor(n))
	blocksQuantity := int64(3 * inodesQuantity)

	// llenando la estructura del superbloque
	superBlock := datos.SuperBloque{
		S_filesystem_type:   3,
		S_inodes_count:      inodesQuantity,
		S_blocks_count:      blocksQuantity,
		S_free_inodes_count: inodesQuantity - 2,
		S_free_blocks_count: blocksQuantity - 2,
		S_mnt_count:         0,
		S_magic:             0xEF53,
		S_inode_size:        int64(unsafe.Sizeof(datos.TablaInodo{})),
		S_block_size:        int64(unsafe.Sizeof(datos.BloqueDeArchivos{})),
		S_first_ino:         2,
		S_first_blo:         2,
	}
	superBlock.S_bm_inode_start = int64(whereToStart) + int64(unsafe.Sizeof(datos.SuperBloque{})) + int64(journalSize)
	superBlock.S_bm_block_start = superBlock.S_bm_inode_start + inodesQuantity
	superBlock.S_inode_start = superBlock.S_bm_block_start + blocksQuantity
	superBlock.S_block_start = superBlock.S_inode_start + int64(unsafe.Sizeof(datos.TablaInodo{})*uintptr(inodesQuantity))
	date := time.Now()
	for i := 0; i < len(superBlock.S_mtime)-1; i++ {
		superBlock.S_mtime[i] = date.String()[i]
	}

	// Escribiendo el superbloque
	Fwrite(&superBlock, nodo.Ruta, int64(whereToStart))

	// Inicializa el Journal (arreglo de 50 entradas vacías)
	for i := 0; i < JOURNAL_ENTRIES; i++ {
		journal := datos.Journal{
			J_count:   int64(i),
			J_content: datos.Information{}, // Estructura vacía
		}
		Fwrite(&journal, nodo.Ruta, int64(whereToStart)+int64(unsafe.Sizeof(datos.SuperBloque{}))+int64(i*int(unsafe.Sizeof(datos.Journal{}))))
	}

	// Buffers para bitmaps
	inodos := make([]byte, inodesQuantity)
	bloques := make([]byte, blocksQuantity)
	for i := 0; i < len(inodos); i++ {
		inodos[i] = '0'
	}
	for i := 0; i < len(bloques); i++ {
		bloques[i] = '0'
	}
	// inodos ocupados
	inodos[0] = '1'
	inodos[1] = '1'
	Fwrite(&inodos, nodo.Ruta, superBlock.S_bm_inode_start)
	// bloques ocupados
	bloques[0] = '1'
	bloques[1] = '1'
	Fwrite(&bloques, nodo.Ruta, superBlock.S_bm_block_start)

	// crear tabla de inodos root
	rootInodeTable := datos.TablaInodo{
		I_uid:  1,
		I_gid:  1,
		I_size: 0,
		I_type: '0',
		I_perm: 664,
	}
	atime := time.Now()
	for i := 0; i < len(rootInodeTable.I_atime)-1; i++ {
		rootInodeTable.I_atime[i] = atime.String()[i]
	}
	ctime := time.Now()
	for i := 0; i < len(rootInodeTable.I_atime)-1; i++ {
		rootInodeTable.I_ctime[i] = ctime.String()[i]
	}
	mtime := time.Now()
	for i := 0; i < len(rootInodeTable.I_atime)-1; i++ {
		rootInodeTable.I_mtime[i] = mtime.String()[i]
	}
	for i := 0; i < len(rootInodeTable.I_block); i++ {
		rootInodeTable.I_block[i] = -1
	}
	rootInodeTable.I_block[0] = 0
	Fwrite(&rootInodeTable, nodo.Ruta, superBlock.S_inode_start)

	// bloque de carpetas root
	bloqueCarpetasRoot := datos.BloqueDeCarpetas{}
	copy(bloqueCarpetasRoot.B_content[0].B_name[:], ".")
	bloqueCarpetasRoot.B_content[0].B_inodo = 0
	copy(bloqueCarpetasRoot.B_content[1].B_name[:], "..")
	bloqueCarpetasRoot.B_content[1].B_inodo = 0
	copy(bloqueCarpetasRoot.B_content[2].B_name[:], "users.txt")
	bloqueCarpetasRoot.B_content[2].B_inodo = 1
	copy(bloqueCarpetasRoot.B_content[3].B_name[:], "")
	bloqueCarpetasRoot.B_content[3].B_inodo = -1

	// users.txt
	content := "1,G,root\n1,U,root,root,123\n"
	fileInodeTable := datos.TablaInodo{
		I_uid:  1,
		I_gid:  1,
		I_size: 0,
		I_type: '1',
		I_perm: 664,
	}
	atime = time.Now()
	for i := 0; i < len(fileInodeTable.I_atime)-1; i++ {
		fileInodeTable.I_atime[i] = atime.String()[i]
	}
	ctime = time.Now()
	for i := 0; i < len(fileInodeTable.I_atime)-1; i++ {
		fileInodeTable.I_ctime[i] = ctime.String()[i]
	}
	mtime = time.Now()
	for i := 0; i < len(fileInodeTable.I_atime)-1; i++ {
		fileInodeTable.I_mtime[i] = mtime.String()[i]
	}
	for i := 0; i < len(fileInodeTable.I_block); i++ {
		fileInodeTable.I_block[i] = -1
	}
	fileInodeTable.I_block[0] = 1
	bloqueArchivos := datos.BloqueDeArchivos{}
	copy(bloqueArchivos.B_content[:], []byte(content))
	Fwrite(&bloqueCarpetasRoot, nodo.Ruta, superBlock.S_block_start)
	Fwrite(&fileInodeTable, nodo.Ruta, superBlock.S_inode_start+int64(unsafe.Sizeof(datos.TablaInodo{})))
	Fwrite(&bloqueArchivos, nodo.Ruta, superBlock.S_block_start+int64(unsafe.Sizeof(datos.BloqueDeArchivos{})))
	consola.AddToConsole("El formateo EXT3 fue exitoso\n")
}


func (m *Mkfs) Ext2(nodo *lista.MountNode) {
	whereToStart := 0
	partSize := 0
	if nodo.Value != nil {
		whereToStart = int(nodo.Value.Part_start)
		partSize = int(nodo.Value.Part_size)
	} else if nodo.ValueL != nil {
		whereToStart = int(nodo.ValueL.Part_start) + int(unsafe.Sizeof(datos.EBR{}))
		partSize = int(nodo.ValueL.Part_size)
	}
	n := float64(float64(partSize-int(unsafe.Sizeof(datos.SuperBloque{}))) / float64(4+int(unsafe.Sizeof(datos.TablaInodo{}))+3*int(unsafe.Sizeof(datos.BloqueDeArchivos{}))))
	// fmt.Println(math.Floor(n))
	if math.Floor(n) < 1 {
		consola.AddToConsole("el tamano de la particion es mas pequeno que el sistema de archivos\n")
		return
	}
	inodesQuantity := int64(math.Floor(n))
	blocksQuantity := int64(3 * inodesQuantity)

	// llenando la estructura del superbloque
	superBlock := datos.SuperBloque{
		S_filesystem_type:   2,
		S_inodes_count:      inodesQuantity,
		S_blocks_count:      blocksQuantity,
		S_free_inodes_count: inodesQuantity - 2,
		S_free_blocks_count: blocksQuantity - 2,
		S_mnt_count:         0,
		S_magic:             0xEF53,
		S_inode_size:        int64(unsafe.Sizeof(datos.TablaInodo{})),
		S_block_size:        int64(unsafe.Sizeof(datos.BloqueDeArchivos{})),
		S_first_ino:         2,
		S_first_blo:         2,
	}
	superBlock.S_bm_inode_start = int64(whereToStart) + int64(unsafe.Sizeof(datos.SuperBloque{}))
	superBlock.S_bm_block_start = superBlock.S_bm_inode_start + inodesQuantity
	superBlock.S_inode_start = superBlock.S_bm_block_start + blocksQuantity
	superBlock.S_block_start = superBlock.S_inode_start + int64(unsafe.Sizeof(datos.TablaInodo{})*uintptr(inodesQuantity))
	date := time.Now()
	for i := 0; i < len(superBlock.S_mtime)-1; i++ {
		superBlock.S_mtime[i] = date.String()[i]
	}

	// escribiendo el superbloque
	Fwrite(&superBlock, nodo.Ruta, int64(whereToStart))

	// buffers para bloques e inodos
	inodos := make([]byte, inodesQuantity)
	bloques := make([]byte, blocksQuantity)

	// llenando los buffers
	for i := 0; i < len(inodos); i++ {
		inodos[i] = '0'
	}
	for i := 0; i < len(bloques); i++ {
		bloques[i] = '0'
	}

	// inodos ocupados
	inodos[0] = '1'
	inodos[1] = '1'
	Fwrite(&inodos, nodo.Ruta, superBlock.S_bm_inode_start)

	// bloques ocupados
	bloques[0] = '1'
	bloques[1] = '1'
	Fwrite(&bloques, nodo.Ruta, superBlock.S_bm_block_start)

	// crear tabla de inodos root
	rootInodeTable := datos.TablaInodo{
		I_uid:  1,
		I_gid:  1,
		I_size: 0,
		I_type: '0',
		I_perm: 664,
	}
	// llenando las fechas
	atime := time.Now()
	for i := 0; i < len(rootInodeTable.I_atime)-1; i++ {
		rootInodeTable.I_atime[i] = atime.String()[i]
	}
	ctime := time.Now()
	for i := 0; i < len(rootInodeTable.I_atime)-1; i++ {
		rootInodeTable.I_ctime[i] = ctime.String()[i]
	}
	mtime := time.Now()
	for i := 0; i < len(rootInodeTable.I_atime)-1; i++ {
		rootInodeTable.I_mtime[i] = mtime.String()[i]
	}
	// llenando a todos los bloques no utilizados
	for i := 0; i < len(rootInodeTable.I_block); i++ {
		rootInodeTable.I_block[i] = -1
	}
	// apuntando al bloque 0 (bloque de carpetas root)
	rootInodeTable.I_block[0] = 0

	// escribiendo la tabla de inodos root
	Fwrite(&rootInodeTable, nodo.Ruta, superBlock.S_inode_start)

	// creando el bloque de carpetas root
	bloqueCarpetasRoot := datos.BloqueDeCarpetas{}

	copy(bloqueCarpetasRoot.B_content[0].B_name[:], ".")
	bloqueCarpetasRoot.B_content[0].B_inodo = 0

	copy(bloqueCarpetasRoot.B_content[1].B_name[:], "..")
	bloqueCarpetasRoot.B_content[1].B_inodo = 0

	copy(bloqueCarpetasRoot.B_content[2].B_name[:], "users.txt")
	bloqueCarpetasRoot.B_content[2].B_inodo = 1

	copy(bloqueCarpetasRoot.B_content[3].B_name[:], "")
	bloqueCarpetasRoot.B_content[3].B_inodo = -1

	// llenando el archivo users.txt
	content := "1,G,root\n1,U,root,root,123\n"

	// crear tabla de inodos de archivo
	fileInodeTable := datos.TablaInodo{
		I_uid:  1,
		I_gid:  1,
		I_size: 0,
		I_type: '1',
		I_perm: 664,
	}
	// llenando las fechas
	atime = time.Now()
	for i := 0; i < len(fileInodeTable.I_atime)-1; i++ {
		fileInodeTable.I_atime[i] = atime.String()[i]
	}
	ctime = time.Now()
	for i := 0; i < len(fileInodeTable.I_atime)-1; i++ {
		fileInodeTable.I_ctime[i] = ctime.String()[i]
	}
	mtime = time.Now()
	for i := 0; i < len(fileInodeTable.I_atime)-1; i++ {
		fileInodeTable.I_mtime[i] = mtime.String()[i]
	}
	// llenando a todos los bloques no utilizados
	for i := 0; i < len(fileInodeTable.I_block); i++ {
		fileInodeTable.I_block[i] = -1
	}
	// apuntando al bloque 1 (primer bloque de archivos creado para users.txt)
	fileInodeTable.I_block[0] = 1

	// crear bloque de archivos y escribiendo el contenido
	bloqueArchivos := datos.BloqueDeArchivos{}
	copy(bloqueArchivos.B_content[:], []byte(content))

	// escribiendo el bloque de carpetas root
	Fwrite(&bloqueCarpetasRoot, nodo.Ruta, superBlock.S_block_start)

	// escribiendo la tabla de inodos del archivo users.txt
	Fwrite(&fileInodeTable, nodo.Ruta, superBlock.S_inode_start+int64(unsafe.Sizeof(datos.TablaInodo{})))

	// escribiendo el bloque 1 del archivo users.txt
	Fwrite(&bloqueArchivos, nodo.Ruta, superBlock.S_block_start+int64(unsafe.Sizeof(datos.BloqueDeArchivos{})))

	if nodo.Value != nil {
		// aqui deberia de ir un metodo para guardar para la consola
		fmt.Println("")
	} else if nodo.ValueL != nil {
		// aqui igual deberia de ir
		fmt.Println("")
	}
	consola.AddToConsole("El formateo fue exitoso\n")
}

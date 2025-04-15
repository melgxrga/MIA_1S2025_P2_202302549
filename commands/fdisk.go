package comandos

import (
	"bytes"
	"fmt"
	"regexp" // Paquete para trabajar con expresiones regulares, útil para encontrar y manipular patrones en cadenas
	"strconv"
	"strings"
	"unsafe"

	"github.com/melgxrga/proyecto1Archivos/consola"
	"github.com/melgxrga/proyecto1Archivos/functions"
	datos "github.com/melgxrga/proyecto1Archivos/structures"
)

type ParametrosFdisk struct {
	Size int
	Unit byte
	Path string
	Type byte
	Fit  byte
	Name [16]byte
	Add  int
}

type Fdisk struct {
	Params ParametrosFdisk
}

func (f *Fdisk) Exe(parametros []string) {
	var eliminado bool
	f.Params, eliminado = f.SaveParams(parametros)
	if eliminado {
		return
	}
	if f.Fdisk(f.Params.Name, f.Params.Path, f.Params.Size, f.Params.Unit, f.Params.Fit, f.Params.Type) {
		consola.AddToConsole(fmt.Sprintf("\nfdisk realizado con exito para la ruta: %s y particion: %s\n\n", f.Params.Path, string(f.Params.Name[:])))
	} else {
		consola.AddToConsole(fmt.Sprintf("\n[ERROR!] no se logro realizar el comando fdisk para la ruta: %s\n\n", f.Params.Path))
	}
}
func (f *Fdisk) SaveParams(parametros []string) (ParametrosFdisk, bool) {
	var params ParametrosFdisk
	// Unir todos los parámetros en una sola cadena
	args := strings.Join(parametros, " ")

	// Expresión regular para capturar los parámetros
	re := regexp.MustCompile(`-size=\d+|-unit=[kKmM]|-fit=[bBfFwW]{2}|-path="[^"]+"|-path=[^\s]+|-type=[pPeElL]|-name="[^"]+"|-name=[^\s]+|-delete=[^\s]+|-add=-?\d+`)
	matches := re.FindAllString(args, -1)

	// Iterar sobre cada coincidencia
	var paramsDelete bool = false
	var deleteType string = ""
	var deleteName string = ""

	for _, match := range matches {
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			fmt.Printf("Formato de parámetro inválido: %s\n", match)
			continue
		}
		key, value := strings.ToLower(kv[0]), kv[1]

		// Quitar comillas si las tiene
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		// Procesar según el parámetro encontrado
		switch key {
		case "-size":
			size, err := strconv.Atoi(value)
			if err != nil || size <= 0 {
				fmt.Println("Error: el tamaño debe ser un número entero positivo")
				continue
			}
			params.Size = size
		case "-unit":
			value = strings.ToUpper(value)
			if value != "B" && value != "K" && value != "M" {
				fmt.Println("Error: la unidad debe ser B (bytes), K (Kilobytes) o M (Megabytes)")
				continue
			}
			params.Unit = value[0]
		case "-fit":
			value = strings.ToUpper(value)
			if value != "BF" && value != "FF" && value != "WF" {
				fmt.Println("Error: el ajuste debe ser BF, FF o WF")
				continue
			}
			params.Fit = value[0]
		case "-path":
			if value == "" {
				fmt.Println("Error: el path no puede estar vacío")
				continue
			}
			params.Path = value
		case "-type":
			value = strings.ToUpper(value)
			if value != "P" && value != "E" && value != "L" {
				fmt.Println("Error: el tipo debe ser P, E o L")
				continue
			}
			params.Type = value[0]
		case "-name":
			if value == "" {
				fmt.Println("Error: el nombre no puede estar vacío")
				continue
			}
			copy(params.Name[:], value)
			deleteName = value
		case "-add":
			addVal, err := strconv.Atoi(value)
			if err != nil {
				fmt.Println("Error: el valor de -add debe ser un número entero (positivo o negativo)")
				continue
			}
			params.Add = addVal
		case "-delete":
			if value == "" {
				fmt.Println("Error: El valor de -delete no puede estar vacío")
				continue
			}
			paramsDelete = true
			deleteType = strings.ToLower(value)
		}
	}

	if paramsDelete {
		if params.Path == "" || deleteName == "" {
			fmt.Println("Error: Debe especificar -path y -name para eliminar una partición.")
			return params, true
		}
		if deleteType != "fast" && deleteType != "full" {
			fmt.Println("Error: El valor de -delete debe ser 'fast' o 'full'")
			return params, true
		}
		var name [16]byte
		copy(name[:], deleteName)
		master := GetMBR(params.Path)
		if !ExisteParticion(&master, name) {
			fmt.Printf("Error: No se encontró una partición con el nombre: %s\n", deleteName)
			return params, true
		}
		success := f.EliminarParticion(&master, params.Path, name)
		if success {
			fmt.Printf("La partición %s fue eliminada correctamente con método %s.\n", deleteName, deleteType)
		} else {
			fmt.Printf("Hubo un problema al eliminar la partición %s.\n", deleteName)
		}
		return params, true
	}

	if params.Size == 0 {
		fmt.Println("Error: Falta el parámetro obligatorio -size")
	}
	if params.Path == "" {
		fmt.Println("Error: Falta el parámetro obligatorio -path")
	}

	// Valores por defecto si no se proporcionaron
	if params.Unit == 0 {
		params.Unit = 'M' // Valor por defecto: Megabytes
	}
	if params.Fit == 0 {
		params.Fit = 'F' // Valor por defecto: First Fit
	}
	if params.Type == 0 {
		params.Type = 'P' // Valor por defecto: Partición primaria
	}

	return params, false
}

func (f *Fdisk) Fdisk(name [16]byte, path string, size int, unit byte, fit byte, t byte) bool {
	if f.Params.Add != 0 {
		master := GetMBR(f.Params.Path)
		var idx = -1
		for i, v := range master.Mbr_partitions {
			if string(v.Part_name[:]) == string(f.Params.Name[:]) && v.Part_status == '1' {
				idx = i
				break
			}
		}
		if idx == -1 {
			consola.AddToConsole(fmt.Sprintf("No se encontró una partición activa con el nombre: %s\n", string(f.Params.Name[:])))
			return false
		}
		consola.AddToConsole("\nEstado de las particiones ANTES de la operación:\n")
		PrintPartitions(&master)
		addBytes := f.Params.Add
		switch f.Params.Unit {
		case 'b', 'B':
		case 'k', 'K':
			addBytes = addBytes * 1024
		case 'm', 'M', 0:
			addBytes = addBytes * 1024 * 1024
		}
		if addBytes > 0 {
			endCurrent := master.Mbr_partitions[idx].Part_start + master.Mbr_partitions[idx].Part_size
			var nextStart int64 = int64(master.Mbr_tamano)
			for j := idx + 1; j < len(master.Mbr_partitions); j++ {
				if master.Mbr_partitions[j].Part_status == '1' {
					nextStart = master.Mbr_partitions[j].Part_start
					break
				}
			}
			freeSpace := nextStart - endCurrent
			if endCurrent+int64(addBytes) > nextStart {
				consola.AddToConsole(fmt.Sprintf("No hay suficiente espacio libre después de la partición para agregar el tamaño solicitado. Espacio libre real: %d bytes (%.2f KB, %.2f MB)\n", freeSpace, float64(freeSpace)/1024.0, float64(freeSpace)/(1024.0*1024.0)))
				return false
			}
			master.Mbr_partitions[idx].Part_size += int64(addBytes)
		} else if addBytes < 0 {
			// Quitar espacio: verificar que no quede tamaño negativo o cero
			if master.Mbr_partitions[idx].Part_size+int64(addBytes) <= 0 {
				consola.AddToConsole("No se puede quitar más espacio del que tiene la partición\n")
				return false
			}
			master.Mbr_partitions[idx].Part_size += int64(addBytes)
		}
		// Reajustar particiones siguientes
		f.ReajustarParticiones(&master)
		WriteMBR(&master, f.Params.Path)
		consola.AddToConsole("\nEstado de las particiones DESPUÉS de la operación:\n")
		PrintPartitions(&master)
		consola.AddToConsole(fmt.Sprintf("El tamaño de la partición %s fue actualizado correctamente.\n", string(f.Params.Name[:])))
		return true
	}
	if path == "" {
		consola.AddToConsole("no se encontro una ruta\n")
		return false
	}
	master := GetMBR(path)
	newPartition := datos.Partition{}
	fileSize := 0

	// Handle unit parameter according to specifications
	switch {
	case unit == 'b' || unit == 'B':
		fileSize = size
	case unit == 'k' || unit == 'K':
		fileSize = size * 1024
	case unit == 'm' || unit == 'M':
		fileSize = size * 1024 * 1024
	case unit == 0:
		fileSize = size * 1024
	default:
		consola.AddToConsole("Error: unidad no válida. Debe ser B (bytes), K (Kilobytes) o M (Megabytes)\n")
		return false
	}

	if ExisteParticion(&master, name) {
		consola.AddToConsole(fmt.Sprintf("ya existe una particion con nombre: \"%s\"\n", string(functions.TrimArray(name[:]))))
		return false
	}
	if size <= 0 {
		consola.AddToConsole("el tamano de la particion tiene que ser mayor a 0\n")
		return false
	}
	if strconv.Itoa(int(fit)) == "87" || fit == 'W' {
		newPartition.Part_fit = 'w'
	} else if strconv.Itoa(int(fit)) == "66" || fit == 'B' {
		newPartition.Part_fit = 'b'
	} else if strconv.Itoa(int(fit)) == "70" || fit == 'F' {
		newPartition.Part_fit = 'f'
	} else if fit == 0 {
		newPartition.Part_fit = 'w'
	} else {
		consola.AddToConsole("se debe ingresar un tipo de fit valido\n")
		return false
	}
	// verificando que el tamano de la particion a crear sea menor
	// o igual que el tamano que queda en el disco.
	totalSize := int(unsafe.Sizeof(datos.MBR{}))
	for _, v := range master.Mbr_partitions {
		if v.Part_status == '1' {
			totalSize += int(v.Part_size)
		}
	}
	// fmt.Println("espacio disponible, espacio a utilizar:", int(master.Mbr_tamano)-totalSize, fileSize)
	if t != 'l' && t != 'L' {
		if fileSize > int(master.Mbr_tamano)-int(totalSize) {
			consola.AddToConsole("el tamano de la particion es mas grande que el disco\n")
			return false
		}
	}

	// indicando el tipo de particion
	if t == 0 {
		t = 'p'
	} else if t != 'p' && t != 'e' && t != 'l' && t != 'P' && t != 'E' && t != 'L' {
		consola.AddToConsole(fmt.Sprintf("el tipo de la particion no es valido: \"%c\"\n", t))
		return false
	}
	newPartition.Part_size = int64(fileSize)
	newPartition.Part_type = t
	newPartition.Part_status = '1'
	copy(newPartition.Part_name[:], name[:])

	// revisando que no exista mas de una particion Extendida y que Exista en caso de que se vaya a crear una particion logica
	existeParticionExtendida := false //esta variable se utiliza para encontrar si existe una particion extendida
	var whereToStart int              // con este valor le vamos a pasar a la particion logica donde comienza la particion extendida
	var partitionSize int             // con este valor le indicamos a la particion logica cuanto espacio ocupa la particion extendida
	var extendedFit byte              // con este valor le indicamos a la particion logica el tipo de ajuste que tiene la particion extendida
	var extendedName [16]byte         // con este valor le indicamos el nombre de la particion extendida a la particion logica
	// aqui le agregamos a las variables anteriores su correspondiente valor
	for _, v := range master.Mbr_partitions {
		if v.Part_type == 'e' || v.Part_type == 'E' {
			copy(extendedName[:], v.Part_name[:])
			existeParticionExtendida = true
			extendedFit = v.Part_fit
			whereToStart = int(v.Part_start)
			partitionSize = int(v.Part_size)
		}
	}

	// comprobamos que exista una particion libre
	existeParticionLibre := false
	if t != 'l' && t != 'L' {
		for _, v := range master.Mbr_partitions {
			if v.Part_status == '0' {
				existeParticionLibre = true
			}
		}
	} else if t == 'l' || t == 'L' {
		existeParticionLibre = true
	}
	// sino se encuentra un espacio libre para particion
	if !existeParticionLibre {
		consola.AddToConsole("no se encuentra ninguna particion libre para crear dentro del disco\n")
		return false
	}
	// comprobamos que tipo de particion es, luego la creamos
	if t == 'p' || t == 'P' {
		f.CreatePrimaryPartition(&master, newPartition)
	} else if t == 'e' || t == 'E' {
		if existeParticionExtendida {
			consola.AddToConsole("no puede haber mas de una particion extendida\n")
			return false
		}
		f.CreateExtendedPartition(&master, newPartition, path)
	} else if t == 'l' || t == 'L' {
		if !existeParticionExtendida {
			consola.AddToConsole("no existe una particion extendida para crear una particion logica\n")
			return false
		}
		particionLogica := datos.EBR{}
		particionLogica.Part_fit = newPartition.Part_fit
		particionLogica.Part_next = -1
		particionLogica.Part_size = newPartition.Part_size
		particionLogica.Part_status = newPartition.Part_status
		copy(particionLogica.Part_name[:], newPartition.Part_name[:])
		// vamos a mandar que tipo de ajuste tiene la particion
		// dentro de este metodo se le indica donde es que comienza la particion logica
		return f.CreateLogicPartition(&particionLogica, path, whereToStart, partitionSize, extendedFit, extendedName)

	}
	WriteMBR(&master, path)
	PrintPartitions(&master)
	return true
}

func (f *Fdisk) CreatePrimaryPartition(master *datos.MBR, newPartition datos.Partition) {
	// Asignacion de que particion es la que se utilizara
	if master.Dsk_fit == 'b' {
		BestFit(master, &newPartition)
	} else if master.Dsk_fit == 'w' {
		WorstFit(master, &newPartition)
	} else if master.Dsk_fit == 'f' {
		FirstFit(master, &newPartition)
	}
}

func (f Fdisk) CreateExtendedPartition(master *datos.MBR, newPartition datos.Partition, path string) {
	// Asignacion de que particion es la que se utilizara
	if master.Dsk_fit == 'b' {
		BestFit(master, &newPartition)
	} else if master.Dsk_fit == 'w' {
		WorstFit(master, &newPartition)
	} else if master.Dsk_fit == 'f' {
		FirstFit(master, &newPartition)
	}
	temp := datos.EBR{}
	temp.Part_status = '0'
	temp.Part_fit = '0'
	temp.Part_start = newPartition.Part_start
	temp.Part_size = 0
	temp.Part_next = -1
	copy(temp.Part_name[:], "vacio")
	WriteEBR(&temp, path, newPartition.Part_start)
}
func FirstFit(master *datos.MBR, newPartition *datos.Partition) {
	firstFit := 0
	for i, v := range master.Mbr_partitions {
		if v.Part_status == '5' && v.Part_size >= newPartition.Part_size || v.Part_start == 0 {
			firstFit = i
			break
		}
	}
	master.Mbr_partitions[firstFit] = *newPartition
	if firstFit == 0 {
		master.Mbr_partitions[firstFit].Part_start = int64(unsafe.Sizeof(datos.MBR{}))
		newPartition.Part_start = int64(unsafe.Sizeof(datos.MBR{}))
	} else {
		master.Mbr_partitions[firstFit].Part_start = master.Mbr_partitions[firstFit-1].Part_start + master.Mbr_partitions[firstFit-1].Part_size
		newPartition.Part_start = master.Mbr_partitions[firstFit-1].Part_start + master.Mbr_partitions[firstFit-1].Part_size
	}
}

func BestFit(master *datos.MBR, newPartition *datos.Partition) {
	bestFit := 0
	// para encontrar el mejor fit lo primero que hay que hacer
	// es recorrer la lista de particiones y verificar que exista
	// una particion disponible, si esta particion se encuentra
	// disponible se comprobara que el tamano de esta sea mayor o
	// igual al tamano de la particion que estamos creando. de ser
	// asi, le asignaremos esa posicion, luego seguir iterando para
	// buscar si existe alguna particion con menor cantidad de espacio
	// donde se ajuste la particion que estamos creando.
	encontroParticion := false
	for i, v := range master.Mbr_partitions {
		if v.Part_status == '5' && v.Part_size >= newPartition.Part_size {
			if i != bestFit {
				if v.Part_size < master.Mbr_partitions[bestFit].Part_size {
					encontroParticion = true
					bestFit = i
				}
			}
		}
	}
	if !encontroParticion {
		for i, v := range master.Mbr_partitions {
			if v.Part_start == 0 {
				bestFit = i
				break
			}
		}
	}
	master.Mbr_partitions[bestFit] = *newPartition
	if bestFit == 0 {
		master.Mbr_partitions[bestFit].Part_start = int64(unsafe.Sizeof(datos.MBR{}))
		newPartition.Part_start = int64(unsafe.Sizeof(datos.MBR{}))
	} else {
		master.Mbr_partitions[bestFit].Part_start = master.Mbr_partitions[bestFit-1].Part_start + master.Mbr_partitions[bestFit-1].Part_size
		newPartition.Part_start = master.Mbr_partitions[bestFit-1].Part_start + master.Mbr_partitions[bestFit-1].Part_size
	}
}

func WorstFit(master *datos.MBR, newPartition *datos.Partition) {
	worstFit := 0
	encontroParticion := false
	for i, v := range master.Mbr_partitions {
		if v.Part_status == '5' && v.Part_size >= newPartition.Part_size {
			if i != worstFit {
				if v.Part_size > master.Mbr_partitions[worstFit].Part_size {
					worstFit = i
					encontroParticion = true
				}
			}
		}
	}
	if !encontroParticion {
		for i, v := range master.Mbr_partitions {
			fmt.Println(v.Part_start)
			if v.Part_start == 0 {
				worstFit = i
				break
			}
		}
	}
	master.Mbr_partitions[worstFit] = *newPartition
	if worstFit == 0 {
		master.Mbr_partitions[worstFit].Part_start = int64(unsafe.Sizeof(datos.MBR{}))
		newPartition.Part_start = int64(unsafe.Sizeof(datos.MBR{}))
	} else {
		master.Mbr_partitions[worstFit].Part_start = master.Mbr_partitions[worstFit-1].Part_start + master.Mbr_partitions[worstFit-1].Part_size
		newPartition.Part_start = master.Mbr_partitions[worstFit-1].Part_start + master.Mbr_partitions[worstFit-1].Part_size
	}
}

func (f *Fdisk) EliminarParticionLogica(master *datos.MBR, path string, name [16]byte) bool {
	// Buscar la partición extendida
	var foundExtended bool
	for _, v := range master.Mbr_partitions {
		if v.Part_type == 'e' || v.Part_type == 'E' {
			foundExtended = true
			break
		}
	}

	// Si no hay partición extendida, no podemos eliminar una partición lógica
	if !foundExtended {
		consola.AddToConsole("No se encontró una partición extendida para eliminar particiones lógicas\n")
		return false
	}

	// Buscar y eliminar la partición lógica dentro de la partición extendida
	var logicalPartitionIndex = -1
	for i, v := range master.Mbr_partitions {
		if v.Part_type == 'l' || v.Part_type == 'L' {
			if string(v.Part_name[:]) == string(name[:]) {
				logicalPartitionIndex = i
				break
			}
		}
	}

	// Si no se encuentra la partición lógica, no se puede eliminar
	if logicalPartitionIndex == -1 {
		consola.AddToConsole(fmt.Sprintf("No se encontró una partición lógica con el nombre: %s\n", string(name[:])))
		return false
	}

	// Marcar la partición lógica como eliminada
	master.Mbr_partitions[logicalPartitionIndex].Part_status = '0'

	// Si se elimina la partición lógica, se actualizan las particiones en el disco
	// (No se reajustan las primarias/extendidas, solo se marca como eliminada la lógica)
	WriteMBR(master, path)
	PrintPartitions(master)
	consola.AddToConsole(fmt.Sprintf("Partición lógica %s eliminada exitosamente\n", string(name[:])))
	return true
}

func (f *Fdisk) EliminarParticionExtendida(master *datos.MBR, path string, name [16]byte) bool {
	// Buscar la partición extendida
	var extendedPartitionIndex = -1
	for i, v := range master.Mbr_partitions {
		if v.Part_type == 'e' || v.Part_type == 'E' {
			if string(v.Part_name[:]) == string(name[:]) {
				extendedPartitionIndex = i
				break
			}
		}
	}

	// Si no se encuentra la partición extendida, no se puede eliminar
	if extendedPartitionIndex == -1 {
		consola.AddToConsole(fmt.Sprintf("No se encontró una partición extendida con el nombre: %s\n", string(name[:])))
		return false
	}

	// Eliminar todas las particiones lógicas dentro de la partición extendida
	for i, v := range master.Mbr_partitions {
		if (v.Part_type == 'l' || v.Part_type == 'L') && v.Part_start > master.Mbr_partitions[extendedPartitionIndex].Part_start && v.Part_start < (master.Mbr_partitions[extendedPartitionIndex].Part_start+master.Mbr_partitions[extendedPartitionIndex].Part_size) {
			master.Mbr_partitions[i].Part_status = '0' // Marcar como eliminada
		}
	}

	// Marcar la partición extendida como eliminada
	master.Mbr_partitions[extendedPartitionIndex].Part_status = '0'

	// Escribir los cambios en el disco
	f.ReajustarParticiones(master)
	WriteMBR(master, path)
	PrintPartitions(master)
	consola.AddToConsole(fmt.Sprintf("Partición extendida %s y todas sus particiones lógicas eliminadas exitosamente\n", string(name[:])))
	return true
}
func (f *Fdisk) EliminarParticionPrimaria(master *datos.MBR, path string, name [16]byte) bool {
	// Buscar la partición primaria
	var primaryPartitionIndex = -1
	for i, v := range master.Mbr_partitions {
		if v.Part_type == 'p' || v.Part_type == 'P' {
			if string(v.Part_name[:]) == string(name[:]) {
				primaryPartitionIndex = i
				break
			}
		}
	}

	// Si no se encuentra la partición primaria, no se puede eliminar
	if primaryPartitionIndex == -1 {
		consola.AddToConsole(fmt.Sprintf("No se encontró una partición primaria con el nombre: %s\n", string(name[:])))
		return false
	}

	// Marcar la partición primaria como eliminada
	master.Mbr_partitions[primaryPartitionIndex].Part_status = '0'

	// Escribir los cambios en el disco
	f.ReajustarParticiones(master)
	WriteMBR(master, path)
	PrintPartitions(master)
	consola.AddToConsole(fmt.Sprintf("Partición primaria %s eliminada exitosamente\n", string(name[:])))
	return true
}

func (f *Fdisk) ReajustarParticiones(master *datos.MBR) {
	// Crear un slice temporal para las particiones activas
	var activas []datos.Partition
	for _, part := range master.Mbr_partitions {
		if part.Part_status == '1' {
			activas = append(activas, part)
		}
	}

	// Limpiar las particiones
	for i := range master.Mbr_partitions {
		master.Mbr_partitions[i] = datos.Partition{}
	}

	// Reasignar las particiones activas, recalculando Part_start
	for i := range activas {
		if i == 0 {
			activas[i].Part_start = int64(unsafe.Sizeof(datos.MBR{}))
		} else {
			activas[i].Part_start = activas[i-1].Part_start + activas[i-1].Part_size
		}
		master.Mbr_partitions[i] = activas[i]
	}
}

func (f *Fdisk) EliminarParticion(master *datos.MBR, path string, name [16]byte) bool {
	// Verificar si la partición es extendida, primaria o lógica
	for _, v := range master.Mbr_partitions {
		if string(v.Part_name[:]) == string(name[:]) {
			switch v.Part_type {
			case 'e', 'E': // Eliminar partición extendida y lógicas
				return f.EliminarParticionExtendida(master, path, name)
			case 'p', 'P': // Reajusta las particiones activas tras una eliminación, compactando y recalculando Part_start
				return f.EliminarParticionPrimaria(master, path, name)
			case 'l', 'L': // Reajusta las particiones activas tras una eliminación, compactando y recalculando Part_start
				return f.EliminarParticionLogica(master, path, name)
			}
		}
	}

	consola.AddToConsole(fmt.Sprintf("No se encontró una partición con el nombre: %s\n", string(name[:])))
	return false
}

func (f *Fdisk) CreateLogicPartition(logicPartition *datos.EBR, path string, whereToStart int, partitionSize int, extendedFit byte, extendedName [16]byte) bool {
	if extendedFit == 'f' {
		return FirstFitLogicPart(logicPartition, path, whereToStart, partitionSize, extendedName)
	} else if extendedFit == 'b' {
		return BestFitLogicPart(logicPartition, path, whereToStart, partitionSize, extendedName)
	} else if extendedFit == 'w' {
		return WorstFitLogicPart(logicPartition, path, whereToStart, partitionSize, extendedName)
	}
	return false
}

func FirstFitLogicPart(logicPartition *datos.EBR, path string, whereToStart int, partitionSize int, extendedName [16]byte) bool {
	temp := datos.EBR{}
	totalSize := 0
	totalSize += int(logicPartition.Part_size)
	temp = ReadEBR(path, int64(whereToStart))
	flag := true
	for flag {
		if temp.Part_size == 0 {
			if partitionSize < int(logicPartition.Part_size) {
				fmt.Println("la particion logica es mas grande que la extendida")
				return false
			}
			logicPartition.Part_start = int64(whereToStart)
			WriteEBR(logicPartition, path, int64(whereToStart))
			flag = false
		} else if temp.Part_status == '5' {
			if temp.Part_size >= logicPartition.Part_size {
				logicPartition.Part_start = temp.Part_start
				logicPartition.Part_next = temp.Part_next
				WriteEBR(logicPartition, path, temp.Part_start)
				flag = false
			}
		} else if temp.Part_next == -1 {
			totalSize += int(temp.Part_size)
			if partitionSize < totalSize {
				fmt.Println("el tamano de todas las particiones logicas unidas son mas grandes que la particion extendida, espacio insuficiente")
				return false
			}
			temp.Part_next = temp.Part_start + temp.Part_size
			logicPartition.Part_start = temp.Part_next
			WriteEBR(&temp, path, temp.Part_start)
			WriteEBR(logicPartition, path, temp.Part_next)
			flag = false
		} else {
			totalSize += int(temp.Part_size)
			temp = ReadEBR(path, temp.Part_next)
		}
	}
	// aqui deberia ir un print a la consola
	PrintLogicPartitions(path, int64(whereToStart), int64(partitionSize), extendedName)
	return true
}

func BestFitLogicPart(logicPartition *datos.EBR, path string, whereToStart int, partitionSize int, extendedName [16]byte) bool {
	var particionesLogicas []datos.EBR
	temp := datos.EBR{}
	totalSize := 0
	totalSize += int(logicPartition.Part_size)
	temp = ReadEBR(path, int64(whereToStart))
	Wrote := false
	flag := true
	for flag {
		if temp.Part_size == 0 {
			if partitionSize < int(logicPartition.Part_size) {
				fmt.Println("la particion logica es mas grande que la extendida")
				return false
			}
			logicPartition.Part_start = int64(whereToStart)
			WriteEBR(logicPartition, path, int64(whereToStart))
			flag = false
			Wrote = true
		} else if temp.Part_status == '5' {
			particionesLogicas = append(particionesLogicas, temp)
		} else if temp.Part_next == -1 {
			flag = false
		} else {
			totalSize += int(temp.Part_size)
			temp = ReadEBR(path, temp.Part_next)
		}
	}
	bestFit := 0
	tempSize := 0
	if len(particionesLogicas) != 0 {
		for i, v := range particionesLogicas {
			if tempSize != 0 {
				bestFit = i
			} else if tempSize > int(v.Part_size) && v.Part_size >= logicPartition.Part_size {
				tempSize = int(v.Part_size)
				bestFit = i
			}
		}
		logicPartition.Part_start = particionesLogicas[bestFit].Part_start
		logicPartition.Part_next = particionesLogicas[bestFit].Part_next
		WriteEBR(logicPartition, path, logicPartition.Part_start)
		Wrote = true
	}
	if !Wrote {
		totalSize = int(logicPartition.Part_size)
		temp = ReadEBR(path, int64(whereToStart))
		flag2 := true
		for flag2 {
			if temp.Part_next == -1 {
				totalSize += int(temp.Part_size)
				if partitionSize < totalSize {
					fmt.Println("el tamano de todas las particiones logicas unidas son mas grandes que la particion extendida, espacio insuficiente")
					return false
				}
				temp.Part_next = temp.Part_start + temp.Part_size
				logicPartition.Part_start = temp.Part_next
				WriteEBR(&temp, path, temp.Part_start)
				WriteEBR(logicPartition, path, temp.Part_next)
				flag2 = false
			} else {
				totalSize += int(temp.Part_size)
				temp = ReadEBR(path, temp.Part_next)
			}
		}
	}
	// aqui deberia ir un print a la consola
	PrintLogicPartitions(path, int64(whereToStart), int64(partitionSize), extendedName)
	return true
}

func WorstFitLogicPart(logicPartition *datos.EBR, path string, whereToStart int, partitionSize int, extendedName [16]byte) bool {
	var particionesLogicas []datos.EBR
	temp := datos.EBR{}
	totalSize := 0
	totalSize += int(logicPartition.Part_size)
	temp = ReadEBR(path, int64(whereToStart))
	Wrote := false
	flag := true
	for flag {
		if temp.Part_size == 0 {
			if partitionSize < int(logicPartition.Part_size) {
				fmt.Println("la particion logica es mas grande que la extendida")
				return false
			}
			logicPartition.Part_start = int64(whereToStart)
			WriteEBR(logicPartition, path, int64(whereToStart))
			flag = false
			Wrote = true
		} else if temp.Part_status == '5' {
			particionesLogicas = append(particionesLogicas, temp)
		} else if temp.Part_next == -1 {
			flag = false
		} else {
			totalSize += int(temp.Part_size)
			temp = ReadEBR(path, temp.Part_next)
		}
	}
	worstFit := 0
	tempSize := 0
	if len(particionesLogicas) != 0 {
		for i, v := range particionesLogicas {
			if tempSize != 0 {
				worstFit = i
			} else if tempSize < int(v.Part_size) && v.Part_size >= logicPartition.Part_size {
				tempSize = int(v.Part_size)
				worstFit = i
			}
		}
		logicPartition.Part_start = particionesLogicas[worstFit].Part_start
		logicPartition.Part_next = particionesLogicas[worstFit].Part_next
		WriteEBR(logicPartition, path, logicPartition.Part_start)
		Wrote = true
	}
	if !Wrote {
		totalSize = int(logicPartition.Part_size)
		temp = ReadEBR(path, int64(whereToStart))
		flag2 := true
		for flag2 {
			if temp.Part_next == -1 {
				totalSize += int(temp.Part_size)
				if partitionSize < totalSize {
					fmt.Println("el tamano de todas las particiones logicas unidas son mas grandes que la particion extendida, espacio insuficiente")
					return false
				}
				temp.Part_next = temp.Part_start + temp.Part_size
				logicPartition.Part_start = temp.Part_next
				WriteEBR(&temp, path, temp.Part_start)
				WriteEBR(logicPartition, path, temp.Part_next)
				flag2 = false
			} else {
				totalSize += int(temp.Part_size)
				temp = ReadEBR(path, temp.Part_next)
			}
		}
	}
	// aqui deberia ir un print a la consola
	PrintLogicPartitions(path, int64(whereToStart), int64(partitionSize), extendedName)
	return true
}

func ExisteParticion(master *datos.MBR, name [16]byte) bool {
	for _, v := range master.Mbr_partitions {
		if bytes.Equal(v.Part_name[:], name[:]) {
			return true
		}
	}
	return false
}

func PrintPartitions(master *datos.MBR) {
	str := ""
	for i := 0; i < 70; i++ {
		str += "-"
	}
	contenido := ""
	contenido += fmt.Sprintf("%s\n", str)
	contenido += fmt.Sprintf("%-20s%-10s%-10s%-10s%-10s%-10s\n", "Name", "Type", "Fit", "Start", "Size", "Status")
	for _, part := range master.Mbr_partitions {
		if part.Part_status != '1' {
			continue // Saltar particiones eliminadas
		}
		contenido += fmt.Sprintf("%s\n", str)
		if string(functions.TrimArray(part.Part_name[:])) == "" {
			contenido += fmt.Sprintf("%-20s", "-")
		} else {
			contenido += fmt.Sprintf("%-20s", string(functions.TrimArray(part.Part_name[:])))
		}
		if part.Part_type == '0' {
			contenido += fmt.Sprintf("%-10c", '-')
		} else {
			contenido += fmt.Sprintf("%-10c", part.Part_type)
		}
		if part.Part_fit == '0' {
			contenido += fmt.Sprintf("%-10c", '-')
		} else {
			contenido += fmt.Sprintf("%-10c", part.Part_fit)
		}
		contenido += fmt.Sprintf("%-10d", part.Part_start)
		contenido += fmt.Sprintf("%-10d", part.Part_size)
		contenido += fmt.Sprintf("%-10c\n", part.Part_status)
	}
	contenido += fmt.Sprintf("%s\n", str)
	consola.AddToConsole(contenido)
}

func PrintLogicPartitions(path string, whereToStart, PartitionSize int64, extendedName [16]byte) {
	str := ""
	for i := 0; i < 70; i++ {
		str += "-"
	}
	contenido := ""
	contenido += fmt.Sprintf("Partition name: %s\n", string(functions.TrimArray(extendedName[:])))
	contenido += fmt.Sprintf("Partition size: %d\n", PartitionSize)
	contenido += fmt.Sprintf("%s\n", str)
	contenido += fmt.Sprintf("%-20s%-12s%-10s%-10s%-10s%-10s\n", "Name", "Next Part", "Fit", "Start", "Size", "Status")
	var temp datos.EBR
	Fread(&temp, path, whereToStart)
	flag := true
	for flag {
		contenido += fmt.Sprintf("%s\n", str)
		if string(functions.TrimArray(temp.Part_name[:])) == "" {
			contenido += fmt.Sprintf("%-20s", "Disponible")
		} else {
			contenido += fmt.Sprintf("%-20s", string(functions.TrimArray(temp.Part_name[:])))
		}
		contenido += fmt.Sprintf("%-12d", temp.Part_next)
		if temp.Part_fit == '0' {
			contenido += fmt.Sprintf("%-10c", '-')
		} else {
			contenido += fmt.Sprintf("%-10c", temp.Part_fit)
		}
		contenido += fmt.Sprintf("%-10d", temp.Part_start)
		contenido += fmt.Sprintf("%-10d", temp.Part_size)
		contenido += fmt.Sprintf("%-10c\n", temp.Part_status)
		if temp.Part_next == -1 {
			flag = false
		} else {
			Fread(&temp, path, temp.Part_next)
		}
	}
	contenido += fmt.Sprintf("%s\n", str)
	consola.AddToConsole(contenido)
}

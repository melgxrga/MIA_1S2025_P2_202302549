package comandos

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"path"
	"reflect"
	"time"

	"github.com/melgxrga/proyecto1Archivos/consola"
	"github.com/melgxrga/proyecto1Archivos/structures"
)


func FileExist(path string) bool {
	fmt.Printf("-%s-\n", path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}

func WriteMBR(master *datos.MBR, path string) {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		consola.AddToConsole(fmt.Sprintf("no se pudo abrir el archivo para escribir el MBR %s\n", err.Error()))
		return
	}
	// Posicionandonos en el principio del archivo
	_, err = file.Seek(0, 0)
	if err != nil {
		consola.AddToConsole(fmt.Sprintf("no se pudo posicionar en el principio del archivo %s]n", err.Error()))
		return
	}
	// Escribiendo el MBR
	// var masterBuffer bytes.Buffer
	err = binary.Write(file, binary.LittleEndian, master)
	if err != nil {
		consola.AddToConsole(fmt.Sprintf("no se pudo escribir el MBR %s\n", err.Error()))
		file.Close()
		return
	}
	// consola.AddToConsole(fmt.Sprintf("se escribio correctamente! :D")
	defer file.Close()
}

func GetMBR(path string) datos.MBR {
	var mbr datos.MBR
	file, err := os.Open(path)
	if err != nil {
		consola.AddToConsole(fmt.Sprintf("no se pudo abrir el archivo para obtener el MBR, %s\n", err.Error()))
		return mbr
	}

	defer file.Close()

	// leyendo el mbr del archivo
	file.Seek(0, 0)
	err = binary.Read(file, binary.LittleEndian, &mbr)
	if err != nil {
		consola.AddToConsole(fmt.Sprintf("no se pudo obtener la informacion del archivo para obtener el MBR %s\n", err.Error()))
		return mbr
	}
	return mbr
}

func WriteEBR(ebr *datos.EBR, path string, position int64) {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		consola.AddToConsole(fmt.Sprintf("no se pudo abrir el archivo para escribir el MBR %s\n", err.Error()))
		return
	}
	// Posicionandonos en el principio del archivo
	_, err = file.Seek(position, 0)
	if err != nil {
		consola.AddToConsole(fmt.Sprintf("no se pudo posicionar en el principio del archivo %s\n", err.Error()))
		return
	}
	// Escribiendo el MBR
	// var masterBuffer bytes.Buffer
	err = binary.Write(file, binary.LittleEndian, ebr)
	if err != nil {
		consola.AddToConsole(fmt.Sprintf("no se pudo escribir el MBR %s\n", err.Error()))
		file.Close()
		return
	}
	// consola.AddToConsole(fmt.Sprintf("se escribio correctamente! :D")
	defer file.Close()
}

func ReadEBR(path string, position int64) datos.EBR {
	var ebr datos.EBR
	file, err := os.Open(path)
	if err != nil {
		consola.AddToConsole(fmt.Sprintf("no se pudo abrir el archivo para obtener el MBR %s\n", err.Error()))
		return ebr
	}

	defer file.Close()

	// leyendo el mbr del archivo
	file.Seek(position, 0)
	err = binary.Read(file, binary.LittleEndian, &ebr)
	if err != nil {
		consola.AddToConsole(fmt.Sprintf("no se pudo obtener la informacion del archivo para obtener el MBR %s\n", err.Error()))
		return ebr
	}
	return ebr
}

func MkDirectory(fullPath string) {
	directory := path.Dir(fullPath)
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		err = os.MkdirAll(directory, 0777)
		if err != nil {
			consola.AddToConsole(fmt.Sprintf("no se pudo crear el directorio %s\n", err.Error()))
		}
	}
}

func GetRandom() int64 {
	rand.Seed(time.Now().UnixNano())
	n := 150
	randomNumber := rand.Intn(n)
	return int64(randomNumber)
}

// funcion general para escribir SuperBloque, TablaInodo, BloqueDeArchivos, BloqueDeCarpetas
func Fwrite(estructura interface{}, path string, position int64) {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		consola.AddToConsole(fmt.Sprintf("no se pudo abrir el archivo para escribir la estructura %s\n", err.Error()))
		return
	}
	// Posicionandonos en donde necesitamos dentro del archivo
	_, err = file.Seek(position, 0)
	if err != nil {
		consola.AddToConsole(fmt.Sprintf("no se pudo posicionar en donde se desea: %d, %s\n", position, err.Error()))
		return
	}
	// Escribiendo la estructura
	err = binary.Write(file, binary.LittleEndian, estructura)
	if err != nil {
		consola.AddToConsole(fmt.Sprintf("no se pudo escribir la estructura %s\n", err.Error()))
		file.Close()
		return
	}
	// consola.AddToConsole(fmt.Sprintf("se escribio correctamente! :D")
	defer file.Close()
}
// WriteFile escribe contenido en un archivo dentro del sistema de archivos
func WriteFile(path string, superbloque *datos.SuperBloque, inodo *datos.TablaInodo, contenido string) bool {
    // Convertir contenido a bytes
    contenidoBytes := []byte(contenido)
    sizeContenido := len(contenidoBytes)
    
    // Verificar si el contenido cabe en los bloques actuales
    blocksNeeded := (sizeContenido + int(superbloque.S_block_size) - 1) / int(superbloque.S_block_size)
    
    // Verificar bloques disponibles
    if blocksNeeded > 15 { // Límite de bloques directos
        consola.AddToConsole("Error: El archivo excede el tamaño máximo permitido\n")
        return false
    }
    
    // Escribir en los bloques existentes o asignar nuevos si es necesario
    offset := 0
    for i := 0; i < blocksNeeded; i++ {
        var blockPos int64
        
        // Si el bloque no está asignado, buscar uno libre
        if i < len(inodo.I_block) && inodo.I_block[i] == -1 {
            // Buscar bloque libre y asignarlo
            blockPos = FindFreeBlock(path, superbloque)
            if blockPos == -1 {
                consola.AddToConsole("Error: No hay bloques disponibles\n")
                return false
            }
            inodo.I_block[i] = blockPos
        } else if i >= len(inodo.I_block) {
            consola.AddToConsole("Error: No hay suficientes punteros de bloque\n")
            return false
        } else {
            blockPos = inodo.I_block[i]
        }
        
        // Calcular cuánto escribir en este bloque
        writeSize := int(superbloque.S_block_size)
        if offset+writeSize > sizeContenido {
            writeSize = sizeContenido - offset
        }
        
        // Escribir el bloque
        if !WriteBlock(path, blockPos, contenidoBytes[offset:offset+writeSize]) {
            return false
        }
        offset += writeSize
    }
    
    // Actualizar tamaño del inodo
    inodo.I_size = int64(sizeContenido)
    
    // Actualizar tiempo de modificación
    mtime := time.Now()
    copy(inodo.I_mtime[:], mtime.String())
    
    return true
}

// WriteBlock escribe datos en un bloque específico
func WriteBlock(path string, blockPos int64, data []byte) bool {
    file, err := os.OpenFile(path, os.O_WRONLY, 0644)
    if err != nil {
        consola.AddToConsole(fmt.Sprintf("Error al abrir archivo: %s\n", err.Error()))
        return false
    }
    defer file.Close()
    
    _, err = file.Seek(blockPos, 0)
    if err != nil {
        consola.AddToConsole(fmt.Sprintf("Error al posicionar en bloque: %s\n", err.Error()))
        return false
    }
    
    err = binary.Write(file, binary.LittleEndian, data)
    if err != nil {
        consola.AddToConsole(fmt.Sprintf("Error al escribir bloque: %s\n", err.Error()))
        return false
    }
    
    return true
}

// FindFreeBlock encuentra el primer bloque libre en el bitmap
func FindFreeBlock(path string, superbloque *datos.SuperBloque) int64 {
    // Leer bitmap de bloques
    bitmap := make([]byte, superbloque.S_blocks_count)
    Fread(&bitmap, path, superbloque.S_bm_block_start)
    
    // Buscar primer bloque libre (bit 0)
    for i := 0; i < int(superbloque.S_blocks_count); i++ {
        if bitmap[i] == 0 {
            // Marcar como ocupado
            bitmap[i] = 1
            Fwrite(&bitmap, path, superbloque.S_bm_block_start)
            return superbloque.S_block_start + int64(i)*int64(superbloque.S_block_size)
        }
    }
    
    return -1 // No hay bloques libres
}
func Fread(estructura interface{}, path string, position int64) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		consola.AddToConsole(fmt.Sprintf("no se pudo abrir el archivo para escribir la estructura %s\n", err.Error()))
		return
	}
	// Posicionandonos en donde necesitamos dentro del archivo
	_, err = file.Seek(position, 0)
	if err != nil {
		consola.AddToConsole(fmt.Sprintf("no se pudo posicionar en donde se desea: %d, %s\n", position, err.Error()))
		return
	}
	// Leyendo La estructura
	err = binary.Read(file, binary.LittleEndian, estructura)
	if err != nil {
		consola.AddToConsole(fmt.Sprintf("no se pudo leer la estructura, %s, %s, %s\n", reflect.TypeOf(estructura).String(), ":", err.Error()))
		return
	}
	defer file.Close()
}

func Fopen(path, contenido string) {
	file, err := os.Create(path)
	if err != nil {
		consola.AddToConsole(fmt.Sprintf("no se pudo crear el archivo: %s, %s\n", path, err.Error()))
		return
	}
	defer file.Close()

	_, err = file.Write([]byte(contenido))
	if err != nil {
		consola.AddToConsole(fmt.Sprintf("no se pudo escribir al archivo: %s, %s\n", path, err.Error()))
	}

	consola.AddToConsole(fmt.Sprintf("archivo creado con exito: %s\n", path))
}

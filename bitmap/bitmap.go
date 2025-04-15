package bitmap

import (
	"encoding/binary"
	"fmt"
	"os"
	"unsafe"

	"github.com/melgxrga/proyecto1Archivos/commands"
	"github.com/melgxrga/proyecto1Archivos/structures"
)
func WriteSuperbloque(path string, superbloque *datos.SuperBloque) {
    file, err := os.OpenFile(path, os.O_RDWR, 0644)
    if err != nil {
        fmt.Println("Error al abrir el archivo para escribir el superbloque:", err)
        return
    }
    defer file.Close()

    // Escribir el SuperBloque al inicio del archivo (offset 0)
    file.Seek(0, 0)

    err = binary.Write(file, binary.LittleEndian, superbloque)
    if err != nil {
        fmt.Println("Error al escribir el superbloque:", err)
    } else {
        fmt.Println("SuperBloque escrito correctamente.")
    }
}


func WriteInBitmapInode(path string, superbloque *datos.SuperBloque) int64 {
    valor := byte('1')
    position := superbloque.S_first_ino

    if position == -1 {
        fmt.Println("No se encontró posición vacía en el bitmap de inodos")
        return -1
    }

    comandos.Fwrite(&valor, path, superbloque.S_bm_inode_start+position*int64(unsafe.Sizeof(valor)))
    superbloque.S_first_ino = SearchFirstFreeBitmapInodePos(path, superbloque)
    superbloque.S_free_inodes_count--

    // Escribir los cambios en el archivo
    WriteSuperbloque(path, superbloque)

    fmt.Println("Depuración: S_first_ino actualizado a", superbloque.S_first_ino)
    return position
}


func WriteInBitmapBlock(path string, superbloque *datos.SuperBloque) int64 {
    valor := byte('1')
    position := superbloque.S_first_blo

    if position == -1 {
        fmt.Println("No se encontró posición vacía en el bitmap de bloques")
        return -1
    }

    comandos.Fwrite(&valor, path, superbloque.S_bm_block_start+position*int64(unsafe.Sizeof(valor)))
    superbloque.S_first_blo = SearchFirstFreeBitmapBlockPos(path, superbloque)
    superbloque.S_free_blocks_count--

    // Escribir los cambios en el archivo
    WriteSuperbloque(path, superbloque)

    fmt.Println("Depuración: S_first_blo actualizado a", superbloque.S_first_blo)
    return position
}


func DeleteBitmapInode(path string, superbloque *datos.SuperBloque, posicion int64) {
    valor := byte('0')
    file, err := os.OpenFile(path, os.O_RDWR, 0644) 
    if err != nil {
        fmt.Println("Error al abrir el archivo:", err)
        return
    }
    defer file.Close() 

    file.Seek(superbloque.S_bm_inode_start+(posicion*int64(unsafe.Sizeof(valor))), 0)
    FwriteByte(file, &valor)
    superbloque.S_first_ino = SearchFirstFreeBitmapInodePos(path, superbloque)
    superbloque.S_free_inodes_count++

    // Escribir los cambios en el archivo
    WriteSuperbloque(path, superbloque)

    fmt.Println("Depuración: S_first_ino actualizado a", superbloque.S_first_ino)
}


func DeleteBitmapBlock(path string, superbloque *datos.SuperBloque, posicion int64) {
    valor := byte('0')
    file, err := os.OpenFile(path, os.O_RDWR, 0644) 
    if err != nil {
        fmt.Println("Error al abrir el archivo:", err)
        return
    }
    defer file.Close() 

    file.Seek(superbloque.S_bm_block_start+(posicion*int64(unsafe.Sizeof(valor))), 0)
    FwriteByte(file, &valor)
    superbloque.S_first_blo = SearchFirstFreeBitmapBlockPos(path, superbloque)
    superbloque.S_free_blocks_count++

    // Escribir los cambios en el archivo
    WriteSuperbloque(path, superbloque)

    fmt.Println("Depuración: S_first_blo actualizado a", superbloque.S_first_blo)
}


// buscar primer bit libre en los bitmaps

func SearchFirstFreeBitmapInodePos(path string, superbloque *datos.SuperBloque) int64 {
	contar := 0
	for contar < int(superbloque.S_inodes_count) {
		i := byte('0')
		comandos.Fread(&i, path, superbloque.S_bm_inode_start+int64(contar)*int64(unsafe.Sizeof(i)))
		// fmt.Println("byte en bitmap de inodo", i)
		if i == '0' {
			return int64(contar)
		}
		contar++
	}
	return -1
}

func SearchFirstFreeBitmapBlockPos(path string, superbloque *datos.SuperBloque) int64 {
	contar := 0
	for contar < int(superbloque.S_blocks_count) {
		i := byte('0')
		comandos.Fread(&i, path, superbloque.S_bm_block_start+int64(contar)*int64(unsafe.Sizeof(i)))
		// fmt.Println("byte en bitmap de bloque", i)
		if i == '0' {
			return int64(contar)
		}
		contar++
	}
	return -1
}

// leer un byte en archivo

func FreadByte(file *os.File, temp *byte) {
	err := binary.Read(file, binary.LittleEndian, temp)
	if err != nil {
		fmt.Println("no se pudo leer,", err.Error())
	}
}

// escribir un byte en archivo

func FwriteByte(file *os.File, temp *byte) {
	err := binary.Write(file, binary.LittleEndian, temp)
	if err != nil {
		fmt.Println("no se pudo escribir,", err.Error())
	}
}
func OpenNewFile(path string) *os.File {
    // Verificar si el directorio tiene permisos de escritura
    fileInfo, err := os.Stat(path)
    if err != nil {
        fmt.Println("Error al obtener información del archivo:", err.Error())
    } else {
        fmt.Println("Permisos del directorio:", fileInfo.Mode())
    }

    file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
    if err != nil {
        fmt.Println("Error al abrir el archivo para Bitmap:", err.Error())
        return nil
    }
    fmt.Println("Archivo abierto correctamente:", file.Name())

    return file
}

func S_bm_inode_print(file *os.File, superbloque *datos.SuperBloque) {
	contador := 0
	bit := byte('2')
	for contador < int(superbloque.S_inodes_count) {
		FreadByte(file, &bit)
		fmt.Println(bit)
		contador++
	}
}

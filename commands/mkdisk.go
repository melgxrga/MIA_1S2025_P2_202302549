package comandos

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"regexp"      
	"github.com/melgxrga/proyecto1Archivos/consola"
	"github.com/melgxrga/proyecto1Archivos/structures"
)

type ParametrosMkdisk struct {
	Size int
	Fit  byte
	Unit byte
	Path string
}

type Mkdisk struct {
	Params ParametrosMkdisk
}

func (m *Mkdisk) Exe(parametros []string) {
	m.Params = m.SaveParams(parametros)
	if m.Mkdisk(m.Params.Size, m.Params.Fit, m.Params.Unit, m.Params.Path) {
		consola.AddToConsole(fmt.Sprintf("\nmkdisk realizado con exito para la ruta raiz: %s\n\n", m.Params.Path))
	} else {
		consola.AddToConsole(fmt.Sprintf("\n[ERROR!] no se logro realizar el comando mkdisk para la ruta raiz: %s\n\n", m.Params.Path))
	}
}

func (m *Mkdisk) SaveParams(parametros []string) ParametrosMkdisk {
	var params ParametrosMkdisk
	// Unir todos los parámetros en una sola cadena
	args := strings.Join(parametros, " ")

	// Expresión regular para capturar los parámetros
	re := regexp.MustCompile(`-size=\d+|-unit=[kKmM]|-fit=[bBfFwW]{2}|-path="[^"]+"|-path=[^\s]+`)
	matches := re.FindAllString(args, -1)

	// Iterar sobre cada coincidencia
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
			if value != "K" && value != "M" {
				fmt.Println("Error: la unidad debe ser K o M")
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
		default:
			fmt.Printf("Parámetro desconocido: %s\n", key)
		}
	}

	// Validación final de los parámetros obligatorios
	if params.Size == 0 {
		fmt.Println("Error: Falta el parámetro obligatorio -size")
	}
	if params.Path == "" {
		fmt.Println("Error: Falta el parámetro obligatorio -path")
	}

	// Valores por defecto si no se proporcionaron
	if params.Unit == 0 {
		params.Unit = 'M'
	}
	if params.Fit == 0 {
		params.Fit = 'F'
	}

	return params
}


func (m *Mkdisk) Mkdisk(size int, fit byte, unit byte, path string) bool {
	var fileSize = 0
	var master datos.MBR
	// Comprobando si existe una ruta valida para la creacion del disco
	if path == "" {
		consola.AddToConsole("no se encontro una ruta\n")
		return false
	}
	// comprobando el tamano del disco, debe ser mayor que cero
	if size <= 0 {
		consola.AddToConsole("el tamano del disco debe ser mayor que 0\n")
		return false
	}
	// tipo de unidad a utilizar, si el parametro esta vacio se utilizaran MegaBytes como default size
	if unit == 'k' || unit == 'K' {
		fileSize = size
	} else if unit == 'm' || unit == 'M' {
		fileSize = size * 1024
	} else if unit == 0 {
		fileSize = size * 1024
	} else {
		consola.AddToConsole("se debe ingresar una letra que corresponda un tamano valido\n")
		return false
	}
	// definiendo el tipo de fit que el disco tendra, como default sera First Fit
	//fmt.Printf("tipo de la variable fit %T\n", fit)
	//fmt.Println("el fit es:", fit)
	if strconv.Itoa(int(fit)) == "66" || string(fit) == "BF" {
		master.Dsk_fit = 'b'
	} else if strconv.Itoa(int(fit)) == "70" || string(fit) == "FF" {
		master.Dsk_fit = 'f'
	} else if strconv.Itoa(int(fit)) == "87" || string(fit) == "WF" {
		master.Dsk_fit = 'w'
	} else if fit == 0 {
		master.Dsk_fit = 'f'
	} else {
		consola.AddToConsole("se debe ingresar un tipo de fit valido\n")
		return false
	}
	// llenando el buffer con '0' para indicar que esta vacio.
	bloque := make([]byte, 1024)
	for i := 0; i < len(bloque); i++ {
		bloque[i] = 0
	}

	iterator := 0
	MkDirectory(path) // creando el directorio para el disco sino existe
	binaryFile, err := os.Create(path)
	if err != nil {
		consola.AddToConsole("error al crear el disco\n")
		return false
	}
	defer binaryFile.Close()
	for iterator < fileSize {
		_, err := binaryFile.Write(bloque[:])
		if err != nil {
			consola.AddToConsole("error al llenar el disco creado\n")
		}
		iterator++
	}
	master.Mbr_tamano = int64(fileSize * 1024)
	master.Mbr_dsk_signature = GetRandom()
	// formateando el tiempo
	date := time.Now()
	for i := 0; i < len(master.Mbr_fecha_creacion)-1; i++ {
		master.Mbr_fecha_creacion[i] = date.String()[i]
	}
	FillPartitions(&master)
	WriteMBR(&master, path)
	return true
}

func FillPartitions(master *datos.MBR) {
	for i := 0; i < len(master.Mbr_partitions); i++ {
		master.Mbr_partitions[i].Part_status = '0'
		master.Mbr_partitions[i].Part_fit = '0'
		master.Mbr_partitions[i].Part_start = 0
		master.Mbr_partitions[i].Part_size = 0
		master.Mbr_partitions[i].Part_type = '0'
		copy(master.Mbr_partitions[i].Part_name[:], "")
	}
}

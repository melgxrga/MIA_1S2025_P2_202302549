package comandos

import (
	"bytes"
	"fmt"
	"strings"
	"regexp"  // Paquete para trabajar con expresiones regulares, útil para encontrar y manipular patrones en cadenas
	"github.com/melgxrga/proyecto1Archivos/consola"
	"github.com/melgxrga/proyecto1Archivos/functions"
	"github.com/melgxrga/proyecto1Archivos/list"
)

type ParametrosMount struct {
	Path string
	Name [16]byte
}

type Mount struct {
	Params ParametrosMount
}

func (m *Mount) Exe(parametros []string) {
	m.Params = m.SaveParams(parametros)
	if m.Mount(m.Params.Path, m.Params.Name) {
		consola.AddToConsole(fmt.Sprintf("\nparticion %s montada con exito\n\n", m.Params.Path))
	} else {
		consola.AddToConsole(fmt.Sprintf("no se logro montar la particion %s\n", m.Params.Path))
	}
}

func (m *Mount) SaveParams(parametros []string) ParametrosMount {
	var params ParametrosMount
	// Unir todos los parámetros en una sola cadena
	args := strings.Join(parametros, " ")

	// Expresión regular para capturar los parámetros
	re := regexp.MustCompile(`-path="[^"]+"|-path=[^\s]+|-name="[^"]+"|-name=[^\s]+`)
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

		// Procesar según el parámetro encontrado
		switch key {
		case "-path":
			if value == "" {
				fmt.Println("Error: el path no puede estar vacío")
				continue
			}
			params.Path = value
		case "-name":
			if value == "" {
				fmt.Println("Error: el nombre no puede estar vacío")
				continue
			}
			copy(params.Name[:], value)
		default:
			fmt.Printf("Parámetro desconocido: %s\n", key)
		}
	}

	// Validación final de los parámetros obligatorios
	if params.Path == "" {
		fmt.Println("Error: Falta el parámetro obligatorio -path")
	}
	if params.Name == [16]byte{} {
		fmt.Println("Error: Falta el parámetro obligatorio -name")
	}

	return params
}


func (m *Mount) Mount(path string, name [16]byte) bool {
    // Validaciones básicas
    if path == "" {
        consola.AddToConsole("Error: No se encontró una ruta válida\n")
        return false
    }
    if bytes.Equal(name[:], []byte("")) {
        consola.AddToConsole("Error: Se debe proporcionar un nombre de partición\n")
        return false
    }

    // Verificar si ya está montada
    if lista.ListaMount.IsMounted(path, name) {
        consola.AddToConsole(fmt.Sprintf("Error: La partición %s ya está montada\n", 
            string(functions.TrimArray(name[:]))))
        return false
    }

    master := GetMBR(path)
    partitionMounted := false
    particionEncontrada := false
    
    // Buscar particiones primarias/extendidas
    for _, particion := range master.Mbr_partitions {
        if bytes.Equal(particion.Part_name[:], name[:]) {
            particionEncontrada = true
            
            if particion.Part_type == 'e' || particion.Part_type == 'E' {
                consola.AddToConsole("Error: No se puede montar una partición extendida directamente\n")
                return false
            }
            
            // Marcar como montada y agregar a la lista
            particion.Part_status = '2'
            lista.ListaMount.Mount(path, 49, &particion, nil)
            partitionMounted = true
            lista.ListaMount.PrintList()
            break
        }
    }

    // Si no se encontró como primaria, buscar como lógica
    if !particionEncontrada {
        for _, particion := range master.Mbr_partitions {
            if particion.Part_type == 'e' || particion.Part_type == 'E' {
                if m.MountParticionLogica(path, int(particion.Part_start), name) {
                    partitionMounted = true
                    lista.ListaMount.PrintList()
                }
            }
        }
    }

    if !partitionMounted {
		consola.AddToConsole(fmt.Sprintf("Error: No se encontró la partición %s en %s\n", 
		string(functions.TrimArray(name[:])), path))
        return false
    }
    
    WriteMBR(&master, path)
    return true
}

func (m *Mount) MountParticionLogica(path string, whereToStart int, name [16]byte) bool {
    // Verificar si ya está montada
    if lista.ListaMount.IsMounted(path, name) {
        consola.AddToConsole(fmt.Sprintf("Error: La partición lógica %s ya está montada\n", 
            string(functions.TrimArray(name[:]))))
        return false
    }

    temp := ReadEBR(path, int64(whereToStart))
    
    for {
        if bytes.Equal(temp.Part_name[:], name[:]) {
            temp.Part_status = '2'
            lista.ListaMount.Mount(path, 49, nil, &temp)
            return true
        }
        
        if temp.Part_next == -1 {
            break
        }
        temp = ReadEBR(path, temp.Part_next)
    }
    
    consola.AddToConsole(fmt.Sprintf("Error: No se encontró la partición lógica %s\n", 
        string(functions.TrimArray(name[:]))))
    return false
}
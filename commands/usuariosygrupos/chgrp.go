package usuariosygrupos

import (
	"fmt"
	"strings"
	"regexp"
	"github.com/melgxrga/proyecto1Archivos/consola"
	"github.com/melgxrga/proyecto1Archivos/structures"
	"github.com/melgxrga/proyecto1Archivos/list"
	"github.com/melgxrga/proyecto1Archivos/logger"
	commands "github.com/melgxrga/proyecto1Archivos/commands"
	"unsafe"
)

type ParametrosChgrp struct {
	User string
	Grp  string
}

type Chgrp struct {
	Params ParametrosChgrp
}

func (c *Chgrp) Exe(parametros []string) {
	c.Params = c.SaveParams(parametros)
	if c.Chgrp(c.Params.User, c.Params.Grp) {
		consola.AddToConsole(fmt.Sprintf("\nGrupo del usuario %s cambiado a %s con éxito\n\n", c.Params.User, c.Params.Grp))
	} else {
		consola.AddToConsole(fmt.Sprintf("No se logró cambiar el grupo del usuario %s\n\n", c.Params.User))
	}
}

func (c *Chgrp) SaveParams(parametros []string) ParametrosChgrp {
	var params ParametrosChgrp
	args := strings.Join(parametros, " ")
	re := regexp.MustCompile(`-user="[^"]+"|-user=[^\s]+|-grp="[^"]+"|-grp=[^\s]+`)
	matches := re.FindAllString(args, -1)
	
	for _, match := range matches {
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			fmt.Printf("Formato de parámetro inválido: %s\n", match)
			continue
		}
		key, value := strings.ToLower(kv[0]), kv[1]
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}
		switch key {
		case "-user":
			if value == "" {
				fmt.Println("Error: el usuario no puede estar vacío")
				continue
			}
			params.User = value
		case "-grp":
			if value == "" {
				fmt.Println("Error: el grupo no puede estar vacío")
				continue
			}
			params.Grp = value
		}
	}

	if params.User == "" {
		fmt.Println("Error: Falta el parámetro obligatorio -user")
	}
	if params.Grp == "" {
		fmt.Println("Error: Falta el parámetro obligatorio -grp")
	}

	return params
}

func (c *Chgrp) Chgrp(user string, grp string) bool {
    // Verificar permisos de root
    if !logger.Log.IsLoggedIn() || !logger.Log.UserIsRoot() {
        consola.AddToConsole("Error: Solo el usuario root puede ejecutar este comando\n")
        return false
    }

    // Obtener partición montada
    mountNode := lista.ListaMount.GetNodeById(logger.Log.GetUserId())
    if mountNode == nil {
        consola.AddToConsole("Error: No se encontró la partición montada\n")
        return false
    }

    // Determinar posición de inicio
    var startPos int64
    if mountNode.Value != nil {
        startPos = mountNode.Value.Part_start
    } else if mountNode.ValueL != nil {
        startPos = mountNode.ValueL.Part_start + int64(unsafe.Sizeof(datos.EBR{}))
    } else {
        consola.AddToConsole("Error: No se pudo determinar la posición de la partición\n")
        return false
    }

    // Leer superbloque
    var superbloque datos.SuperBloque
    commands.Fread(&superbloque, mountNode.Ruta, startPos)

    // Leer inodo de users.txt
    var tablaInodo datos.TablaInodo
    commands.Fread(&tablaInodo, mountNode.Ruta, superbloque.S_inode_start+int64(unsafe.Sizeof(datos.TablaInodo{})))

    // Leer contenido de users.txt
    contenido := ReadFile(&tablaInodo, mountNode.Ruta, &superbloque)
	// Validar existencia de usuario y grupo
    if !c.ExisteUsuario(contenido, user) {
        consola.AddToConsole(fmt.Sprintf("Error: El usuario %s no existe\n", user))
        return false
    }
    if !c.ExisteGrupo(contenido, grp) {
        consola.AddToConsole(fmt.Sprintf("Error: El grupo %s no existe\n", grp))
        return false
    }

    // Actualizar grupo del usuario
    nuevoContenido := c.ActualizarGrupoUsuario(contenido, user, grp)
    if nuevoContenido == "" {
        consola.AddToConsole("Error: No se pudo actualizar el grupo del usuario\n")
        return false
    }

    // Mostrar contenido modificado
    consola.AddToConsole("\nContenido modificado de users.txt:\n")
    consola.AddToConsole("-------------------------------\n")
    consola.AddToConsole(nuevoContenido)
    consola.AddToConsole("-------------------------------\n\n")

    // Escribir cambios
    if commands.WriteFile(mountNode.Ruta, &superbloque, &tablaInodo, nuevoContenido) {
        commands.Fwrite(&tablaInodo, mountNode.Ruta, superbloque.S_inode_start+int64(unsafe.Sizeof(datos.TablaInodo{})))
        commands.Fwrite(&superbloque, mountNode.Ruta, startPos)
        
        // Mostrar confirmación
        consola.AddToConsole(fmt.Sprintf("\nÉxito: Se cambió el grupo del usuario '%s' a '%s'\n", user, grp))
        return true
    }

    consola.AddToConsole("Error: No se pudo escribir los cambios en el disco\n")
    return false
}

func (c *Chgrp) ExisteUsuario(contenido string, userName string) bool {
	lineas := strings.Split(contenido, "\n")
	for _, linea := range lineas {
		linea = strings.ReplaceAll(linea, "\x00", "")
		parametros := strings.Split(linea, ",")
		if len(parametros) < 4 || parametros[1] != "U" {
			continue
		}
		if parametros[3] == userName {
			return true
		}
	}
	return false
}

func (c *Chgrp) ExisteGrupo(contenido string, groupName string) bool {
	lineas := strings.Split(contenido, "\n")
	for _, linea := range lineas {
		linea = strings.ReplaceAll(linea, "\x00", "")
		parametros := strings.Split(linea, ",")
		if len(parametros) < 3 || parametros[1] != "G" {
			continue
		}
		if parametros[2] == groupName {
			return true
		}
	}
	return false
}

func (c *Chgrp) ActualizarGrupoUsuario(contenido string, user string, newGroup string) string {
	lineas := strings.Split(contenido, "\n")
	var nuevasLineas []string
	cambioRealizado := false

	for _, linea := range lineas {
		linea = strings.ReplaceAll(linea, "\x00", "")
		if linea == "" {
			continue
		}

		parametros := strings.Split(linea, ",")
		if len(parametros) >= 4 && parametros[1] == "U" && parametros[3] == user {
			// Actualizar grupo del usuario
			parametros[2] = newGroup
			linea = strings.Join(parametros, ",")
			cambioRealizado = true
		}
		nuevasLineas = append(nuevasLineas, linea)
	}

	if !cambioRealizado {
		consola.AddToConsole("Error: No se pudo actualizar el grupo del usuario\n")
		return ""
	}

	return strings.Join(nuevasLineas, "\n") + "\n"
}
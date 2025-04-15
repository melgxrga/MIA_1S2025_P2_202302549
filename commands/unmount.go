package comandos

import (
	"fmt"
	"strings"
	"github.com/melgxrga/proyecto1Archivos/consola"
	"github.com/melgxrga/proyecto1Archivos/list"
	"regexp"
)

type ParametrosUnmount struct {
	ID string
}

type Unmount struct {
	Params ParametrosUnmount
}

func (u *Unmount) Exe(parametros []string) {
	u.Params = u.SaveParams(parametros)
	if u.Unmount(u.Params.ID) {
		consola.AddToConsole(fmt.Sprintf("\nparticion con id '%s' desmontada con exito\n\n", u.Params.ID))
	} else {
		consola.AddToConsole(fmt.Sprintf("No se logro desmontar la particion con id '%s'\n", u.Params.ID))
	}
}

func (u *Unmount) SaveParams(parametros []string) ParametrosUnmount {
	var params ParametrosUnmount
	args := strings.Join(parametros, " ")
	re := regexp.MustCompile(`-id=[^\s]+`)
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
			params.ID = value
		}
	}
	// Validación final de los parámetros obligatorios
	if params.ID == "" {
		consola.AddToConsole("Error: el parámetro obligatorio -id no fue proporcionado\n")
	}
	return params
}

func (m *Unmount) Unmount(id string) bool {
	lista.ListaMount.PrintList()
	if id == "" {
		consola.AddToConsole("se debe contar con id de la particion\n")
		return false
	}
	consola.AddToConsole(fmt.Sprintf("Intentando desmontar partición con id: '%s'\n", id))
	unmountptr := lista.ListaMount.UnMount(id)
	if unmountptr == nil {
		consola.AddToConsole(fmt.Sprintf("No existe una partición montada con el ID '%s'\n", string(id)))
		return false
	}
	consola.AddToConsole(fmt.Sprintf("Partición con id '%s' desmontada correctamente.\n", id))
	lista.ListaMount.PrintList()
	return true
}

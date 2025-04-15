package comandos

import (
	"fmt"
	"strings"
	"github.com/melgxrga/proyecto1Archivos/consola"
	"github.com/melgxrga/proyecto1Archivos/list"
)

type ParametrosUnmount struct {
	Path string
	ID   string
}

type Unmount struct {
	Params ParametrosUnmount
}

func (u *Unmount) Exe(parametros []string) {
	u.Params = u.SaveParams(parametros)
	if u.Unmount(u.Params.Path, u.Params.ID) {
		consola.AddToConsole(fmt.Sprintf("\nparticion con id '%s' desmontada con exito\n\n", u.Params.ID))
	} else {
		consola.AddToConsole(fmt.Sprintf("No se logro desmontar la particion con id '%s'\n", u.Params.ID))
	}
}

func (u *Unmount) SaveParams(parametros []string) ParametrosUnmount {
	// fmt.Println(parametros)
	for _, v := range parametros {
		// fmt.Println(v)
		v = strings.TrimSpace(v)
		v = strings.TrimRight(v, " ")
		if strings.Contains(v, "driveletter") {
			v = strings.ReplaceAll(v, "driveletter=", "")
			v = strings.ReplaceAll(v, "\"", "")
			v = v + ".dsk"
			u.Params.Path = v
		} else if strings.Contains(v, "id") {
			v = strings.ReplaceAll(v, "id=", "")
			u.Params.ID = v
		}
	}
	return u.Params
}

func (m *Unmount) Unmount(path string, id string) bool {
	lista.ListaMount.PrintList()
	if path == "" {
		consola.AddToConsole("se debe contar con el nombre del disco\n")
		return false
	}
	if id == "" {
		consola.AddToConsole("se debe contar con id de la particion\n")
		return false
	}
	unmountptr := lista.ListaMount.UnMount(id)
	if unmountptr == nil {
		consola.AddToConsole(fmt.Sprintf("No existe una partición montada con el ID '%s'\n", string(id)))
		return false
	}
	lista.ListaMount.PrintList()
	return true
}

package usuariosygrupos

import (
	"github.com/melgxrga/proyecto1Archivos/consola"
	datos "github.com/melgxrga/proyecto1Archivos/structures"

)


type Cat struct{}

func (c *Cat) Exe(path string, superbloque *datos.SuperBloque) {
	var tablaInodo datos.TablaInodo
    contenido :=ReadFile(&tablaInodo, path, superbloque)
    if contenido == "" {
        consola.AddToConsole("El archivo está vacío o no existe\n")
    } else {
        consola.AddToConsole(contenido + "\n")
    }
}
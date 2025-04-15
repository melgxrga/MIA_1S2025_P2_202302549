package consola

import "fmt"

var content string

func AddToConsole(nuevoContenido string) {
	content += nuevoContenido + "\n" // Acumula los mensajes
	fmt.Print(nuevoContenido)
}

func GetConsole() string {
	returnable := content
	content = "" // Limpia después de recuperar
	return returnable
}
func Nothing(contenido string) {
	content = "" // Limpia después de recuperar
}
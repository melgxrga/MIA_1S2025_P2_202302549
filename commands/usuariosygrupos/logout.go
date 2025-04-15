package usuariosygrupos

import (
	"github.com/melgxrga/proyecto1Archivos/consola"
	"github.com/melgxrga/proyecto1Archivos/logger"
)

type Logout struct {
}

func (l *Logout) Exe(parametros []string) {
	if l.Logout() {
		consola.AddToConsole("\nse ha cerrado la sesion\n\n")
	} else {
		consola.AddToConsole("no se logro hacer logout\n\n")
	}
}

func (l *Logout) Logout() bool {
	return logger.Log.Logout()
}

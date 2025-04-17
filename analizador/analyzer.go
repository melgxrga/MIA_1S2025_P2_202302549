package analizador

import (
	"bufio"

	"os"
	"strings"
	"github.com/melgxrga/proyecto1Archivos/commands"
	"github.com/melgxrga/proyecto1Archivos/commands/usuariosygrupos"
	"github.com/melgxrga/proyecto1Archivos/consola"
)

type Analyzer struct {}

func (a *Analyzer) Analyzer(input string) (interface{}, error) {
	// Dividir la entrada en múltiples comandos
	commands := strings.Split(input, "\n") 

	for _, cmd := range commands {
		cmd = strings.TrimSpace(cmd)
		if cmd == "" {
			continue // Saltar líneas vacías
		}

		tokens := strings.Fields(cmd)
		if len(tokens) == 0 {
			consola.AddToConsole("No se proporcionó ningún comando\n")
			continue
		}

		command := tokens[0]
		params := tokens[1:]

		switch command {
		case "pause":
			consola.AddToConsole("\nPresione 'ENTER' para continuar: ")
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
		
		case "mkdisk":
			m := comandos.Mkdisk{}
			m.Exe(params)

		case "fdisk":
			f := comandos.Fdisk{}
			f.Exe(params)

		case "mount":
			m := comandos.Mount{}
			m.Exe(params)
		
		case "rmdisk":
			r := comandos.Rmdisk{}
			r.Exe(params)

		case "mkfs":
			m := comandos.Mkfs{}
			m.Exe(params)

		case "login":
			l := usuariosygrupos.Login{}
			l.Exe(params)

		case "logout":
			l := usuariosygrupos.Logout{}
			l.Exe(params)

		case "rep":
			r := usuariosygrupos.Rep{}
			r.Exe(params)
		case "mkgrp":
			m := usuariosygrupos.Mkgrp{}
			m.Exe(params)
		case "rmgrp":
			r := usuariosygrupos.Rmgrp{}
			r.Exe(params)
		case "mkusr":
			m := usuariosygrupos.Mkusr{}
			m.Exe(params)
		case "rmusr":
			r := usuariosygrupos.Rmusr{}
			r.Exe(params)
		case "mkdir":
			m := usuariosygrupos.Mkdir{}
			m.Exe(params)
		case "mkfile":
			m := usuariosygrupos.Mkfile{}
			m.Exe(params)
		case "chgrp":
			c:= usuariosygrupos.Chgrp{}
			c.Exe(params)
		case "cat":
			continue;
		case "unmount":
			u := comandos.Unmount{}
			u.Exe(params)
		case "remove":
			r := usuariosygrupos.Remove{}
			r.Exe(params)
			
		case "edit":
			e := usuariosygrupos.Edit{}
			e.Exe(params)
			
		case "rename":
			r := usuariosygrupos.Rename{}
			r.Exe(params)
			
		case "copy":
			c := usuariosygrupos.Copy{}
			c.Exe(params)
			
		default:
			command = strings.TrimSpace(command) // Limpiar espacios y saltos de línea
			if strings.HasPrefix(command, "#") {
				consola.AddToConsole(command + "\n") // El comentario ya viene completo
			} else {
				consola.AddToConsole("Comando desconocido: " + command + "\n")
			}
		}
	}
	return nil, nil
}


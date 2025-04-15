package usuariosygrupos

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unsafe"
	"regexp"
	"github.com/melgxrga/proyecto1Archivos/commands"
	"github.com/melgxrga/proyecto1Archivos/consola"
	"github.com/melgxrga/proyecto1Archivos/structures"
	"github.com/melgxrga/proyecto1Archivos/list"
	"github.com/melgxrga/proyecto1Archivos/logger"
)

type ParametrosMkusr struct {
	User string
	Pwd  string
	Grp  string
}

type Mkusr struct {
	Params ParametrosMkusr
}

func (m *Mkusr) Exe(parametros []string) {
	m.Params = m.SaveParams(parametros)
	if m.Mkusr(m.Params.User, m.Params.Pwd, m.Params.Grp) {
		consola.AddToConsole(fmt.Sprintf("\nusuario \"%s\" creado con exito en el grupo %s\n\n", m.Params.User, m.Params.Grp))
	} else {
		consola.AddToConsole(fmt.Sprintf("no se logro crear el usuario \"%s\"\n\n", m.Params.User))
	}
}

func (m *Mkusr) SaveParams(parametros []string) ParametrosMkusr {
	var params ParametrosMkusr
	args := strings.Join(parametros, " ")
	re := regexp.MustCompile(`-user="[^"]+"|-user=[^\s]+|-pass="[^"]+"|-pass=[^\s]+|-grp="[^"]+"|-grp=[^\s]+`)
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
		case "-pass":
			if value == "" {
				fmt.Println("Error: la contraseña no puede estar vacía")
				continue
			}
			params.Pwd = value
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
	if params.Pwd == "" {
		fmt.Println("Error: Falta el parámetro obligatorio -pass")
	}
	if params.Grp == "" {
		fmt.Println("Error: Falta el parámetro obligatorio -grp")
	}

	return params
}

func (m *Mkusr) Mkusr(user string, pwd string, grp string) bool {
	consola.AddToConsole(user + "\n")
	consola.AddToConsole(pwd + "\n")
	consola.AddToConsole(grp + "\n")
	if user == "" {
		consola.AddToConsole("no se encontro ningun nombre\n")
		return true
	}
	if logger.Log.IsLoggedIn() && logger.Log.UserIsRoot() {
		if lista.ListaMount.GetNodeById(logger.Log.GetUserId()).Value != nil {
			return m.MkusrPartition(user, pwd, grp, lista.ListaMount.GetNodeById(logger.Log.GetUserId()).Value.Part_start, lista.ListaMount.GetNodeById(logger.Log.GetUserId()).Ruta)
		} else if lista.ListaMount.GetNodeById(logger.Log.GetUserId()).Value != nil {
			return m.MkusrPartition(user, pwd, grp, lista.ListaMount.GetNodeById(logger.Log.GetUserId()).ValueL.Part_start+int64(unsafe.Sizeof(datos.EBR{})), lista.ListaMount.GetNodeById(logger.Log.GetUserId()).Ruta)
		}
	}
	return false
}

func (m *Mkusr) MkusrPartition(user string, pwd string, grp string, whereToStart int64, path string) bool {
	// superbloque de la particion
	var superbloque datos.SuperBloque
	comandos.Fread(&superbloque, path, whereToStart)

	// tabla de inodos de archivo Users.txt
	var tablaInodo datos.TablaInodo
	comandos.Fread(&tablaInodo, path, superbloque.S_inode_start+int64(unsafe.Sizeof(datos.TablaInodo{})))
	// modificar la fecha en la que se esta modificando el inodo
	mtime := time.Now()
	for i := 0; i < len(tablaInodo.I_mtime); i++ {
		tablaInodo.I_mtime[i] = mtime.String()[i]
	}
	if m.ExisteUsuario(ReadFile(&tablaInodo, path, &superbloque), user) {
		consola.AddToConsole(fmt.Sprintf("ya existe usuario con ese nombre %s\n", user))
		return false
	}
	if !m.ExisteGrupo(ReadFile(&tablaInodo, path, &superbloque), grp) {
		consola.AddToConsole(fmt.Sprintf("no existe un grupo con el nombre %s\n", grp))
		return false
	}
	numero := m.ContarUsuarios(ReadFile(&tablaInodo, path, &superbloque))
	usuario := m.AgregarUsuario(numero, grp, user, pwd)
	if AppendFile(path, &superbloque, &tablaInodo, usuario) {
		comandos.Fwrite(&tablaInodo, path, superbloque.S_inode_start+int64(unsafe.Sizeof(datos.TablaInodo{})))
		consola.AddToConsole(ReadFile(&tablaInodo, path, &superbloque))
		comandos.Fwrite(&superbloque, path, whereToStart)
		return true
	}
	return false
}

func (m *Mkusr) AgregarUsuario(userNumber int, groupName string, userName string, password string) string {
	return strconv.Itoa(userNumber) + ",U," + groupName + "," + userName + "," + password + "\n"
}

func (m *Mkusr) ContarUsuarios(contenido string) int {
	contador := 1
	lineas := strings.Split(contenido, "\n")
	lineas = lineas[:len(lineas)-1]
	for _, linea := range lineas {
		parametros := strings.Split(linea, ",")
		if parametros[1] != "U" {
			continue
		}
		contador++
	}
	return contador
}

func (m *Mkusr) ExisteUsuario(contenido string, userName string) bool {
	lineas := strings.Split(contenido, "\n")
	lineas = lineas[:len(lineas)-1]
	for _, linea := range lineas {
		linea = strings.ReplaceAll(linea, "\x00", "")
		parametros := strings.Split(linea, ",")
		if parametros[1] != "U" {
			continue
		}
		if parametros[3] == userName {
			return true
		}
	}
	return false
}
func (m *Mkusr) ExisteGrupo(contenido string, groupName string) bool {
	lineas := strings.Split(strings.TrimSpace(contenido), "\n") // Elimina espacios antes/después
	for _, linea := range lineas {
		limpia := strings.TrimSpace(strings.ReplaceAll(linea, "\x00", "")) // Elimina caracteres basura
		parametros := strings.Split(limpia, ",")

		if len(parametros) < 3 || parametros[1] != "G" {
			continue
		}

		fmt.Printf("Comparando grupo '%s' con '%s'\n", parametros[2], groupName)

		if strings.TrimSpace(parametros[2]) == strings.TrimSpace(groupName) {
			return true
		}
	}
	return false
}


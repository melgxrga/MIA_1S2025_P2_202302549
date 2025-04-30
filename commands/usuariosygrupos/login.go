package usuariosygrupos

import (
	"bytes"
	"fmt"
	"github.com/melgxrga/proyecto1Archivos/commands"
	"github.com/melgxrga/proyecto1Archivos/consola"
	"github.com/melgxrga/proyecto1Archivos/functions"
	"github.com/melgxrga/proyecto1Archivos/list"
	"github.com/melgxrga/proyecto1Archivos/logger"
	"github.com/melgxrga/proyecto1Archivos/structures"
	"regexp"
	"strings"
	"unsafe"
)

type ParametrosLogin struct {
	User [10]byte
	Pwd  [10]byte
	Id   string
}

type Login struct {
	Params ParametrosLogin
}

func (l *Login) Exe(parametros []string) {
	l.Params = l.SaveParams(parametros)
	if l.Login(l.Params.User, l.Params.Pwd, l.Params.Id) {
		consola.AddToConsole(fmt.Sprintf("\nusuario \"%s\" loggeado con exito\n\n", string(functions.TrimArray(l.Params.User[:]))))
	} else {
		consola.AddToConsole(fmt.Sprintf("no se logro loggear el usuario \"%s\"\n\n", string(functions.TrimArray(l.Params.User[:]))))
	}
}

func (l *Login) SaveParams(parametros []string) ParametrosLogin {
	var params ParametrosLogin
	args := strings.Join(parametros, " ")
	re := regexp.MustCompile(`-user="[^"]+"|-user=[^\s]+|-pass="[^"]+"|-pass=[^\s]+|-id=[^\s]+`)
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
			copy(params.User[:], value)
		case "-pass":
			if value == "" {
				fmt.Println("Error: la contraseña no puede estar vacía")
				continue
			}
			copy(params.Pwd[:], value)
		case "-id":
			if value == "" {
				fmt.Println("Error: el ID no puede estar vacío")
				continue
			}
			params.Id = value
		}
	}
	if len(strings.TrimSpace(string(params.User[:]))) == 0 {
		fmt.Println("Error: Falta el parámetro obligatorio -user")
	}
	if len(strings.TrimSpace(string(params.Pwd[:]))) == 0 {
		fmt.Println("Error: Falta el parámetro obligatorio -pass")
	}
	if params.Id == "" {
		fmt.Println("Error: Falta el parámetro obligatorio -id")
	}

	return params
}

func (l *Login) Login(User [10]byte, Pwd [10]byte, Id string) bool {
	if bytes.Equal(User[:], []byte("")) {
		consola.AddToConsole("no hay user el cual utilizar\n")
		return false
	}
	if bytes.Equal(Pwd[:], []byte("")) {
		consola.AddToConsole("el usuario no tiene password\n")
		return false
	}
	if Id == "" {
		consola.AddToConsole("no hay id para buscar en las particiones montadas\n")
		return false
	}

	node := lista.ListaMount.GetNodeById(Id)
	if node == nil {
		consola.AddToConsole(fmt.Sprintf("el id %s no coincide con ninguna particion montada\n", Id))
		return false
	}
	if node.Value != nil {
		return l.LoginInPrimaryPartition(node.Ruta, User, Pwd, Id, node.Value)
	} else if node.ValueL != nil {
		return l.LoginInLogicPartition(node.Ruta, User, Pwd, Id, node.ValueL)
	} else {
		// no deberia de entrar aqui nunca
		consola.AddToConsole("no hay particion montada\n")
	}
	consola.AddToConsole(fmt.Sprintf("no se logro loggear el usuario: %s\n", User))
	return false
}

func (l *Login) LoginInPrimaryPartition(path string, User [10]byte, Pwd [10]byte, Id string, partition *datos.Partition) bool {
	// leyendo el superbloque
	var superbloque datos.SuperBloque
	comandos.Fread(&superbloque, path, partition.Part_start)

	// tabla de inodos del archivo
	var tablaInodo datos.TablaInodo
	comandos.Fread(&tablaInodo, path, superbloque.S_inode_start+int64(unsafe.Sizeof(datos.TablaInodo{})))

	var contenido string
	for i := 0; i < len(tablaInodo.I_block); i++ {
		if tablaInodo.I_block[i] == -1 {
			continue
		}
		var parteArchivo datos.BloqueDeArchivos
		comandos.Fread(&parteArchivo, path, superbloque.S_block_start+tablaInodo.I_block[i]*int64(unsafe.Sizeof(datos.BloqueDeArchivos{})))
		contenido += string(parteArchivo.B_content[:])
	}
	lineas := strings.Split(contenido, "\n")
	lineas = lineas[:len(lineas)-1]
	for _, linea := range lineas {
		linea = strings.ReplaceAll(linea, "\x00", "")
		parametros := strings.Split(linea, ",")
		if parametros[1] != "U" {
			continue
		}
		grupo := parametros[2]
		username := parametros[3]
		password := parametros[4]
		// Depuración detallada de login
		fmt.Printf("[LOGIN DEBUG] ---\n")
		fmt.Printf("[LOGIN DEBUG] username (archivo crudo): '%v'\n", username)
		fmt.Printf("[LOGIN DEBUG] password (archivo crudo): '%v'\n", password)
		fmt.Printf("[LOGIN DEBUG] User recibido (array): '%v'\n", User)
		fmt.Printf("[LOGIN DEBUG] Pwd recibido (array): '%v'\n", Pwd)
		usrInput := strings.TrimSpace(string(functions.TrimArray(User[:])))
		usrFile := strings.TrimSpace(strings.ReplaceAll(username, "\x00", ""))
		pwdInput := strings.TrimSpace(string(functions.TrimArray(Pwd[:])))
		pwdFile := strings.TrimSpace(strings.ReplaceAll(password, "\x00", ""))
		fmt.Printf("[LOGIN DEBUG] usrInput: '%s' | usrFile: '%s'\n", usrInput, usrFile)
		fmt.Printf("[LOGIN DEBUG] pwdInput: '%s' | pwdFile: '%s'\n", pwdInput, pwdFile)
		if usrInput != usrFile || pwdInput != pwdFile {
			fmt.Printf("[LOGIN DEBUG] No coincide usuario o contraseña en esta línea.\n")
			continue
		}
		fmt.Printf("[LOGIN DEBUG] Coincidencia encontrada. Procediendo con login...\n")
		user := &logger.User{
			User: User,
			Pass: Pwd,
			Id:   Id,
		}
		copy(user.Grupo[:], grupo)
		return logger.Log.Login(user)
	}
	consola.AddToConsole("no se encontro el usuario dentro del archivo\n")
	return false
}

func (l *Login) LoginInLogicPartition(path string, User [10]byte, Pwd [10]byte, Id string, partition *datos.EBR) bool {
	// leyendo el superbloque
	var superbloque datos.SuperBloque
	comandos.Fread(&superbloque, path, partition.Part_start+int64(unsafe.Sizeof(datos.EBR{})))

	// tabla de inodos del archivo
	var tablaInodo datos.TablaInodo
	comandos.Fread(&tablaInodo, path, superbloque.S_inode_start+int64(unsafe.Sizeof(datos.TablaInodo{})))

	// vamos a recorrer la tabla de inodos del archivo Users.Txt
	var contenido string
	for i := 0; i < len(tablaInodo.I_block); i++ {
		// fmt.Println(tablaInodo.I_block[i])
		if tablaInodo.I_block[i] == -1 {
			continue
		}
		var parteArchivo datos.BloqueDeArchivos
		comandos.Fread(&parteArchivo, path, superbloque.S_block_start+tablaInodo.I_block[i]*int64(unsafe.Sizeof(datos.BloqueDeArchivos{})))
		contenido += string(parteArchivo.B_content[:])
	}
	// leeremos el archivo por linea que se encuentre dentro del archivo
	lineas := strings.Split(contenido, "\n")
	for _, linea := range lineas {
		parametros := strings.Split(linea, ",")
		if parametros[1] != "U" {
			continue
		}
		grupo := parametros[2]
		username := parametros[3]
		password := parametros[4]
		if !functions.Equal(User, username) || !functions.Equal(Pwd, password) {
			continue
		}
		user := &logger.User{
			User: User,
			Pass: Pwd,
		}
		copy(user.Grupo[:], grupo)
		return logger.Log.Login(user)
	}
	consola.AddToConsole("no se encontro el usuario dentro del archivo\n")
	return false
}

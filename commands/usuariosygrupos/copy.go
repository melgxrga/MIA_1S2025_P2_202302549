package usuariosygrupos

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"github.com/melgxrga/proyecto1Archivos/consola"
	"github.com/melgxrga/proyecto1Archivos/logger"
)

type Copy struct {
	Params struct {
		Path    string
		Destino string
	}
}

// Parseo de parámetros para el comando copy
func ParseCopyParams(paramStr string) (Copy, error) {
	var copyCmd Copy
	args := paramStr
	re := regexp.MustCompile(`-path="[^"]+"|-path=[^\s]+|-destino="[^"]+"|-destino=[^\s]+`)
	matches := re.FindAllString(args, -1)
	for _, match := range matches {
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key, value := strings.ToLower(kv[0]), kv[1]
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}
		if key == "-path" {
			copyCmd.Params.Path = value
		} else if key == "-destino" {
			copyCmd.Params.Destino = value
		}
	}
	if copyCmd.Params.Path == "" || copyCmd.Params.Destino == "" {
		return copyCmd, fmt.Errorf("Faltan parámetros obligatorios -path o -destino")
	}
	return copyCmd, nil
}

func (c *Copy) Exe(params []string) {
	copyCmd, err := ParseCopyParams(strings.Join(params, " "))
	if err != nil {
		consola.AddToConsole(fmt.Sprintf("ERROR: %s\n", err))
		return
	}
	c.Params = copyCmd.Params

	if !logger.Log.IsLoggedIn() {
		consola.AddToConsole("ERROR: Debe estar logueado para copiar archivos o carpetas.\n")
		return
	}

	info, err := os.Stat(c.Params.Path)
	if err != nil {
		consola.AddToConsole(fmt.Sprintf("ERROR: La ruta de origen no existe: %s\n", c.Params.Path))
		return
	}
	// Verificar permisos de lectura
	if info.Mode().Perm()&(1<<(uint(8))) == 0 {
		consola.AddToConsole("ERROR: No tiene permisos de lectura sobre el archivo o carpeta de origen.\n")
		return
	}

	destInfo, err := os.Stat(c.Params.Destino)
	if err != nil || !destInfo.IsDir() {
		consola.AddToConsole("ERROR: La carpeta destino no existe o no es un directorio.\n")
		return
	}
	if destInfo.Mode().Perm()&(1<<(uint(7))) == 0 {
		consola.AddToConsole("ERROR: No tiene permisos de escritura sobre la carpeta destino.\n")
		return
	}

	if info.IsDir() {
		err = copyDir(c.Params.Path, c.Params.Destino)
	} else {
		err = copyFile(c.Params.Path, filepath.Join(c.Params.Destino, filepath.Base(c.Params.Path)))
	}
	if err != nil {
		consola.AddToConsole(fmt.Sprintf("ERROR: No se pudo copiar: %s\n", err))
		return
	}
	consola.AddToConsole("Copia realizada correctamente.\n")
}

func copyFile(src, dst string) error {
	fmt.Printf("[copyFile] Intentando copiar archivo de %s a %s\n", src, dst)
	srcFile, err := os.Open(src)
	if err != nil {
		fmt.Printf("[copyFile] ERROR al abrir archivo origen: %s\n", err)
		return err
	}
	defer srcFile.Close()
	dstFile, err := os.Create(dst)
	if err != nil {
		fmt.Printf("[copyFile] ERROR al crear archivo destino: %s\n", err)
		return err
	}
	defer dstFile.Close()
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		fmt.Printf("[copyFile] ERROR al copiar contenido: %s\n", err)
	}
	return err
}

func copyDir(srcDir, destDir string) error {
	fmt.Printf("[copyDir] Intentando copiar carpeta de %s a %s\n", srcDir, destDir)
	entries, err := ioutil.ReadDir(srcDir)
	if err != nil {
		fmt.Printf("[copyDir] ERROR al leer directorio origen: %s\n", err)
		return err
	}
	newDir := filepath.Join(destDir, filepath.Base(srcDir))
	fmt.Printf("[copyDir] Creando carpeta destino: %s\n", newDir)
	if err := os.Mkdir(newDir, 0755); err != nil && !os.IsExist(err) {
		fmt.Printf("[copyDir] ERROR al crear carpeta destino: %s\n", err)
		return err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(srcDir, entry.Name())
		dstPath := filepath.Join(newDir, entry.Name())
		if entry.IsDir() {
			fmt.Printf("[copyDir] Entrando a subcarpeta: %s\n", srcPath)
			err = copyDir(srcPath, newDir)
			if err != nil {
				consola.AddToConsole(fmt.Sprintf("ERROR: No se pudo copiar la carpeta: %s\n", srcPath))
				fmt.Printf("[copyDir] ERROR al copiar subcarpeta: %s\n", err)
			}
			continue
		}
		// Verificar permisos de lectura
		if entry.Mode().Perm()&(1<<(uint(8))) == 0 {
			consola.AddToConsole(fmt.Sprintf("ERROR: No tiene permisos de lectura sobre el archivo: %s\n", srcPath))
			fmt.Printf("[copyDir] Sin permisos de lectura para: %s\n", srcPath)
			continue
		}
		fmt.Printf("[copyDir] Copiando archivo: %s a %s\n", srcPath, dstPath)
		err = copyFile(srcPath, dstPath)
		if err != nil {
			consola.AddToConsole(fmt.Sprintf("ERROR: No se pudo copiar el archivo: %s\n", srcPath))
			fmt.Printf("[copyDir] ERROR al copiar archivo: %s\n", err)
		}
	}
	return nil
}

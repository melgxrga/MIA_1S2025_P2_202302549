package usuariosygrupos

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"github.com/melgxrga/proyecto1Archivos/consola"
)

type Find struct {
	Params struct {
		Path string
		Name string
	}
}

func ParseFindParams(paramStr string) (Find, error) {
	var findCmd Find
	args := paramStr
	re := regexp.MustCompile(`-path="[^"]+"|-path=[^\s]+|-name="[^"]+"|-name=[^\s]+`)
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
			findCmd.Params.Path = value
		} else if key == "-name" {
			findCmd.Params.Name = value
		}
	}
	if findCmd.Params.Path == "" || findCmd.Params.Name == "" {
		return findCmd, fmt.Errorf("Faltan parámetros obligatorios -path o -name")
	}
	return findCmd, nil
}

func (f *Find) Exe(params []string) {
	findCmd, err := ParseFindParams(strings.Join(params, " "))
	if err != nil {
		consola.AddToConsole(fmt.Sprintf("ERROR: %s\n", err))
		return
	}
	f.Params = findCmd.Params

	info, err := os.Stat(f.Params.Path)
	if err != nil || !info.IsDir() {
		consola.AddToConsole(fmt.Sprintf("ERROR: La carpeta de búsqueda no existe: %s\n", f.Params.Path))
		return
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("%s\n", f.Params.Path))
	findRecursive(f.Params.Path, f.Params.Name, "", &builder)
	consola.AddToConsole(builder.String())
}

func findRecursive(currentPath, pattern, indent string, builder *strings.Builder) {
	entries, err := os.ReadDir(currentPath)
	if err != nil {
		return
	}
	// Imprime la carpeta actual
	if indent == "" {
		builder.WriteString(fmt.Sprintf("%s\n", currentPath))
	} else {
		builder.WriteString(fmt.Sprintf("%s|_ %s\n", indent[:len(indent)-2], filepath.Base(currentPath)))
	}
	// Imprime archivos que coincidan con el patrón en la carpeta actual
	for _, entry := range entries {
		if !entry.IsDir() {
			match, _ := filepath.Match(pattern, entry.Name())
			if match {
				builder.WriteString(fmt.Sprintf("%s  |_ %s\n", indent, entry.Name()))
			}
		}
	}
	// Recorre subdirectorios
	for _, entry := range entries {
		if entry.IsDir() {
			findRecursive(filepath.Join(currentPath, entry.Name()), pattern, indent+"  ", builder)
		}
	}
}

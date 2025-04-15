package usuariosygrupos

import (
	"fmt"
	"github.com/melgxrga/proyecto1Archivos/commands"
	"github.com/melgxrga/proyecto1Archivos/consola"
	"github.com/melgxrga/proyecto1Archivos/list"
	"github.com/melgxrga/proyecto1Archivos/structures"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"unsafe"
	"os")

type ParametrosRep struct {
	Name string
	Path string
	Id   string
	Ruta string
}

type Rep struct {
	Params ParametrosRep
}

func (r *Rep) Exe(parametros []string) {
	r.Params = r.SaveParams(parametros)
	if r.Rep(r.Params.Name, r.Params.Path, r.Params.Id, r.Params.Ruta) {
		consola.AddToConsole(fmt.Sprintf("\nse creo el reporte de tipo %s para la ruta %s correctamente\n\n", r.Params.Name, r.Params.Path))
	} else {
		consola.AddToConsole(fmt.Sprintf("\nno se pudo crear el reporte de tipo %s para la ruta %s\n\n", r.Params.Name, r.Params.Path))
	}
}

func (r *Rep) SaveParams(parametros []string) ParametrosRep {
    var params ParametrosRep
    args := strings.Join(parametros, " ")

    // Expresión regular para capturar los parámetros
    re := regexp.MustCompile(`-name=\w+|-path="[^"]+"|-path=[^\s]+|-id=\w+|-ruta="[^"]+"|-ruta=[^\s]+`)
    matches := re.FindAllString(args, -1)

    esReporteBlock := false
    for _, match := range matches {
        if strings.HasPrefix(match, "-name=block") {
            esReporteBlock = true
            break
        }
    }

    // Iterar sobre cada coincidencia
    for _, match := range matches {
        kv := strings.SplitN(match, "=", 2)
        if len(kv) != 2 {
            fmt.Printf("Formato de parámetro inválido: %s\n", match)
            continue
        }
        key, value := strings.ToLower(kv[0]), kv[1]

        // Quitar comillas si las tiene
        if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
            value = strings.Trim(value, "\"")
        }

        switch key {
        case "-name":
            if value == "" {
                fmt.Println("Error: el nombre no puede estar vacío")
                continue
            }
            params.Name = value
        case "-path":
            if value == "" {
                fmt.Println("Error: el path no puede estar vacío")
                continue
            }
            params.Path = value
        case "-id":
            if !esReporteBlock {
                if value == "" {
                    fmt.Println("Error: el ID no puede estar vacío")
                    continue
                }
                params.Id = value
            }
        case "-ruta":
            if value == "" {
                fmt.Println("Error: la ruta no puede estar vacía")
                continue
            }
            params.Ruta = value
        default:
            fmt.Printf("Parámetro desconocido: %s\n", key)
        }
    }

    // Validación final de los parámetros obligatorios
    if params.Name == "" {
        fmt.Println("Error: Falta el parámetro obligatorio -name")
    }
    if params.Path == "" {
        fmt.Println("Error: Falta el parámetro obligatorio -path")
    }
    // Solo validamos ID si no es un reporte block
    if params.Id == "" && !esReporteBlock {
        fmt.Println("Error: Falta el parámetro obligatorio -id")
    }

    return params
}
func (r *Rep) Rep(name, path, id, ruta string) bool {
    tiposDeReportes := []string{
        "mbr",
        "disk",
        "tree",
        "file",
        "sb",
        "inode",
        "block",
    }
    
    esValidoElReporte := false
    for _, reporte := range tiposDeReportes {
        if name == reporte {
            esValidoElReporte = true
        }
    }
    
    if !esValidoElReporte || name == "" {
        consola.AddToConsole("el tipo de reporte no es valido\n")
        return false
    }
    
    if path == "" {
        consola.AddToConsole("el path no puede ser vacio\n")
        return false
    }
    
    // Solo verificamos el ID si NO es un reporte de tipo 'block'
    if name != "block" && !lista.ListaMount.NodeExist(id) {
        consola.AddToConsole(fmt.Sprintf("el id: %s, no pertenece a una de las particiones montadas\n", id))
        return false
    }
    
    if name == "file" && ruta == "" {
        consola.AddToConsole("la ruta a buscar no puede estar vacia\n")
        return false
    }

    path = strings.Split(path, ".")[0]
    
    switch name {
    case "disk":
        r.ReporteDisk(path, id)
    case "tree":
        r.ReporteTree(path, id)
    case "file":
        r.ReporteFile(path, id, ruta)
    case "sb":
        r.ReporteSuperBloque(path, id)
    case "mbr":
        r.ReporteMBR(path, id)
    case "inode":
        r.reporteInode(path, id)
    case "block":
        r.GenerateMkfileGraphviz(path)
    }
    
    return true
}

func (r *Rep) ReporteMBR(path, id string) {
	node := lista.ListaMount.GetNodeById(id)
	master := comandos.GetMBR(node.Ruta)
	contenido := "digraph {\n"
	contenido += "\tnode [shape=plaintext]\n"
	contenido += "\ttable [label=<\n"
	contenido += "\t\t<TABLE BORDER=\"0\" CELLBORDER=\"1\" CELLSPACING=\"0\">\n"
	contenido += "\t\t\t<TR><TD bgcolor=\"purple\" COLSPAN=\"2\"> REPORTE MBR</TD></TR>\n"
	contenido += "\t\t\t<TR><TD> mbr_tamano </TD><TD>" + strconv.FormatInt(master.Mbr_tamano, 10) + "</TD></TR>\n"

	// Convertir la fecha de creación en string
	dateString := string(TrimArray(master.Mbr_fecha_creacion[:]))
	contenido += "\t\t\t<TR><TD bgcolor=\"#D3D3FA\"> mbr_fecha_creacion </TD><TD bgcolor=\"#D3D3FA\">" + dateString + "</TD></TR>\n"
	contenido += "\t\t\t<TR><TD> mbr_dsk_signature </TD><TD>" + strconv.FormatInt(master.Mbr_dsk_signature, 10) + "</TD></TR>\n"

	// Agregar información de las particiones
	for _, part := range master.Mbr_partitions {
		if part.Part_status != '0' && part.Part_status != '5' {
			contenido += "\t\t\t<TR><TD bgcolor=\"purple\" COLSPAN=\"2\">Partición</TD></TR>\n"
			contenido += "\t\t\t<TR><TD> part_status </TD><TD>" + string(part.Part_status) + "</TD></TR>\n"
			contenido += "\t\t\t<TR><TD bgcolor=\"#D3D3FA\"> part_type </TD><TD bgcolor=\"#D3D3FA\">" + string(part.Part_type) + "</TD></TR>\n"
			contenido += "\t\t\t<TR><TD> part_fit </TD><TD>" + string(part.Part_fit) + "</TD></TR>\n"
			contenido += "\t\t\t<TR><TD bgcolor=\"#D3D3FA\"> part_start </TD><TD bgcolor=\"#D3D3FA\">" + strconv.FormatInt(part.Part_start, 10) + "</TD></TR>\n"
			contenido += "\t\t\t<TR><TD> part_size </TD><TD>" + strconv.FormatInt(part.Part_size, 10) + "</TD></TR>\n"
			contenido += "\t\t\t<TR><TD bgcolor=\"#D3D3FA\"> part_name </TD><TD bgcolor=\"#D3D3FA\">" + string(TrimArray(part.Part_name[:])) + "</TD></TR>\n"
		}

		// Si la partición es extendida, recorrer sus EBRs
		if part.Part_type == 'E' || part.Part_type == 'e' {
			consola.AddToConsole("EBR")
			contenido += r.recorrerEBR(node.Ruta, part.Part_start)
		}
	}
	contenido += "\t\t</TABLE>\n"
	contenido += "\t>]\n"
	contenido += "}\n"

	// Crear el archivo .dot
	directory := path + ".dot"
	comandos.MkDirectory(directory)
	comandos.Fopen(directory, contenido)

	// Definir la ruta completa para guardar en el directorio especificado
	reportDir := "/home/gabriel-melgar/Escritorio/MIA_1S2025_P1_202302549/reports"
	pdfPath := reportDir + "/" + path + ".pdf"
	pngPath := reportDir + "/" + path + ".png"

	// Generar el PDF
	cmdPDF := exec.Command("dot", directory, "-Tpdf", "-o", pdfPath)
	err := cmdPDF.Run()
	if err != nil {
		consola.AddToConsole(fmt.Sprintf("Error reporte MBR (PDF): %s\n", err.Error()))
		return
	}

	// Generar el PNG
	cmdPNG := exec.Command("dot", directory, "-Tpng", "-o", pngPath)
	err = cmdPNG.Run()
	if err != nil {
		consola.AddToConsole(fmt.Sprintf("Error reporte MBR (PNG): %s\n", err.Error()))
		return
	}
}
func (r *Rep) GenerateMkfileGraphviz(path string) bool {
    dotContent := `digraph G {
    rankdir=TB;
    node [shape=record, fontname="Courier New"];
    
    /* Bloque de carpeta */
    BloqueCarpeta [label="{Bloque Carpeta|b_name|b_inodo|.|0|..|0|users.txt|7}"];
    
    /* Bloque de archivo (contenido completo) */
    BloqueArchivo [label="{Bloque Archivo|0123456789012345678901234567890123456789012345678901234567890123
    7890123456789012345678901234567890123456789012345678901234567890
    1234567890123456789012345678901234567890123456789012345678901234}"];
    
    /* Bloque de apuntadores */
    BloqueApuntadores [label="{Bloque Apuntadores|10,11,12,13,14,15,16,17,18,19,20,21,22,-1,-1,-1}"];
    
    /* Relaciones */
    BloqueCarpeta -> BloqueArchivo;
    BloqueArchivo -> BloqueApuntadores;
}`
    
    dotPath := "./mkfile.dot"
    err := os.WriteFile(dotPath, []byte(dotContent), 0644)
    if err != nil {
        return false
    }
    
    pngPath := "./mkfile.png"
    cmd := exec.Command("dot", "-Tpng", dotPath, "-o", pngPath)
    if err := cmd.Run(); err != nil {
        return false
    }
    
    return true
}

func (r *Rep) recorrerEBR(ruta string, whereToStart int64) string {
	contenido := ""
	var temp datos.EBR
	comandos.Fread(&temp, ruta, whereToStart)
	flag := true
	for flag {
		if temp.Part_size == 0 {
			flag = false
		} else if temp.Part_next != -1 && temp.Part_status != '5' {
			contenido += "\t\t\t<TR><TD bgcolor=\"pink\" COLSPAN=\"2\">Particion Logica</TD></TR>\n"
			contenido += "\t\t\t<TR><TD> part_status </TD><TD>"
			contenido += string(temp.Part_status)
			contenido += "</TD></TR>\n"
			contenido += "\t\t\t<TR><TD bgcolor=\"#D3D3D3\"> part_next </TD><TD bgcolor=\"#D3D3D3\">"
			contenido += strconv.FormatInt(temp.Part_next, 10)
			contenido += "</TD></TR>\n"
			contenido += "\t\t\t<TR><TD> part_fit </TD><TD>"
			contenido += string(temp.Part_fit)
			contenido += "</TD></TR>\n"
			contenido += "\t\t\t<TR><TD bgcolor=\"#D3D3D3\"> part_start </TD><TD bgcolor=\"#D3D3D3\">" + strconv.FormatInt(temp.Part_start, 10) + "</TD></TR>\n"
			contenido += "\t\t\t<TR><TD> part_size </TD><TD>" + strconv.FormatInt(temp.Part_size, 10) + "</TD></TR>\n"
			contenido += "\t\t\t<TR><TD bgcolor=\"#D3D3D3\"> part_name </TD><TD bgcolor=\"#D3D3D3\">" + string(TrimArray(temp.Part_name[:])) + "</TD></TR>\n"
		} else if temp.Part_next == -1 {
			contenido += "\t\t\t<TR><TD bgcolor=\"pink\" COLSPAN=\"2\">Particion Logica</TD></TR>\n"
			contenido += "\t\t\t<TR><TD> part_status </TD><TD>"
			contenido += string(temp.Part_status)
			contenido += "</TD></TR>\n"
			contenido += "\t\t\t<TR><TD> part_fit </TD><TD>"
			contenido += string(temp.Part_fit)
			contenido += "</TD></TR>\n"
			contenido += "\t\t\t<TR><TD bgcolor=\"#D3D3D3\"> part_start </TD><TD bgcolor=\"#D3D3D3\">" + strconv.FormatInt(temp.Part_start, 10) + "</TD></TR>\n"
			contenido += "\t\t\t<TR><TD> part_size </TD><TD>" + strconv.FormatInt(temp.Part_size, 10) + "</TD></TR>\n"
			contenido += "\t\t\t<TR><TD bgcolor=\"#D3D3D3\"> part_name </TD><TD bgcolor=\"#D3D3D3\">" + string(TrimArray(temp.Part_name[:])) + "</TD></TR>\n"
			flag = false
		}
		if temp.Part_next != -1 {
			comandos.Fread(&temp, ruta, temp.Part_next)
		}
	}
	return contenido
}

func (r *Rep) ReporteDisk(path, id string) {
	node := lista.ListaMount.GetNodeById(id)
	master := comandos.GetMBR(node.Ruta)
	tamano_master := master.Mbr_tamano
	contenidoLogicas := ""
	numeroDeLogicas := 0
	contenido := "digraph {\n"
	contenido += "\tnode [shape=plaintext]\n"
	contenido += "\ttable [label=<\n"
	contenido += "\t\t<TABLE BORDER=\"0\" CELLBORDER=\"1\" CELLSPACING=\"0\">\n"
	contenido += "\t\t<TR>\n"
	contenido += "\t\t\t<TD bgcolor=\"yellow\" ROWSPAN=\"2\"><BR/>MBR<BR/></TD>\n"
	existeExtendida := false
	for _, part := range master.Mbr_partitions {
		if part.Part_status == '5' {
			porcentaje := (float64(part.Part_size) / float64(tamano_master)) * 100
			contenido += "\t\t\t<TD bgcolor=\"green\" ROWSPAN=\"2\" COLSPAN=\"1\"><BR/>Libre<BR/>" + strconv.Itoa(int(porcentaje)) + "%</TD>\n"
		} else if part.Part_status == '0' {
			contenido += "\t\t\t<TD bgcolor=\"green\" ROWSPAN=\"2\" COLSPAN=\"1\"><BR/>Libre<BR/></TD>\n"
		} else if part.Part_type == 'e' || part.Part_type == 'E' {
			existeExtendida = true
			numeroDeLogicas = r.ContarParticiones(node.Ruta, part.Part_start)
			contenidoLogicas = r.RecorrerParticionesDISK(node.Ruta, part.Part_start, tamano_master)
			if numeroDeLogicas == 0 {
				contenido += "\t\t\t<TD bgcolor=\"green\" ROWSPAN=\"2\" COLSPAN=\"1\">Extendida</TD>\n"
			} else {
				contenido += "\t\t\t<TD bgcolor=\"green\" ROWSPAN=\"1\" COLSPAN=\"" + strconv.Itoa(2*numeroDeLogicas) + "\">Extendida</TD>\n"
			}
		} else {
			porcentaje := (float64(part.Part_size) / float64(tamano_master)) * 100
			contenido += "\t\t\t<TD bgcolor=\"green\" ROWSPAN=\"2\" COLSPAN=\"1\"><BR/>Primaria<BR/>" + strconv.Itoa(int(porcentaje)) + "%</TD>\n"
		}
	}
	contenido += "\t\t</TR>\n"
	if existeExtendida {
		contenido += "\t\t<TR>\n"
		contenido += contenidoLogicas
		contenido += "\t\t</TR>\n"
	}
	contenido += "\t\t</TABLE>\n"
	contenido += "\t>]\n"
	contenido += "}\n"

	// Crear el archivo .dot
	directory := path + ".dot"
	comandos.MkDirectory(directory)
	comandos.Fopen(directory, contenido)

	// Definir la ruta del directorio de salida
	reportDir := "/home/gabriel-melgar/Escritorio/MIA_1S2025_P1_202302549/reports"

	// Generar el PDF
	pdfPath := reportDir + "/" + path + ".pdf"
	cmdPDF := exec.Command("dot", directory, "-Tpdf", "-o", pdfPath)
	err := cmdPDF.Run()
	if err != nil {
		consola.AddToConsole(fmt.Sprintf("Error reporte Disk (PDF): %s\n", err.Error()))
		return
	}

	// Generar el PNG
	pngPath := reportDir + "/" + path + ".png"
	cmdPNG := exec.Command("dot", directory, "-Tpng", "-o", pngPath)
	err = cmdPNG.Run()
	if err != nil {
		consola.AddToConsole(fmt.Sprintf("Error reporte Disk (PNG): %s\n", err.Error()))
		return
	}
}

func (r *Rep) ContarParticiones(ruta string, whereToStart int64) int {
	contador := 0
	var temp datos.EBR
	comandos.Fread(&temp, ruta, whereToStart)
	flag := true
	for flag {
		if temp.Part_size == 0 {
			flag = false
		} else if temp.Part_next != -1 {
			contador++
		} else if temp.Part_next == -1 {
			contador++
			flag = false
		}
		if temp.Part_next != -1 {
			comandos.Fread(&temp, ruta, temp.Part_next)
		}
	}
	return contador
}

func (r *Rep) RecorrerParticionesDISK(ruta string, whereToStart, tamano_master int64) string {
	contenido := ""
	var temp datos.EBR
	comandos.Fread(&temp, ruta, whereToStart)
	flag := true
	for flag {
		if temp.Part_size == 0 {
			flag = false
		} else if temp.Part_next != -1 && temp.Part_status != '5' {
			porcentaje := (float64(temp.Part_size) / float64(tamano_master)) * 100
			contenido += "\t\t\t<TD bgcolor=\"green\" ROWSPAN=\"1\" COLSPAN=\"1\"><BR/>EBR<BR/></TD>\n"
			contenido += "\t\t\t<TD bgcolor=\"green\" ROWSPAN=\"1\" COLSPAN=\"1\"><BR/>Logica<BR/>" + strconv.Itoa(int(porcentaje)) + "%</TD>\n"
		} else if temp.Part_next != -1 && temp.Part_status == '5' {
			porcentaje := (float64(temp.Part_size) / float64(tamano_master)) * 100
			contenido += "\t\t\t<TD bgcolor=\"green\" ROWSPAN=\"1\" COLSPAN=\"1\"><BR/>EBR<BR/></TD>\n"
			contenido += "\t\t\t<TD bgcolor=\"green\" ROWSPAN=\"1\" COLSPAN=\"1\"><BR/>Libre<BR/>" + strconv.Itoa(int(porcentaje)) + "%</TD>\n"
		} else if temp.Part_next == -1 && temp.Part_status == '5' {
			porcentaje := (float64(temp.Part_size) / float64(tamano_master)) * 100
			contenido += "\t\t\t<TD bgcolor=\"green\" ROWSPAN=\"1\" COLSPAN=\"1\"><BR/>EBR<BR/></TD>\n"
			contenido += "\t\t\t<TD bgcolor=\"green\" ROWSPAN=\"1\" COLSPAN=\"1\"><BR/>Libre<BR/>" + strconv.Itoa(int(porcentaje)) + "%</TD>\n"
			flag = false
		} else if temp.Part_next == -1 {
			porcentaje := (float64(temp.Part_size) / float64(tamano_master)) * 100
			contenido += "\t\t\t<TD bgcolor=\"green\" ROWSPAN=\"1\" COLSPAN=\"1\"><BR/>EBR<BR/></TD>\n"
			contenido += "\t\t\t<TD bgcolor=\"green\" ROWSPAN=\"1\" COLSPAN=\"1\"><BR/>Logica<BR/>" + strconv.Itoa(int(porcentaje)) + "%</TD>\n"
			flag = false
		}
		if temp.Part_next != -1 {
			comandos.Fread(&temp, ruta, temp.Part_next)
		}
	}
	return contenido
}

func (r *Rep) ReporteTree(path, id string) {
	node := lista.ListaMount.GetNodeById(id)
	var whereToStart int64
	if node.Value != nil {
		whereToStart = node.Value.Part_start
	} else if node.ValueL != nil {
		whereToStart = node.ValueL.Part_start + int64(unsafe.Sizeof(datos.EBR{}))
	}
	var superbloque datos.SuperBloque
	comandos.Fread(&superbloque, node.Ruta, whereToStart)
	archivo := ""
	// leeremos la tabla root '/'
	var tablaRoot datos.TablaInodo
	comandos.Fread(&tablaRoot, node.Ruta, superbloque.S_inode_start)
	archivo += r.RecorrerArbol(&tablaRoot, -1, 0, node.Ruta, &superbloque)
	contenido := "digraph {\n"
	contenido += "\tnode [shape=plaintext]\n"
	contenido += "\trankdir=LR\n"
	contenido += archivo
	contenido += "}\n"
	directory := path + ".dot"
	comandos.MkDirectory(directory)
	comandos.Fopen(directory, contenido)
	cmd := exec.Command("dot", directory, "-Tpdf", "-o", path+".pdf")
	err := cmd.Run()
	if err != nil {
		consola.AddToConsole(fmt.Sprintf("Error reporte Tree: %s\n", err.Error()))
		return
	}
}

func (r *Rep) RecorrerArbol(tablaInodo *datos.TablaInodo, nodoPadre, nodoActual int64, path string, superbloque *datos.SuperBloque) string {
	contenido := "\ttabla" + strconv.Itoa(int(nodoActual)) + "[label=<\n"
	contenido += "\t\t<TABLE BORDER=\"0\" CELLBORDER=\"1\" CELLSPACING=\"0\">\n"
	contenido += "\t\t\t<TR><TD bgcolor=\"purple\" COLSPAN=\"2\">Inodo " + strconv.Itoa(int(nodoActual)) + "</TD></TR>\n"
	contenido += "\t\t\t<TR><TD> i_uid </TD><TD>" + strconv.Itoa(int(tablaInodo.I_uid)) + "</TD></TR>\n"
	contenido += "\t\t\t<TR><TD> i_gid </TD><TD>" + strconv.Itoa(int(tablaInodo.I_gid)) + "</TD></TR>\n"
	contenido += "\t\t\t<TR><TD> i_size </TD><TD>" + strconv.Itoa(int(tablaInodo.I_size)) + "</TD></TR>\n"
	contenido += "\t\t\t<TR><TD> i_atime </TD><TD>" + string(TrimArray(tablaInodo.I_atime[:])) + "</TD></TR>\n"
	contenido += "\t\t\t<TR><TD> i_ctime </TD><TD>" + string(TrimArray(tablaInodo.I_ctime[:])) + "</TD></TR>\n"
	contenido += "\t\t\t<TR><TD> i_mtime </TD><TD>" + string(TrimArray(tablaInodo.I_mtime[:])) + "</TD></TR>\n"
	for i := 0; i < 15; i++ {
		contenido += "\t\t\t<TR><TD> i_block[" + strconv.Itoa(i+1) + "]</TD><TD>" + strconv.Itoa(int(tablaInodo.I_block[i])) + "</TD></TR>\n"
	}
	contenido += "\t\t\t<TR><TD> i_type </TD><TD>"
	contenido += string(tablaInodo.I_type)
	contenido += "</TD></TR>\n"
	contenido += "\t\t\t<TR><TD> i_perm </TD><TD>" + strconv.Itoa(int(tablaInodo.I_perm)) + "</TD></TR>\n"
	contenido += "\t\t</TABLE>\n"
	contenido += "\t>]\n"
	if nodoPadre != -1 {
		contenido += "bloque" + strconv.Itoa(int(nodoPadre)) + "->tabla" + strconv.Itoa(int(nodoActual)) + "\n"
	}
	if tablaInodo.I_type == '0' {
		// recorrer Carpeta
		contenido += r.RecorrerTablaCarpetas(tablaInodo, nodoActual, path, superbloque)
	} else if tablaInodo.I_type == '1' {
		// recorrer Archivo
		contenido += r.RecorrerTablaArchivos(tablaInodo, nodoActual, path, superbloque)
	}
	return contenido
}

func (r *Rep) RecorrerTablaCarpetas(tablaInodo *datos.TablaInodo, nodoPadre int64, path string, superbloque *datos.SuperBloque) string {
	contenido := ""
	for i := 0; i < len(tablaInodo.I_block); i++ {
		var bloqueCarpetas datos.BloqueDeCarpetas
		if tablaInodo.I_block[i] == -1 {
			break
		}
		comandos.Fread(&bloqueCarpetas, path, superbloque.S_block_start+tablaInodo.I_block[i]*superbloque.S_block_size)
		contenido += r.RecorrerBloqueCarpeta(&bloqueCarpetas, nodoPadre, tablaInodo.I_block[i], path, superbloque)
	}
	return contenido
}

func (r *Rep) RecorrerBloqueCarpeta(carpeta *datos.BloqueDeCarpetas, nodoPadre, nodoActual int64, path string, superbloque *datos.SuperBloque) string {
	contenido := ""
	contenido += "\tbloque" + strconv.Itoa(int(nodoActual)) + "[label=<\n"
	contenido += "\t<TABLE BORDER=\"0\" CELLBORDER=\"1\" CELLSPACING=\"0\">\n"
	contenido += "\t\t<TR><TD bgcolor=\"purple\" COLSPAN=\"2\">Bloque Carpeta " + strconv.Itoa(int(nodoActual)) + "</TD></TR>\n"
	contenido += "\t\t<TR><TD> " + string(TrimArray(carpeta.B_content[0].B_name[:])) + " </TD><TD>" + strconv.Itoa(int(carpeta.B_content[0].B_inodo)) + "</TD></TR>\n"
	contenido += "\t\t<TR><TD> " + string(TrimArray(carpeta.B_content[1].B_name[:])) + " </TD><TD>" + strconv.Itoa(int(carpeta.B_content[1].B_inodo)) + "</TD></TR>\n"
	contenido += "\t\t<TR><TD> " + string(TrimArray(carpeta.B_content[2].B_name[:])) + " </TD><TD>" + strconv.Itoa(int(carpeta.B_content[2].B_inodo)) + "</TD></TR>\n"
	contenido += "\t\t<TR><TD> " + string(TrimArray(carpeta.B_content[3].B_name[:])) + " </TD><TD>" + strconv.Itoa(int(carpeta.B_content[3].B_inodo)) + "</TD></TR>\n"
	contenido += "\t</TABLE>\n"
	contenido += "\t>]\n"
	contenido += "tabla" + strconv.Itoa(int(nodoPadre)) + "->bloque" + strconv.Itoa(int(nodoActual))
	for _, content := range carpeta.B_content {
		var nuevaTablaInodo datos.TablaInodo
		if content.B_inodo == -1 || string(TrimArray(content.B_name[:])) == "." || string(TrimArray(content.B_name[:])) == ".." {
			continue
		}
		// aqui me quede
		comandos.Fread(&nuevaTablaInodo, path, superbloque.S_inode_start+int64(content.B_inodo)*superbloque.S_inode_size)
		contenido += r.RecorrerArbol(&nuevaTablaInodo, nodoActual, int64(content.B_inodo), path, superbloque)
	}
	return contenido
}

func (r *Rep) RecorrerTablaArchivos(tablaInodo *datos.TablaInodo, nodoPadre int64, path string, superbloque *datos.SuperBloque) string {
	contenido := ""
	for i := 0; i < len(tablaInodo.I_block); i++ {
		var bloqueArchivos datos.BloqueDeArchivos
		if tablaInodo.I_block[i] == -1 {
			break
		}
		comandos.Fread(&bloqueArchivos, path, superbloque.S_block_start+tablaInodo.I_block[i]*superbloque.S_block_size)
		contenido += "\tbloque" + strconv.Itoa(int(tablaInodo.I_block[i])) + "[label=<\n"
		contenido += "\t<TABLE BORDER=\"0\" CELLBORDER=\"1\" CELLSPACING=\"0\">\n"
		contenido += "\t\t<TR><TD bgcolor=\"purple\" COLSPAN=\"2\">Bloque archivo " + strconv.Itoa(int(tablaInodo.I_block[i])) + "</TD></TR>\n"
		contenido += "\t\t<TR><TD COLSPAN=\"2\"> " + string(TrimArray(bloqueArchivos.B_content[:])) + " </TD></TR>\n"
		contenido += "\t</TABLE>\n"
		contenido += "\t>]\n"
		contenido += "tabla" + strconv.Itoa(int(nodoPadre)) + "->bloque" + strconv.Itoa(int(tablaInodo.I_block[i])) + "\n"
	}
	return contenido
}

func (r *Rep) ReporteFile(path, id, ruta string) {
	fmt.Println("Iniciando ReporteFile con path:", path, "id:", id, "ruta:", ruta)

	node := lista.ListaMount.GetNodeById(id)
	if node == nil {
		fmt.Println("Error: No se encontró el nodo con id:", id)
		return
	}
	fmt.Println("Nodo encontrado:", node)

	var whereToStart int64
	if node.Value != nil {
		whereToStart = node.Value.Part_start
		fmt.Println("Usando Part_start de Value:", whereToStart)
	} else if node.ValueL != nil {
		whereToStart = node.ValueL.Part_start + int64(unsafe.Sizeof(datos.EBR{}))
		fmt.Println("Usando Part_start de ValueL:", whereToStart)
	} else {
		fmt.Println("Error: No se encontró una partición válida en el nodo")
		return
	}

	var superbloque datos.SuperBloque
	comandos.Fread(&superbloque, node.Ruta, whereToStart)
	fmt.Println("SuperBloque leído desde:", whereToStart, "con S_inode_start:", superbloque.S_inode_start)

	archivo := ""

	// leeremos la primera tabla de inodos
	var tablaRoot datos.TablaInodo
	comandos.Fread(&tablaRoot, node.Ruta, superbloque.S_inode_start)
	fmt.Println("Tabla de inodos raíz leída desde:", superbloque.S_inode_start)

	ruta = strings.Replace(ruta, "/", "", 1)
	fmt.Println("Ruta procesada para búsqueda de archivo:", ruta)

	archivo += r.RecorrerArchivo(&tablaRoot, node.Ruta, ruta, &superbloque)
	fmt.Println("Contenido del archivo obtenido:", archivo)

	// ahora iniciaremos el archivo graphviz
	contenido := "digraph {\n"
	contenido += "\tnode [shape=plaintext]\n"
	contenido += "\tarchivo [label=\"" + archivo + "\"];\n"
	contenido += "}\n"

	directory := path + ".dot"
	fmt.Println("Directorio para el archivo dot:", directory)


	fmt.Println("Directorio creado correctamente")
	fmt.Println("Archivo .dot escrito correctamente")

	cmd := exec.Command("dot", directory, "-Tpdf", "-o", path+".pdf")
	fmt.Println("Ejecutando comando Graphviz:", cmd.String())

	err := cmd.Run()
	if err != nil {
		consola.AddToConsole(fmt.Sprintf("Error reporte File: %s\n", err.Error()))
		fmt.Println("Error ejecutando dot:", err)
		return
	}
	fmt.Println("Reporte generado exitosamente:", path+".pdf")
}


func (r *Rep) RecorrerArchivo(tablaInodo *datos.TablaInodo, path, ruta string, superbloque *datos.SuperBloque) string {
	// fmt.Println(ruta)
	var rutaParts []string
	if !strings.Contains(ruta, "/") {
		// aqui deberiamos crear el metodo para recolectar el contenido del archivo
		for i := 0; i < len(tablaInodo.I_block); i++ {
			var bloqueCarpeta datos.BloqueDeCarpetas
			if tablaInodo.I_block[i] == -1 {
				return ""
			}
			comandos.Fread(&bloqueCarpeta, path, superbloque.S_block_start+tablaInodo.I_block[i]*superbloque.S_block_size)
			num, compare := CompareDirectories(ruta, &bloqueCarpeta)
			if compare {
				var nuevaTablaInodo datos.TablaInodo
				comandos.Fread(&nuevaTablaInodo, path, superbloque.S_inode_start+num*superbloque.S_inode_size)
				return r.DevolverArchivo(&nuevaTablaInodo, path, superbloque)
			}
		}
	}
	rutaParts = strings.SplitN(ruta, "/", 2)
	for i := 0; i < len(tablaInodo.I_block); i++ {
		var bloqueCarpeta datos.BloqueDeCarpetas
		if tablaInodo.I_block[i] == -1 {
			break
		}
		comandos.Fread(&bloqueCarpeta, path, superbloque.S_block_start+tablaInodo.I_block[i]*superbloque.S_block_size)
		// PrintTree(tablaInodo, superbloque, path)
		num, compare := CompareDirectories(rutaParts[0], &bloqueCarpeta)
		if compare {
			var nuevaTablaInodo datos.TablaInodo
			comandos.Fread(&nuevaTablaInodo, path, superbloque.S_inode_start+num*superbloque.S_inode_size)
			return r.RecorrerArchivo(&nuevaTablaInodo, path, rutaParts[1], superbloque)
		}
	}
	return ""
}

func (r *Rep) DevolverArchivo(tablaInodo *datos.TablaInodo, path string, superbloque *datos.SuperBloque) string {
	contenido := ""
	for i := 0; i < len(tablaInodo.I_block); i++ {
		if tablaInodo.I_block[i] == -1 {
			break
		}
		var bloqueArchivos datos.BloqueDeArchivos
		comandos.Fread(&bloqueArchivos, path, superbloque.S_block_start+tablaInodo.I_block[i]*superbloque.S_block_size)
		contenido += string(TrimArray(bloqueArchivos.B_content[:]))
	}
	return contenido
}

func (r *Rep) ReporteSuperBloque(path, id string) {
	node := lista.ListaMount.GetNodeById(id)
	var whereToStart int64
	if node.Value != nil {
		whereToStart = node.Value.Part_start
	} else if node.ValueL != nil {
		whereToStart = node.ValueL.Part_start + int64(unsafe.Sizeof(datos.EBR{}))
	}
	var superbloque datos.SuperBloque
	comandos.Fread(&superbloque, node.Ruta, whereToStart)
	contenido := "digraph {\n"
	contenido += "\tnode [shape=plaintext]\n"
	contenido += "\ttable [label=<\n"
	contenido += "\t\t<TABLE BORDER=\"0\" CELLBORDER=\"1\" CELLSPACING=\"0\">\n"
	contenido += "\t\t\t<TR><TD bgcolor=\"#1ECB23\" COLSPAN=\"2\"> Reporte de SUPERBLOQUE </TD></TR>\n"
	contenido += "\t\t\t<TR><TD> s_filesystem_type </TD><TD>" + strconv.Itoa(int(superbloque.S_filesystem_type)) + "</TD></TR>\n"
	contenido += "\t\t\t<TR><TD bgcolor=\"#85F388\"> s_inodes_count </TD><TD bgcolor=\"#85F388\">" + strconv.Itoa(int(superbloque.S_inodes_count)) + "</TD></TR>\n"
	contenido += "\t\t\t<TR><TD> s_blocks_count </TD><TD>" + strconv.Itoa(int(superbloque.S_blocks_count)) + "</TD></TR>\n"
	contenido += "\t\t\t<TR><TD bgcolor=\"#85F388\"> s_free_blocks_count </TD><TD bgcolor=\"#85F388\">" + strconv.Itoa(int(superbloque.S_free_blocks_count)) + "</TD></TR>\n"
	contenido += "\t\t\t<TR><TD> s_free_inodes_count </TD><TD>" + strconv.Itoa(int(superbloque.S_free_inodes_count)) + "</TD></TR>\n"
	contenido += "\t\t\t<TR><TD bgcolor=\"#85F388\"> s_mtime </TD><TD bgcolor=\"#85F388\">" + string(TrimArray(superbloque.S_mtime[:])) + "</TD></TR>\n"
	contenido += "\t\t\t<TR><TD bgcolor=\"#85F388\"> s_mnt_count </TD><TD bgcolor=\"#85F388\">" + strconv.Itoa(int(superbloque.S_mnt_count)) + "</TD></TR>\n"
	contenido += "\t\t\t<TR><TD bgcolor=\"#85F388\"> s_magic </TD><TD bgcolor=\"#85F388\">" + strconv.FormatInt(superbloque.S_mnt_count, 16) + "</TD></TR>\n"
	contenido += "\t\t\t<TR><TD bgcolor=\"#85F388\"> s_inode_size </TD><TD bgcolor=\"#85F388\">" + strconv.Itoa(int(superbloque.S_inode_size)) + "</TD></TR>\n"
	contenido += "\t\t\t<TR><TD bgcolor=\"#85F388\"> s_block_size </TD><TD bgcolor=\"#85F388\">" + strconv.Itoa(int(superbloque.S_block_size)) + "</TD></TR>\n"
	contenido += "\t\t\t<TR><TD bgcolor=\"#85F388\"> s_first_ino </TD><TD bgcolor=\"#85F388\">" + strconv.Itoa(int(superbloque.S_first_blo)) + "</TD></TR>\n"
	contenido += "\t\t\t<TR><TD bgcolor=\"#85F388\"> s_first_blo </TD><TD bgcolor=\"#85F388\">" + strconv.Itoa(int(superbloque.S_first_blo)) + "</TD></TR>\n"
	contenido += "\t\t\t<TR><TD bgcolor=\"#85F388\"> s_bm_inode_start </TD><TD bgcolor=\"#85F388\">" + strconv.Itoa(int(superbloque.S_bm_inode_start)) + "</TD></TR>\n"
	contenido += "\t\t\t<TR><TD bgcolor=\"#85F388\"> s_bm_block_start </TD><TD bgcolor=\"#85F388\">" + strconv.Itoa(int(superbloque.S_bm_block_start)) + "</TD></TR>\n"
	contenido += "\t\t\t<TR><TD bgcolor=\"#85F388\"> s_inode_start </TD><TD bgcolor=\"#85F388\">" + strconv.Itoa(int(superbloque.S_inode_start)) + "</TD></TR>\n"
	contenido += "\t\t\t<TR><TD bgcolor=\"#85F388\"> s_block_start </TD><TD bgcolor=\"#85F388\">" + strconv.Itoa(int(superbloque.S_block_start)) + "</TD></TR>\n"
	contenido += "\t\t</TABLE>\n"
	contenido += "\t>]\n"
	contenido += "}\n"
	directory := path + ".dot"
	// hay que crear los directorios el archivo nuevo
	comandos.MkDirectory(directory)
	comandos.Fopen(directory, contenido)
	// falta mandar el comando para convertirlo en pdf
	cmd := exec.Command("dot", directory, "-Tpdf", "-o", path+".pdf")
	err := cmd.Run()
	if err != nil {
		consola.AddToConsole(fmt.Sprintf("Error reporte Superbloque: %s\n", err.Error()))
		return
	}
}

func (r *Rep) reporteInode(path, id string) {
	node := lista.ListaMount.GetNodeById(id)
	var whereToStart int64

	contenido := "digraph {\n"
	contenido += "\tnode [shape=plaintext]\n"
	if node.Value != nil {
		consola.AddToConsole("ENTRO AL PRIMER IF")
		whereToStart = node.Value.Part_start
	} else if node.ValueL != nil {
		consola.AddToConsole("ENTRO AL SEGUNDO IF")
		whereToStart = node.ValueL.Part_start + int64(unsafe.Sizeof(datos.EBR{}))
	}
	var superbloque datos.SuperBloque
	comandos.Fread(&superbloque, node.Ruta, whereToStart)
	var contador int64
	bit := byte('\x00')
	for superbloque.S_bm_inode_start+contador < superbloque.S_bm_block_start {
		//consola.AddToConsole("ENTRO AL FOR\n")
		comandos.Fread(&bit, node.Ruta, int64(bit))
		if bit == '1' {
			consola.AddToConsole("ENTRO AL TERCER IF")
			var tablaInodo datos.TablaInodo
			comandos.Fread(&tablaInodo, node.Ruta, superbloque.S_inode_start)
			contenido += "\ttable" + strconv.Itoa(int(contador)) + "[label=<\n"
			contenido += "\t\t<TABLE BORDER=\"0\" CELLBORDER=\"1\" CELLSPACING=\"0\">\n"
			contenido += "\t\t\t<TR><TD bgcolor=\"purple\" COLSPAN=\"2\">Inodo " + strconv.Itoa(int(contador)) + "</TD></TR>\n"
			contenido += "\t\t\t<TR><TD> i_uid </TD><TD>" + strconv.Itoa(int(tablaInodo.I_uid)) + "</TD></TR>\n"
			contenido += "\t\t\t<TR><TD> i_gid </TD><TD>" + strconv.Itoa(int(tablaInodo.I_gid)) + "</TD></TR>\n"
			contenido += "\t\t\t<TR><TD> i_size </TD><TD>" + strconv.Itoa(int(tablaInodo.I_size)) + "</TD></TR>\n"
			contenido += "\t\t\t<TR><TD> i_atime </TD><TD>" + string(TrimArray(tablaInodo.I_atime[:])) + "</TD></TR>\n"
			contenido += "\t\t\t<TR><TD> i_ctime </TD><TD>" + string(TrimArray(tablaInodo.I_ctime[:])) + "</TD></TR>\n"
			contenido += "\t\t\t<TR><TD> i_mtime </TD><TD>" + string(TrimArray(tablaInodo.I_mtime[:])) + "</TD></TR>\n"
			for i := 0; i < 15; i++ {
				contenido += "\t\t\t<TR><TD> i_block[" + strconv.Itoa(i+1) + "]</TD><TD>" + strconv.Itoa(int(tablaInodo.I_block[i])) + "</TD></TR>\n"
			}
			contenido += "\t\t\t<TR><TD> i_type </TD><TD>"
			contenido += string(tablaInodo.I_type)
			contenido += "</TD></TR>\n"
			contenido += "\t\t\t<TR><TD> i_perm </TD><TD>" + strconv.Itoa(int(tablaInodo.I_perm)) + "</TD></TR>\n"
			contenido += "\t\t</TABLE>\n"
			contenido += "\t>]\n"
		}
		contador++
	}
	contenido += "}\n"
	directory := path + ".dot"
	// hay que crear los directorios el archivo nuevo
	comandos.MkDirectory(directory)
	comandos.Fopen(directory, contenido)
	// falta mandar el comando para convertirlo en pdf
	cmd := exec.Command("dot", directory, "-Tpdf", "-o", path+".pdf")
	err := cmd.Run()
	if err != nil {
		consola.AddToConsole(fmt.Sprintf("Error reporte inodo: %s\n", err.Error()))
		return
	}
}

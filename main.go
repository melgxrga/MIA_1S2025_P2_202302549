package main

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	commands "github.com/melgxrga/proyecto1Archivos/commands"
	analyzer "github.com/melgxrga/proyecto1Archivos/analizador"
	"github.com/melgxrga/proyecto1Archivos/consola"
	"net/http"
	"io/ioutil"
	"strings"
	"path/filepath"
)

func main() {
	// Crear un router Gin
	router := gin.Default()

	// Configurar CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Endpoint para analizar comandos
	router.POST("/analyze", func(c *gin.Context) {
		var json struct {
			Command string `json:"command" binding:"required"`
		}

		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Comando es requerido"})
			return
		}

		// Llamar a la función Analyzer del paquete analyzer para analizar el comando ingresado
		an := analyzer.Analyzer{}
		output, err := an.Analyzer(json.Command)

		if err != nil {
			// Si hay un error al analizar el comando, devolver el error
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Devolver una respuesta de éxito con el output del análisis
		c.JSON(http.StatusOK, gin.H{"message": "Comando analizado exitosamente", "output": output})
	})

	// Nuevo endpoint para obtener la salida de la consola
	router.GET("/getConsole", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"console": consola.GetConsole()})
	})

	// Nuevo endpoint para obtener particiones de un disco específico
	// GET /partitions?disk=<ruta>
	router.GET("/partitions", func(c *gin.Context) {
		disk := c.Query("disk")
		fmt.Println("[DEBUG] /partitions recibido disk:", disk)
		if disk == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Falta el parámetro disk"})
			return
		}
		mbr := commands.GetMBR(disk)
		fmt.Printf("[DEBUG] MBR de %s: %+v\n", disk, mbr)
		partitions := []struct {
			Status   byte   `json:"status"`
			Type     byte   `json:"type"`
			Fit      byte   `json:"fit"`
			Start    int64  `json:"start"`
			Size     int64  `json:"size"`
			Name     string `json:"name"`
		}{}
		for _, p := range mbr.Mbr_partitions {
			name := strings.TrimRight(string(p.Part_name[:]), "\x00")
			if name != "" && p.Part_status == '1' {
				partitions = append(partitions, struct {
					Status   byte   `json:"status"`
					Type     byte   `json:"type"`
					Fit      byte   `json:"fit"`
					Start    int64  `json:"start"`
					Size     int64  `json:"size"`
					Name     string `json:"name"`
				}{
					Status: p.Part_status,
					Type:   p.Part_type,
					Fit:    p.Part_fit,
					Start:  p.Part_start,
					Size:   p.Part_size,
					Name:   name,
				})
			}
		}
		fmt.Printf("[DEBUG] Particiones encontradas: %+v\n", partitions)
		c.JSON(http.StatusOK, partitions)
	})

	// Endpoint para obtener información de discos .smia en una carpeta
	// GET /disks?folder=<ruta>
	router.GET("/disks", func(c *gin.Context) {
		folder := c.Query("folder")
		if folder == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Falta el parámetro folder"})
			return
		}

		files, err := ioutil.ReadDir(folder)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo leer la carpeta"})
			return
		}

		type PartitionInfo struct {
			Status   byte   `json:"status"`
			Type     byte   `json:"type"`
			Fit      byte   `json:"fit"`
			Start    int64  `json:"start"`
			Size     int64  `json:"size"`
			Name     string `json:"name"`
		}
		type DiskInfo struct {
			Path          string          `json:"path"`
			Size          int64           `json:"size"`
			CreationDate  string          `json:"creation_date"`
			Signature     int64           `json:"signature"`
			Fit           string          `json:"fit"`
			Partitions    []PartitionInfo `json:"partitions"`
		}

		var disks []DiskInfo
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".mia") {
				fullPath := filepath.Join(folder, file.Name())
				mbr := commands.GetMBR(fullPath)
				partitions := []PartitionInfo{}
				for _, p := range mbr.Mbr_partitions {
					name := strings.TrimRight(string(p.Part_name[:]), "\x00")
					if name != "" && p.Part_status == '1' {
						partitions = append(partitions, PartitionInfo{
							Status: p.Part_status,
							Type:   p.Part_type,
							Fit:    p.Part_fit,
							Start:  p.Part_start,
							Size:   p.Part_size,
							Name:   name,
						})
					}
				}
				disks = append(disks, DiskInfo{
					Path:         fullPath,
					Size:         mbr.Mbr_tamano,
					CreationDate: strings.TrimRight(string(mbr.Mbr_fecha_creacion[:]), "\x00"),
					Signature:    mbr.Mbr_dsk_signature,
					Fit:          string(mbr.Dsk_fit),
					Partitions:   partitions,
				})
			}
		}
		c.JSON(http.StatusOK, disks)
	})

	// Iniciar el servidor en el puerto 8080
	router.Run(":8080")
}

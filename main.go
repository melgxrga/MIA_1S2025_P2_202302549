package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	analyzer "github.com/melgxrga/proyecto1Archivos/analizador"
	"github.com/melgxrga/proyecto1Archivos/consola"
	"net/http"
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

	// Iniciar el servidor en el puerto 8080
	router.Run(":8080")
}

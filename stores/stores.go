package stores

import (
	"fmt"
	"errors"
)

// Carnet de estudiante
const Carnet string = "49"

// Lista de letras disponibles (A-Z)
var alphabet = []string{
	"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M",
	"N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z",
}

// Mapa para almacenar la asignación de letras a los diferentes paths
var pathToLetter = make(map[string]string)

// Mapa para almacenar el contador de particiones por path
var pathToPartitionCount = make(map[string]int)

// Índice para la siguiente letra disponible en el abecedario
var nextLetterIndex = 0

// Función para obtener la letra asignada a un path y el siguiente índice de partición
func GetLetterAndPartitionCorrelative(path string) (string, int, error) {
	// Asignar una letra al path si no tiene una asignada
	if _, exists := pathToLetter[path]; !exists {
		if nextLetterIndex < len(alphabet) {
			pathToLetter[path] = alphabet[nextLetterIndex]
			pathToPartitionCount[path] = 0 // Inicializar el contador de particiones
			nextLetterIndex++
		} else {
			return "", 0, errors.New("no hay más letras disponibles para asignar")
		}
	}

	// Incrementar y obtener el siguiente índice de partición para este path
	pathToPartitionCount[path]++
	nextIndex := pathToPartitionCount[path]

	return pathToLetter[path], nextIndex, nil
}

// Generar el ID de partición de manera directa
func GeneratePartitionID(path string) (string, int, error) {
	// Obtener la letra y el número correlativo de partición
	letter, partitionCorrelative, err := GetLetterAndPartitionCorrelative(path)
	if err != nil {
		return "", 0, err
	}

	// Crear el ID de partición, combinando el carnet y los valores obtenidos
	idPartition := fmt.Sprintf("%s%d%s", Carnet, partitionCorrelative, letter)

	return idPartition, partitionCorrelative, nil
}

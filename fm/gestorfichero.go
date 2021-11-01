package fm // file manager

import (
	"bufio"
	"fmt"
	"os"
)

const PATH = "file.txt"

func checkError(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func LeerFichero() string {
	// Abre el fichero en modo solo lectura
	file, err := os.OpenFile("file.txt", os.O_APPEND|os.O_RDONLY, 0644)
	checkError(err)
	defer file.Close()

	// Crea el reader
	reader := bufio.NewReader(file)

	var read string
	var readError error = nil
	var str string

	// Leemos todas las lineas
	for readError == nil {
		read += str
		str, readError = reader.ReadString('\n')
	}

	return read
}

func EscribirFichero(datos string) {
	// Abrimos el fichero en lectura-escritura
	// Lectura:   Obtener el contenido del fichero
	// Escritura: Escribir los datos enviados
	file, err := os.OpenFile("file.txt", os.O_APPEND|os.O_RDWR, 0644)
	checkError(err)
	defer file.Close()

	// Creamos el writer
	writter := bufio.NewWriter(file)
	_, err = writter.WriteString(datos)
	checkError(err)

	writter.Flush()
}

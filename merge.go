package main

import (
	"fmt"
	"io"
	"os"
)

func checkError(err error) {
	if err != nil {
		fmt.Println(err.Error())
	}
}

// PRE: [path fichero a unir]
func main() {
	args := os.Args[1:]

	// Abrimos el fichero principal
	f1, err := os.OpenFile("Comunicacion.txt", os.O_APPEND|os.O_RDWR, 0644)
	checkError(err)
	defer f1.Close()

	// Abrimos el fichero secundario
	f2, err := os.Open(args[0])
	checkError(err)
	defer f2.Close()

	// Copiamos en el fichero principal el contenido del secundario
	_, err = io.Copy(f1, f2)
	checkError(err)
}

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

func main() {
	args := os.Args[1:]

	f1, err := os.OpenFile("Comunicacion.txt", os.O_APPEND|os.O_RDWR, 0644)
	checkError(err)
	defer f1.Close()

	f2, err := os.Open(args[0])
	checkError(err)
	defer f2.Close()

	_, err = io.Copy(f1, f2)
	checkError(err)
}

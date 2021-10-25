package main

import (
	"main/ra"
	"os"
	"strconv"
	"time"
)

const ACCION = "escribir"
const ITERACIONES = 20

func escribir(ra *ra.RASharedDB, done chan bool) {
	for i := 0; i < ITERACIONES; i++ {
		println("Escritor - Preprotocolo")
		ra.PreProtocol()
		println("Escritor - Escribe")
		time.Sleep(2 * time.Second) //escribirFichero()
		println("Escritor - Postprotocolo")
		ra.PostProtocol()
	}
}

//PRE: [ID, PathFichero]
func main() {
	args := os.Args[1:]
	me, _ := strconv.Atoi(args[0])

	ra := ra.New(me, args[1], ACCION)
	done := make(chan bool)
	go escribir(ra, done)

	<-done
}

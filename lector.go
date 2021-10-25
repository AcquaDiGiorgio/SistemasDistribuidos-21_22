package main

import (
	"main/ra"
	"os"
	"strconv"
	"time"
)

const ACCION = "leer"
const ITERACIONES = 20

func leer(ra *ra.RASharedDB, done chan bool) {
	for i := 0; i < ITERACIONES; i++ {
		println("Lector - Preprotocolo")
		ra.PreProtocol()
		println("Lector - Escribe")
		time.Sleep(1 * time.Second) //leerFichero()
		println("Lector - Postprotocolo")
		ra.PostProtocol()
	}
}

//PRE: [ID, PathFichero]
func main() {
	args := os.Args[1:]
	me, _ := strconv.Atoi(args[0])

	ra := ra.New(me, args[1], ACCION)
	done := make(chan bool)
	go leer(ra, done)

	<-done
}

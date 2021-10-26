package main

import (
	"main/ra"
	"os"
	"strconv"
)

const ACTOR = "lector"
const ITERACIONES = 20

func leer(ra *ra.RASharedDB) {
	for i := 0; i < ITERACIONES; i++ {
		ra.PreProtocol()
		ra.AccedoSC("")
		ra.PostProtocol()
	}
}

//PRE: [ID, PathFichero]
func main() {
	args := os.Args[1:]
	me, _ := strconv.Atoi(args[0])

	ra := ra.New(me, args[1], ACTOR)
	leer(ra)
}

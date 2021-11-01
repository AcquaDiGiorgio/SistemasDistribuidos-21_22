package main

import (
	"main/fm"
	"main/ra"
	"os"
	"strconv"

	"github.com/DistributedClocks/GoVector/govec"
)

const ACTOR = "escritor"
const ITERACIONES = 20

func escribir(ra *ra.RASharedDB, logger *govec.GoLog) {
	for i := 0; i < ITERACIONES; i++ {
		ra.PreProtocol()

		logger.LogLocalEvent("Escribo en el Fichero", govec.GetDefaultLogOptions())
		fm.EscribirFichero("Escritor-" + strconv.Itoa(ra.Me) + " [*] Cadena no: " + strconv.Itoa(i) + "\n")

		ra.PostProtocol()
	}
}

//PRE: [ID, PathFichero]
func main() {
	args := os.Args[1:]
	me, _ := strconv.Atoi(args[0])

	logger := govec.InitGoVector(ACTOR+"-"+strconv.Itoa(me), "LOG_"+strconv.Itoa(me), govec.GetDefaultConfig())
	ra := ra.New(me, args[1], ACTOR, logger)
	defer ra.Stop()

	escribir(ra, logger)

}

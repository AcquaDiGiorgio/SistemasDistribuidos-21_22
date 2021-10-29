package main

import (
	"main/fm"
	"main/ra"

	"github.com/DistributedClocks/GoVector/govec"
)

const ACTOR = "escritor"
const ITERACIONES = 20

func escribir(ra *ra.RASharedDB, logger *govec.GoLog) {
	for i := 0; i < ITERACIONES; i++ {
		ra.PreProtocol()

		fm.LeerFichero()

		ra.PostProtocol()
	}
}

//PRE: [ID, PathFichero]
func main() {
	//args := os.Args[1:]
	//me, _ := strconv.Atoi(args[0])

	logger := govec.InitGoVector("Escritor", "LogFile", govec.GetDefaultConfig())
	logger.LogLocalEvent("", govec.GetDefaultLogOptions())

	//ra := ra.New(me, args[1], ACTOR, logger)

	//escribir(ra, logger)
}

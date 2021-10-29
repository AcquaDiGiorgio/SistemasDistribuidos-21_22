package main

import (
	"main/fm"
	"main/ra"

	"github.com/DistributedClocks/GoVector/govec"
)

const ACTOR = "lector"
const ITERACIONES = 20

func leer(ra *ra.RASharedDB, logger *govec.GoLog) {
	for i := 0; i < ITERACIONES; i++ {
		ra.PreProtocol()

		fm.EscribirFichero()

		ra.PostProtocol()
	}
}

//PRE: [ID, PathFichero]
func main() {
	//args := os.Args[1:]
	//me, _ := strconv.Atoi(args[0])
	logger := govec.InitGoVector("lector", "LogFile", govec.GetDefaultConfig())
	logger.LogLocalEvent("arh", govec.GetDefaultLogOptions())

	//ra := ra.New(me, args[1], ACTOR, logger)

	//leer(ra, logger)
}

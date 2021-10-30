package main

import (
	"fmt"
	"main/fm"
	"main/ra"
	"os"
	"strconv"

	"github.com/DistributedClocks/GoVector/govec"
)

const ACTOR = "lector"
const ITERACIONES = 20

func leer(ra *ra.RASharedDB, logger *govec.GoLog) {
	for i := 0; i < ITERACIONES; i++ {
		ra.PreProtocol()

		logger.LogLocalEvent("Leo del Fichero", govec.GetDefaultLogOptions())
		fmt.Println(fm.LeerFichero())

		ra.PostProtocol()
	}
}

//PRE: [ID, PathFichero]
func main() {
	args := os.Args[1:]
	me, _ := strconv.Atoi(args[0])

	logger := govec.InitGoVector(ACTOR+"-"+strconv.Itoa(me), "LOG_"+strconv.Itoa(me), govec.GetDefaultConfig())
	ra := ra.New(me, args[1], ACTOR, logger)

	leer(ra, logger)
}

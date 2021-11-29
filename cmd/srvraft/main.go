package main

import (
	"fmt"
	"os"
	"raft/internal/raft"
	"strconv"
	"time"
)

func main() {
	args := os.Args[1:]
	nodo, _ := strconv.Atoi(args[0])

	nr := raft.NuevoNodo(nodo, nil)
	fmt.Println("============= Nodo Creado ==============")

	for {
		var estado raft.Estado
		nr.ObtenerEstado(nil, &estado)
		fmt.Println("Soy el Nodo: ", estado.Yo)
		fmt.Println("Estamos en el Mandato: ", estado.Mandato)

		if estado.EsLider {
			fmt.Println("Soy el Master Actual")
			fmt.Println("##########################################")
			fmt.Print("Introduce Operacion: ")

			var op string
			fmt.Scanln(&op)

			//var OpASometer raft.OpASometer
			//nr.SometerOperacion(&op, &OpASometer)
			fmt.Println("==========================================")

		} else {
			fmt.Println("NO soy el Master Actual")
			fmt.Println("==========================================")
			time.Sleep(5 * time.Second)
		}
	}
}

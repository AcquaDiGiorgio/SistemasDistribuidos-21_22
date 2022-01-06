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

	if len(args) != 1 {
		fmt.Println("Debe se ejecutado con los argumentos <idNodo>")
		os.Exit(1)
	}

	nodo, err := strconv.Atoi(args[0])
	if err != nil {
		panic(err)
	}

	nr := raft.NuevoNodo(nodo)

	fmt.Print("\n")
	fmt.Printf("Nodo %d Creado\n", nodo)
	time.Sleep(1 * time.Second)

	for {
		yo, lider := mostrarInfoNodo(nr)

		if lider == -1 {
			time.Sleep(1 * time.Second)

		} else if yo == lider {
			fmt.Print("Introduce Operacion: ")

			var op string
			fmt.Scanln(&op)

			if op == "Stop" {
				desconectar(nr)
				time.Sleep(2 * time.Second)

			} else {
				var OpASometer raft.OpASometer
				nr.SometerOperacion(op, &OpASometer)
			}

		} else {
			fmt.Print("Desconectar Nodo [y|otherValue]: ")

			var disconn string
			fmt.Scanln(&disconn)

			if disconn == "y" {
				desconectar(nr)
			}
		}

		time.Sleep(1 * time.Second)
	}
}

func mostrarInfoNodo(nr *raft.NodoRaft) (int, int) {
	var empty raft.EmptyValue
	var estado raft.Estado

	nr.ObtenerEstado(empty, &estado)

	fmt.Print("\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n")
	fmt.Printf("================ Nodo %d =================\n", estado.Yo)

	if estado.CandidaturaActual == 0 {
		fmt.Println("Se acaba de iniciar el sistema")
		fmt.Println("==========================================")
	} else {
		fmt.Println("Estamos en el Mandato: ", estado.CandidaturaActual)
		if estado.Yo == estado.LiderActual {
			fmt.Println("Soy el líder Actual")
		} else {
			if estado.EstamosEnCandidatura {
				fmt.Println("No hay líder, se está buscando uno")
			} else {
				fmt.Printf("El líder actual es %d\n", estado.LiderActual)
			}
		}
		fmt.Println("==========================================")
		mostrarEntradas(estado.Entradas, estado.UltimaEntrada, estado.UltimaEntradaComprometida)
	}

	return estado.Yo, estado.LiderActual
}

func mostrarEntradas(entradas []string, ultimaEntrada int, ultimaComprometida int) {
	fmt.Println()
	fmt.Println("$$$$$$$$$$$$$$$ OPERACIONES $$$$$$$$$$$$$$")
	for i := 0; i <= ultimaEntrada; i++ {
		fmt.Print("Entrada", i, "=>", entradas[i])
		if i == ultimaComprometida {
			fmt.Print("\t<- Ultima Comprometida")
		}
		fmt.Println()
	}
	fmt.Println("$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")
	fmt.Println()
}

func desconectar(nr *raft.NodoRaft) {
	var empty raft.EmptyValue

	nr.Para(empty, &empty)
	fmt.Println("Se ha desconectado el nodo")

	fmt.Print("Reconectar al sistema?: ")

	var reconect string
	fmt.Scanln(&reconect)
	fmt.Println()

	err := nr.Reconectar(empty, &empty)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("XD")
}

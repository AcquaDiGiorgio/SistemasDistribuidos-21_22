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

	if len(args) != 2 {
		fmt.Println("Debe se ejecutado con los argumentos <idNodo> <opciónDebug>")
		fmt.Println("OpciónDebug:")
		fmt.Println("\t0: No debug + Entrada de usuario")
		fmt.Println("\t1: XXXXXXXXXXXXXX")
		fmt.Println("\t2: XXXXXXXXXXXXXX")
		fmt.Println("\t3: XXXXXXXXXXXXXX")
		fmt.Println("\t4: XXXXXXXXXXXXXX")
		os.Exit(1)
	}

	nodo, err := strconv.Atoi(args[0])
	if err != nil {
		panic(err)
	}

	opt, err := strconv.Atoi(args[1])

	if err != nil {
		panic(err)
	}

	switch opt {
	case 0:
		noDebug(nodo)

	case 1:
		//IniciarYParar(nodo)

	case 2:
		//IniciarYElegirLider(nodo)

	case 3:
		//IniciarYTumbarLider(nodo)

	case 4:
		//IniciarYComprometer3Entradas(nodo)

	default:
		fmt.Println("Opción no válida")
	}
}

func noDebug(nodo int) {
	nr := raft.NuevoNodo(nodo, nil)

	fmt.Print("\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n")
	fmt.Println("============== Nodo Creado ===============")
	var empty raft.EmptyValue

	for {
		var estado raft.Estado

		nr.ObtenerEstado(empty, &estado)
		fmt.Println("Soy el Nodo: ", estado.Yo)
		fmt.Println("Estamos en el Mandato: ", estado.CandidaturaActual)

		if estado.EsLider {
			fmt.Println("Soy el Master Actual")
			fmt.Println("==========================================")

			mostrarEntradas(estado.Entradas, estado.UltimaEntrada, estado.UltimaEntradaComprometida)

			fmt.Print("Introduce Operacion: ")

			var op string
			fmt.Scanln(&op)

			if op == "Stop" {
				nr.Para(empty, &empty)
				os.Exit(0)
			}

			var OpASometer raft.OpASometer
			nr.SometerOperacion(op, &OpASometer)

		} else {
			if estado.EstamosEnCandidatura {
				fmt.Println("No hay master, se está buscando uno")
			} else {
				fmt.Printf("El master actual es %d\n", estado.MasterActual)
			}

			fmt.Println("==========================================")

			mostrarEntradas(estado.Entradas, estado.UltimaEntrada, estado.UltimaEntradaComprometida)

		}

		fmt.Print("\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n")
		time.Sleep(1 * time.Second)
		fmt.Println("==========================================")
	}
}

func mostrarEntradas(entradas []string, ultimaEntrada int, ultimaComprometida int) {
	fmt.Println()
	fmt.Println("$$$$$$$$$$$$$$$ OPERACIONES $$$$$$$$$$$$$$")
	for i := 0; i <= ultimaEntrada; i++ {
		fmt.Print("Entrada", i, "=>", entradas[i])
		if i == ultimaComprometida {
			fmt.Print("\t\t<- Ultima Comprometida")
		}
		fmt.Println()
	}
	fmt.Println("$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")
	fmt.Println()
}

func IniciarYComprometer3Entradas(nodo int) {
	var empty raft.EmptyValue
	idOp := 0
	nr := raft.NuevoNodo(nodo, nil)
	fmt.Printf("Nodo %d Creado\n", nodo)
	for {
		var estado raft.Estado
		nr.ObtenerEstado(empty, &estado)

		// Me he convertido en líder del mandato 0
		if estado.EsLider {
			fmt.Printf("Nodo %d Somete Operación\n", nodo)
			operacion := "Operacion" + strconv.Itoa(idOp)
			idOp++

			if idOp == 3 {
				fmt.Printf("Nodo %d Termina\n", nodo)
				nr.Para(empty, &empty)
				return
			}

			var OpASometer raft.OpASometer
			nr.SometerOperacion(operacion, &OpASometer)
			time.Sleep(100 * time.Millisecond)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

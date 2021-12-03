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
		fmt.Println("\t1: Iniciar y parar el nodo")
		fmt.Println("\t2: Iniciar y elegir líder")
		fmt.Println("\t3: Inciar, tumbar si soy lider y hacer elecciones")
		fmt.Println("\t4: Comprometer 3 entradas si soy lider")
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
		IniciarYParar(nodo)

	case 2:
		IniciarYElegirLider(nodo)

	case 3:
		IniciarYTumbarLider(nodo)

	case 4:
		IniciarYComprometer3Entradas(nodo)

	default:
		fmt.Println("Opción no válida")
	}
}

func noDebug(nodo int) {
	nr := raft.NuevoNodo(nodo, nil)
	fmt.Println("============= Nodo Creado ==============")
	var empty raft.EmptyValue

	for {
		var estado raft.Estado

		nr.ObtenerEstado(empty, &estado)
		fmt.Println("Soy el Nodo: ", estado.Yo)
		fmt.Println("Estamos en el Mandato: ", estado.Mandato)

		if estado.EsLider {
			fmt.Println("Soy el Master Actual")
			fmt.Println("##########################################")
			fmt.Print("Introduce Operacion: ")

			var op string
			fmt.Scanln(&op)

			if op == "Stop" {
				nr.Para(empty, &empty)
				os.Exit(0)
				return
			}

			var OpASometer raft.OpASometer
			nr.SometerOperacion(op, &OpASometer)
			fmt.Println("==========================================")

		} else {
			fmt.Println("NO soy el Master Actual")
			fmt.Println("==========================================")
			time.Sleep(5 * time.Second)
		}
	}
}

func IniciarYParar(nodo int) {
	var empty raft.EmptyValue
	nr := raft.NuevoNodo(nodo, nil)
	fmt.Printf("Nodo %d Creado\n", nodo)
	time.Sleep(3 * time.Second)
	fmt.Printf("Nodo %d Parado\n", nodo)
	nr.Para(empty, &empty)
}

func IniciarYElegirLider(nodo int) {
	var empty raft.EmptyValue
	nr := raft.NuevoNodo(nodo, nil)
	fmt.Printf("Nodo %d Creado\n", nodo)
	time.Sleep(1 * time.Second)
	for {
		var estado raft.Estado
		nr.ObtenerEstado(empty, &estado)

		// Me he convertido en líder o alguien lo ha hecho
		if estado.Mandato > 0 {
			nr.Para(empty, &empty)
			return

		} else if estado.EsLider {
			fmt.Printf("Nodo %d Se ha convertido en Líder del mandato %d\n", estado.Yo, estado.Mandato)
			nr.Para(empty, &empty)
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func IniciarYTumbarLider(nodo int) {
	var empty raft.EmptyValue
	nr := raft.NuevoNodo(nodo, nil)
	fmt.Printf("Nodo %d Creado\n", nodo)
	for {
		var estado raft.Estado

		nr.ObtenerEstado(empty, &estado)

		// Me he convertido en líder del mandato 0
		if estado.EsLider && estado.Mandato == 0 {
			fmt.Printf("Nodo %d se ha hecho Líder en el mandato 0\n", nodo)
			fmt.Printf("Nodo %d Tumbado\n", nodo)
			nr.Para(empty, &empty)
			return

			// Me he convertido en líder del mandato 1
		} else if estado.EsLider && estado.Mandato == 1 {
			fmt.Printf("Nodo %d se ha hecho Líder en el mandato 1\n", nodo)
			nr.Para(empty, &empty)
			return

			// Ha caído el master del mandato 1
		} else if estado.Mandato == 2 {
			nr.Para(empty, &empty)
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
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

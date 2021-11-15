package main

import (
	"fmt"
	"log"
	"main/com"
	"math"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"syscall"

	"golang.org/x/term"
)

const (
	// Estado actual del Sistema
	SUFIC_WORKERS  = 0
	POCOS_WORKERS  = -1
	MUCHOS_WORKERS = 1

	// Otras constantes
	LIM_SUP_THR = 2.91
	LIM_INF_THR = 1.49
	THR_PEOR    = 1.49
	ipCoord     = ""
)

type Estado struct {
	// Info del estado actual del sistema
	actual_thoughput float64 // En milisegundos
	mutex            sync.Mutex
	estadoWorker     []bool
	workersActivos   int

	// Info del usuario para lanzar workers
	user string
	pass string
}

/*
	FUNCIONES RPC
*/

func (e *Estado) LanzarWorker(id int, levantado *bool) {
	*levantado = false

	//Creamos el ssh hacia la máquina en la que se encuentra el worker
	ssh, err := com.NewSshClient(
		e.user,
		com.Workers[id].Host,
		22,
		RSA,
		e.pass)
	if err != nil {
		log.Printf("SSH init error %v", err)
		os.Exit(1)
	}

	err = ssh.RunCommand(PATH + "worker " + com.Workers[id].Ip)

	if err != nil {
		e.mutex.Lock()
		e.estadoWorker[id] = true
		e.workersActivos++
		e.mutex.Unlock()
		*levantado = true
	}
}

// Se acaba de introducir un dato
func (e *Estado) NuevaEntrada(interval com.TPInterval, noReturn interface{}) {
	e.mutex.Lock()
	e.actual_thoughput += aproxThr(interval)
	estado := e.checkWorkers()
	e.mutex.Unlock()

	if estado == POCOS_WORKERS {
		e.relanzarWorker()
	} else if estado == MUCHOS_WORKERS {
		e.terminarWorker()
	}
}

// Al salir se comprueba si hay que terminar algún worker
func (e *Estado) NuevaSalida(interval com.TPInterval, noReturn interface{}) {
	e.mutex.Lock()
	e.actual_thoughput -= aproxThr(interval)
	estado := e.checkWorkers()
	e.mutex.Unlock()

	if estado == POCOS_WORKERS {
		e.relanzarWorker()
	} else if estado == MUCHOS_WORKERS {
		e.terminarWorker()
	}
}

// Devuelve el estado acutal del worker
// Preparado / No preparado para recibir tareas
func (e *Estado) PedirWorker(id int, accesible *bool) {
	e.mutex.Lock()
	*accesible = e.estadoWorker[id]
	e.mutex.Unlock()
}

// retVal = worker Iniciado
func (e *Estado) InformarWorkerCaido(id int, workIniciado *bool) {
	e.mutex.Lock()
	e.workersActivos--
	e.mutex.Unlock()

	*workIniciado = false

	if e.checkWorkers() == POCOS_WORKERS {
		e.LanzarWorker(id, workIniciado)
	}
}

/*
	FUNCIONES INTERNAS
*/

// Checkeamos el estado actual del sistema
// Si hay más workers de los necesarios, devuelve MUCHOS_WORKERS
// Si hay menos workers de los necesarios, devuelve POCOS_WORKERS
func (e *Estado) checkWorkers() int {
	retVal := SUFIC_WORKERS
	// Si no podemos meter una tarea de máximo coste introducimos un worker
	if e.actual_thoughput+THR_PEOR > float64(e.workersActivos)*THR_PEOR {
		retVal = POCOS_WORKERS

		// Si podemos meter +2 tareas de máximo coste terminamos un worker
	} else if e.actual_thoughput+2*THR_PEOR < float64(e.workersActivos)*THR_PEOR {
		retVal = MUCHOS_WORKERS
	}
	return retVal
}

// Relanzamos el worker con menor id
func (e *Estado) relanzarWorker() {
	done := false
	e.mutex.Lock()
	for i := 0; i < com.POOL || done; i++ {
		if !e.estadoWorker[i] {
			e.estadoWorker[i] = true
			done = true
		}
	}
	if done {
		e.workersActivos++
	}
	e.mutex.Unlock()
}

// Terminamos el worker con mayor id
func (e *Estado) terminarWorker() {
	done := false
	e.mutex.Lock()
	for i := com.POOL - 1; i >= 0 || done; i-- {
		if e.estadoWorker[i] {
			e.estadoWorker[i] = false
			done = true
		}
	}
	if done {
		e.workersActivos--
	}
	e.mutex.Unlock()
}

func aproxThr(interval com.TPInterval) float64 {

	retVal := 0.0
	for j := interval.A; j <= interval.B; j += 1000 {
		retVal += 0.00164 * math.Pow(float64(j), 0.9055)
	}

	return retVal
}

func checkErrorCoord(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func main() {

	e := new(Estado)

	fmt.Print("Introduzca el usuario: ")
	fmt.Scanf("%s", &e.user)

	fmt.Print("Introduzca la Contraseña: ")
	pass, err := term.ReadPassword(int(syscall.Stdin))
	checkErrorCoord(err)

	e.pass = string(pass)

	for i := 0; i < com.POOL; i++ {
		e.estadoWorker[i] = false
	}

	// Registro y Creación del RPC
	rpc.Register(e)
	rpc.HandleHTTP()

	// Inicio Escucha
	listener, err := net.Listen("tcp", ipCoord)
	checkErrorMaster(err)
	defer listener.Close()

	// Sirve petiticiones
	http.Serve(listener, nil)
}

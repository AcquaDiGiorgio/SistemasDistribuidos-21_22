package main

import (
	"fmt"
	"log"
	"main/com"
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
	POCOS_WORKERS  = -1
	MUCHOS_WORKERS = 1

	// Otras constantes
	LIM_SUP_THR = 2.8
	LIM_INF_THR = 1.4
	ipCoord     = ""
)

type Estado struct {
	// Info del estado actual del sistema
	actual_thoughput int
	mutex            sync.Mutex
	estadoWorker     []bool

	// Info del usuario para lanzar workers
	user string
	pass string
}

/*
	FUNCIONES RPC
*/

func (e *Estado) LanzarWorker(id int) {
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
	checkErrorCoord(err)

	e.mutex.Lock()
	e.estadoWorker[id] = true
	e.mutex.Unlock()
}

// Se acaba de introducir un dato
func (e *Estado) NuevaEntrada(interval com.TPInterval) {
	e.mutex.Lock()
	e.actual_thoughput += aproxThr(interval)
	e.mutex.Unlock()
}

// Al salir se comprueba si hay que terminar algún worker
func (e *Estado) NuevaSalida(interval com.TPInterval) {
	e.mutex.Lock()
	e.actual_thoughput -= aproxThr(interval)
	e.mutex.Unlock()

	systemCapability := e.checkWorkers()

	switch systemCapability {
	case POCOS_WORKERS:
		e.relanzarWorker()
		break

	case MUCHOS_WORKERS:
		e.terminarWorker()
		break

	default:
		break
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

	systemCapability := e.checkWorkers()
	switch systemCapability {

	// Si hay pocos workers lo lanzamos
	case POCOS_WORKERS:
		e.LanzarWorker(id)
		*workIniciado = true
		break

	// Si hay muchos o suficientes workers no lo lanzamos
	default:
		*workIniciado = false
		break
	}
}

/*
	FUNCIONES INTERNAS
*/

// Checkeamos el estado actual del sistema
// Si hay más workers de los necesarios, devuelve MUCHOS_WORKERS
// Si hay menos workers de los necesarios, devuelve POCOS_WORKERS
func (e *Estado) checkWorkers() int {
	var estadoWorkers int
	// Calculo del estado
	return estadoWorkers
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
	e.mutex.Unlock()
}

func aproxThr(interval com.TPInterval) int {
	var calc int
	// Calculo aproximado del coste
	return calc
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

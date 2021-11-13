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

func (e *Estado) NuevaEntrada(id int, interval com.TPInterval) {
	e.mutex.Lock()
	e.actual_thoughput += aproxThr(interval)
	e.mutex.Unlock()

	systemCapability := e.checkWorkers()

	switch systemCapability {
	case POCOS_WORKERS:
		e.LanzarWorker()
		break

	case MUCHOS_WORKERS:
		e.terminarWorker()
		break

	default:
		break
	}

}

func (e *Estado) NuevaSalida(id int, interval com.TPInterval) {
	e.mutex.Lock()
	e.actual_thoughput -= aproxThr(interval)
	e.mutex.Unlock()

	systemCapability := e.checkWorkers()

	switch systemCapability {
	case POCOS_WORKERS:
		e.LanzarWorker()
		break

	case MUCHOS_WORKERS:
		e.terminarWorker()
		break

	default:
		break
	}
}

func (e *Estado) PedirWorker(id int) (accesible bool) {
	e.mutex.Lock()
	accesible = e.estadoWorker[id]
	e.mutex.Unlock()
	return
}

// retVal = worker Iniciado
func (e *Estado) InformarWorkerCaido(id int) (workIniciado bool) {

	systemCapability := e.checkWorkers()
	switch systemCapability {

	// Si hay suficientes workers no lo lanzamos
	case MUCHOS_WORKERS:
		workIniciado = false
		break

	default:
		e.LanzarWorker()
		workIniciado = true
		break
	}
	return
}

/*
	FUNCIONES INTERNAS
*/

func (e *Estado) checkWorkers() int {
	var estadoWorkers int
	// Calculo del estado
	return estadoWorkers
}

func (e *Estado) terminarWorker(id int) {
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

	err = ssh.RunCommand(PATH + "worker " + com.Workers[id].Ip) // Kill?
	checkErrorCoord(err)

	e.mutex.Lock()
	e.estadoWorker[id] = true
	e.mutex.Unlock()
}

func aproxThr(interval com.TPInterval) int {
	var calc int
	// Calculo aproximado del coste
	return calc
}

func checkErrorCoord(err error) {

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

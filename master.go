/*
* AUTORES: Jorge Lisa y David Zandundo
* ASIGNATURA: 30221 Sistemas Distribuidos del Grado en Ingeniería Informática
*			  Escuela de Ingeniería y Arquitectura - Universidad de Zaragoza
* FECHA: noviembre de 2021
* FICHERO: master.go
 */

package main

import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"time"

	"main/com"
)

const (
	PATH      = "/home/a774248/SSDD/Practica1/"
	RSA       = "/home/a774248/.ssh/id_rsa"
	CONN_TYPE = "tcp"
	MAX_TIME  = time.Duration(5 * time.Second)
)

//Struct usado para realizar el envío de mensajes por canal.
//Consta de un encoder, para devolver el dato y la petición del cliente.
type Respuesta struct {
	reply []int
	err   error
}

type Mensaje struct {
	intervalo com.TPInterval
	resp      chan Respuesta
}

type Primes struct {
	canal chan Mensaje
	coord *rpc.Client
}

func checkErrorMaster(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func trabajar(id int, primes *Primes, worker *rpc.Client, callChan chan *rpc.Call) {
	for {
		msj := <-primes.canal // Recibimos la petición del master
		fmt.Printf("INTERVALO %d -> %d RECIBIDO POR EL WORKER %d\n", msj.intervalo.A, msj.intervalo.B, id)

		var respuesta Respuesta
		// Enviamos al worker el trabajo
		worker.Go("PrimesImpl.FindPrimes", msj.intervalo, &respuesta.reply, callChan) //NIL POINTER DEREFERENCE
		select {
		case msg := <-callChan: // Recepción del mensaje a tiempo
			respuesta.err = msg.Error
			if msg.Error != nil {
				// Avisamos al coordinador que hemos terminado correctamente
				primes.coord.Go("Estado.NuevaSalida", msj.intervalo, nil, callChan)
				// Enviamos la respuesta al master
				msj.resp <- respuesta
			} else {
				msj.resp <- respuesta
				return
			}

			msj.resp <- respuesta
		case <-time.After(MAX_TIME): // Más retraso del permitido
			// Le pasamos el mensaje a otro worker
			primes.canal <- msj
			time.Sleep(1 * time.Second)
			break
		}
	}
}

//Gorutina capaz de lanzar por ssh un worker y esperar a que entre por el canal de mensajes
//una petición del cliente
//Esta función recibe el host del worker, su ip, el usuario que hace el ssh y su contraseña
func workerLanzar(worker int, primes *Primes) {

	for {
		callChan := make(chan *rpc.Call, 10)
		var accesoPermitido bool = false
		var work *rpc.Client
		var err error

		for !accesoPermitido {
			for !accesoPermitido {
				primes.coord.Call("Estado.PedirWorker", worker, &accesoPermitido)
				time.Sleep(3 * time.Second)
			}
			work, err = rpc.DialHTTP("tcp", com.Workers[worker].Ip)
			if err != nil {
				primes.coord.Call("Estado.InformarWorkerCaido", worker, &accesoPermitido)
			}
		}

		fmt.Printf("WORKER %d EN MARCHA\n", worker)
		trabajar(worker, primes, work, callChan)
	}
}

func (p *Primes) FindPrimes(interval com.TPInterval, primeList *[]int) error {
	fmt.Printf("RECIBIDO INTERVALO %d -> %d DEL CLIENTE\n", interval.A, interval.B)

	resp := make(chan Respuesta)
	callChan := make(chan *rpc.Call, 10)

	p.coord.Go("Estado.NuevaEntrada", interval, nil, callChan)

	fmt.Printf("ENVÍO EL INTERVALO %d -> %d AL CANAL\n", interval.A, interval.B)
	p.canal <- Mensaje{interval, resp} // Enviamos la petición a un worker
	respuesta := <-resp                // Esperamos a la respuesta del worker
	fmt.Printf("RECIBO RESPUESTA DEL CANAL\n")
	if respuesta.err != nil {
		*primeList = respuesta.reply // Devolvemos la respuesta
	}
	return respuesta.err
}

func main() {

	// Nos conectamos con el coordinador
	conn, err := rpc.DialHTTP("tcp", com.ENPOINT_COORD)
	checkErrorMaster(err)

	// Creamos un canal que pasa las tareas a las gorutines
	primes := new(Primes)
	primes.canal = make(chan Mensaje)
	primes.coord = conn

	var errSSH bool
	callChan := make(chan *rpc.Call, 10)
	// Llama por ssh a los workers y los prepara para escuchar
	for i := 0; i < com.POOL; i++ {
		primes.coord.Go("Estado.LanzarWorker", i, errSSH, callChan)
		go workerLanzar(i, primes)
	}

	// Registro y Creación del RPC
	rpc.Register(primes)
	rpc.HandleHTTP()

	// Inicio Escucha
	listener, err := net.Listen(CONN_TYPE, com.ENPOINT_MASTER)
	checkErrorMaster(err)
	defer listener.Close()

	// Sirve petiticiones
	http.Serve(listener, nil)
}

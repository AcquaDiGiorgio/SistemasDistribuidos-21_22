/*
* AUTOR: Rafael Tolosana Calasanz
* EDITADO: Jorge Lisa y David Zandundo
* ASIGNATURA: 30221 Sistemas Distribuidos del Grado en Ingeniería Informática
*			  Escuela de Ingeniería y Arquitectura - Universidad de Zaragoza
* FECHA: septiembre de 2021
* FICHERO: server.go
* DESCRIPCIÓN: contiene la funcionalidad esencial para realizar los servidores
*			   correspondientes al trabajo 1
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
	PATH        = "/home/a774248/SSDD/Practica1/"
	RSA         = "/home/a774248/.ssh/id_rsa"
	CONN_TYPE   = "tcp"
	CONN_HOST   = "155.210.154.200"
	COORDINADOR = ""
	MAX_TIME    = time.Duration(1 * time.Second)
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

//Gorutina capaz de lanzar por ssh un worker y esperar a que entre por el canal de mensajes
//una petición del cliente
//Esta función recibe el host del worker, su ip, el usuario que hace el ssh y su contraseña
func workerLanzar(worker int, primes *Primes) {

	for {
		callChan := make(chan *rpc.Call)
		var acceso bool

		work, err := rpc.DialHTTP("tcp", com.Workers[worker].Ip)

		for err != nil {
			primes.coord.Call("Estado.PedirWorker", worker, &acceso)
			if acceso {
				work, err = rpc.DialHTTP("tcp", com.Workers[worker].Host)
			} else {
				// Worker no accesible
				// Wait?
			}
		}

		for {
			msj := <-primes.canal // Recibimos la petición del master
			var respuesta Respuesta
			// Enviamos al worker el trabajo
			work.Go("PrimesImpl.FindPrimes", msj.intervalo, &respuesta.reply, callChan)
			select {
			case msg := <-callChan: // Recepción del mensaje a tiempo
				if msg.Error != nil {
					break
				}
			case <-time.After(MAX_TIME): // Más retraso del permitido
				// Avisamos al coordinador que hemos terminado erróneamente
				primes.coord.Go("Estado.NuevaSalida", com.Salida{worker, com.TPInterval{-1, -1}}, nil, callChan)
				// Le pasamos el mensaje a otro worker
				primes.canal <- msj
				break
			}
			// Avisamos al coordinador que hemos terminado correctamente
			primes.coord.Go("Estado.NuevaSalida", com.Salida{worker, msj.intervalo}, nil, callChan)
			// Enviamos la respuesta al master
			msj.resp <- respuesta
		}
	}
}

func (p *Primes) FindPrimes(interval com.TPInterval, primeList *[]int) error {
	resp := make(chan Respuesta)
	callChan := make(chan *rpc.Call)

	p.coord.Go("Estado.NuevaEntrada", interval, nil, callChan)

	p.canal <- Mensaje{interval, resp} // Enviamos la petición a un worker
	respuesta := <-resp                // Esperamos a la respuesta del worker
	*primeList = respuesta.reply       // Devolvemos la respuesta

	return respuesta.err
}

func main() {

	args := os.Args[1:]
	if len(args) != 1 {
		os.Exit(1)
	}

	// Nos conectamos con el coordinador
	conn, err := rpc.DialHTTP("tcp", COORDINADOR)

	// Creamos un canal que pasa las tareas a las gorutines
	primes := new(Primes)
	primes.canal = make(chan Mensaje)
	primes.coord = conn

	// Llama por ssh a los workers y los prepara para escuchar
	for i := 0; i < com.POOL; i++ {
		go workerLanzar(i, primes)
	}

	// Registro y Creación del RPC
	rpc.Register(primes)
	rpc.HandleHTTP()

	// Inicio Escucha
	listener, err := net.Listen(CONN_TYPE, CONN_HOST+":"+args[0])
	checkErrorMaster(err)
	defer listener.Close()

	// Sirve petiticiones
	http.Serve(listener, nil)
}

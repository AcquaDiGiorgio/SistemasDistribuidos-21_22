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

	"main/com"
)

const PATH = "/home/a774248/SSDD/Practica1/"
const RSA = "/home/a774248/.ssh/id_rsa"
const CONN_TYPE = "tcp"
const CONN_HOST = "155.210.154.200"

//Struct usado para realizar el envío de mensajes por canal.
//Consta de un encoder, para devolver el dato y la petición del cliente.
type Mensaje struct {
	intervalo com.TPInterval
	reply     *[]int
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
	callChan := make(chan *rpc.Call)
	var acceso bool

	client, err := rpc.DialHTTP("tcp", com.Workers[worker].Ip)

	for err != nil {
		primes.coord.Call("Estado.", worker, &acceso)
		if acceso {
			client, err = rpc.DialHTTP("tcp", com.Workers[worker].Host)
		} else {
			// Worker no accesible
			// Wait?
		}
	}

	for {
		msj := <-primes.canal
		primes.coord.Go("Estado.NuevaEntrada", com.Workers[worker], nil, callChan)
		client.Call()
		primes.coord.Go("Estado.NuevaSalida", com.Workers[worker], nil, callChan)
	}
}

type Primes struct {
	canal chan Mensaje
	coord *rpc.Client
}

const worker = ""

func (p *Primes) FindPrimes(interval com.TPInterval, primeList *[]int) error {
	p.canal <- Mensaje{interval, primeList}
	return nil
}

const coordinador = ""

func main() {

	args := os.Args[1:]
	if len(args) != 1 {
		os.Exit(1)
	}

	// Nos conectamos con el coordinador
	conn, err := rpc.DialHTTP("tcp", coordinador)

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

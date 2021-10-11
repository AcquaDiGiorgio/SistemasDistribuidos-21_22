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
	"encoding/gob"
	"fmt"
	"math"
	"net"
	"os"

	"main/com"
)

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func siguiente(ini int) (fin int) {
	ini64 := float64(ini)
	fin = ini + int(26000-785*math.Pow(ini64, 0.3))
	return
}

func descomponerTarea(interval com.TPInterval) (intervalos []com.TPInterval) {
	ini := interval.A
	fin := siguiente(ini)

	for fin <= interval.B {
		intervalos = append(intervalos, com.TPInterval{ini, fin})
		ini = fin + 1
		fin = siguiente(ini)
	}
	if fin > interval.B {
		intervalos = append(intervalos, com.TPInterval{ini, interval.B})
	}
	return
}

func AtenderCliente(canal chan net.Conn, dirWorker string) {
	var peticion com.Request
	var respuesta com.Reply

	//Lee de canal una conexion por iteracion y lo guarda en conn. Sale del bucle cuando esta vacio.
	for conn := range canal {

		//Recibe del cliente la peticion con los datos
		dec := gob.NewDecoder(conn)
		dec.Decode(&peticion)

		//Le envia al worker los datos
		work, err := net.Dial("tcp", dirWorker)
		checkError(err)

		enc := gob.NewEncoder(work)
		enc.Encode(&peticion)

		//Recibe la solucion del worker
		rec := gob.NewDecoder(work)
		rec.Decode(&respuesta)

		//Enviar solucion al cliente
		env := gob.NewEncoder(conn)
		env.Encode(respuesta)

		conn.Close()
	}
}

const CONN_TYPE = "tcp"
const CONN_HOST = "localhost"
const CONN_PORT = "30000"
const POOL = 5

func main() {
	canal := make(chan net.Conn) //Canal que pasa las tareas a las gorutines
	workers := [POOL]string{"localhost:8000", "localhost:8001", "localhost:8002", "localhost:8003", "localhost:8004"}

	listener, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	checkError(err)
	defer listener.Close()

	for i := 0; i < POOL; i++ {
		go AtenderCliente(canal, workers[i]) //Crea la pool gorutines
	}

	for {
		conn, err := listener.Accept()
		checkError(err)

		canal <- conn //Manda la conexion por el canal hacia las gorutines
	}
}

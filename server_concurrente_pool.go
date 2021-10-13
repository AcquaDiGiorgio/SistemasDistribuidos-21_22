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

// PRE: verdad
// POST: IsPrime devuelve verdad si n es primo y falso en caso contrario
func IsPrime(n int) (foundDivisor bool) {
	foundDivisor = false
	for i := 2; (i < n) && !foundDivisor; i++ {
		foundDivisor = (n%i == 0)
	}
	return !foundDivisor
}

// PRE: interval.A < interval.B
// POST: FindPrimes devuelve todos los números primos comprendidos en el
// 		intervalo [interval.A, interval.B]
func FindPrimes(interval com.TPInterval) (primes []int) {
	for i := interval.A; i <= interval.B; i++ {
		if IsPrime(i) {
			primes = append(primes, i)
		}
	}
	return primes
}

func AtenderCliente(canal chan net.Conn) {

	for {
		conn := <-canal

		fallo := false

		dec := gob.NewDecoder(conn)
		enc := gob.NewEncoder(conn)

		var peticion com.Request
		var respuesta com.Reply

		for !fallo {
			err := dec.Decode(&peticion)

			if err != nil {
				fallo = true
				continue
			}

			respuesta.Id = peticion.Id
			respuesta.Primes = FindPrimes(peticion.Interval)

			enc.Encode(respuesta)
		}
	}

}

const CONN_TYPE = "tcp"
const CONN_HOST = "localhost"
const CONN_PORT = "8002"

func main() {
	pool := 5                    //Tamaño de la pool de gorutines
	canal := make(chan net.Conn) //Canal que pasa las tareas a las gorutines

	listener, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	checkError(err)
	//defer listener.Close()

	for i := 1; i <= pool; i++ {
		go AtenderCliente(canal) //Crea la pool de gorutines
	}

	for {
		conn, err := listener.Accept()
		checkError(err)
		canal <- conn //Manda la conexion por el canal hacia las gorutines
	}
}

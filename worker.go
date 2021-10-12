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
	"main/com"
	"net"
	"os"
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

const CONN_TYPE = "tcp"

func main() {
	args := os.Args[1:]

	if len(args) != 1 {
		fmt.Println("Número de parámetros incorrecto, ejecutar como:\n\tgo run worker.go ip:port")
		os.Exit(1)
	}

	listener, err := net.Listen(CONN_TYPE, args[0])
	checkError(err)
	defer listener.Close()

	var peticion com.Request
	var respuesta com.Reply
	for {
		conn, err := listener.Accept()
		checkError(err)

		dec := gob.NewDecoder(conn)
		enc := gob.NewEncoder(conn)

		//Recibe del master la peticion para el calculo
		dec.Decode(&peticion)
		respuesta.Id = peticion.Id
		respuesta.Primes = FindPrimes(peticion.Interval)

		//Envia al master el array y cierra
		enc.Encode(respuesta)
		conn.Close()
	}
}

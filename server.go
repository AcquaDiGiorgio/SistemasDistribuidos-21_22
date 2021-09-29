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

func codificarRerspuesta(reply com.Reply) (codigo []byte) {
	codigo = make([]byte, len(reply.Primes)+1)
	codigo[0] = byte(reply.Id)

	for i := 1; i < len(codigo); i++ {
		codigo[i] = byte(reply.Primes[i-1])
	}
	return
}

func descodificarPeticion(codigo []byte) (reply com.Request) {
	reply.Id = int(codigo[0])
	reply.Interval.A = int(codigo[1])
	reply.Interval.B = int(codigo[2])
	return
}

const CONN_TYPE = "tcp"
const CONN_HOST = "localhost"
const CONN_PORT = "30000"

func main() {
	/*
		listener, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
		checkError(err)

		conn, err := listener.Accept()
		defer conn.Close()
		checkError(err)

		checkError(err)
	*/

	peticion := []byte{1, 232, 96}
	fmt.Println(descodificarPeticion(peticion))
}

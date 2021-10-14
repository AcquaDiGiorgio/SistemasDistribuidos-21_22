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

//Struct usado para realizar el envío de mensajes por canal.
//Consta de un encoder, para devolver el dato y la petición del cliente.
type Mensaje struct {
	encoder *gob.Encoder
	request com.Request
}

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

//Gorutina que acepta un mensaje y calcula sus primos para posteriormente
//devolverlos al cliente
func AtenderCliente(msj Mensaje) {

	enc := msj.encoder
	peticion := msj.request

	var respuesta com.Reply

	respuesta.Id = peticion.Id
	respuesta.Primes = FindPrimes(peticion.Interval)

	enc.Encode(respuesta)
}

const CONN_TYPE = "tcp"
const CONN_HOST = "155.210.154.210"

func main() {
	args := os.Args[1:]
	if len(args) != 1 {
		os.Exit(1)
	}
	listener, err := net.Listen(CONN_TYPE, CONN_HOST+":"+args[0])
	checkError(err)
	defer listener.Close()

	//Aceptamos a un cliente
	for {
		conn, err := listener.Accept()
		checkError(err)

		dec := gob.NewDecoder(conn)
		enc := gob.NewEncoder(conn)

		var request com.Request

		fallo := false
		//Mientras tenga algo que darnos y no haya cerrado conexión,
		//acptamos lo que nos dé
		for !fallo {
			err = dec.Decode(&request)
			if err != nil {
				fallo = true
				continue
			}

			go AtenderCliente(Mensaje{enc, request})
		}
	}
}

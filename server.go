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
	"math/big"
	"net"
	"os"
	"time"

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

func int_to_byte(ent int) (byt []byte) {

	var s = big.NewInt(int64(ent))
	b := s.Bytes()

	byt = make([]byte, 4)
	var pos = 0

	for j := 0; j < 4; j++ {
		if j < (4 - len(b)) {
			byt[j] = 0x0
		} else {
			byt[j] = b[pos]
			pos++
		}
	}

	return
}

func byte_to_int(byt []byte) (ent int) {
	var r = big.NewInt(0).SetBytes(byt)
	ent = int(r.Int64())
	return
}

func codificarRerspuesta(reply com.Reply) (codigo []byte) {

	codigo = append(codigo, int_to_byte(reply.Id)...)

	for i := 0; i < len(reply.Primes); i++ {
		codigo = append(codigo, int_to_byte(reply.Primes[i])...)
	}

	return
}

func descodificarPeticion(codigo []byte) (reply com.Request) {
	reply.Id = byte_to_int(codigo[0:4])
	reply.Interval.A = byte_to_int(codigo[4:8])
	reply.Interval.B = byte_to_int(codigo[8:12])
	return
}

func thread(c net.Conn) {
	// Cierra la conexión cuando termina la ejecución de la función integrada (la conexión con un cliente)
	defer c.Close()
	/*
		Recepción de mensajes
	*/
	var codigo [12]byte                          // buffer de max 512 bytes
	n, _ := c.Read(codigo[:])                    // Leemos todo el buffer
	recibido := descodificarPeticion(codigo[:n]) // Mostramos hasta el tam leido
	/*
		Buscamos los primos
	*/
	var respuesta com.Reply
	respuesta.Id = recibido.Id
	respuesta.Primes = FindPrimes(recibido.Interval)

	mensaje := codificarRerspuesta(respuesta)
	/*
		Envío de mensajes
	*/
	c.Write(mensaje) // Escribimos a través de la conexión
}

const CONN_TYPE = "tcp"
const CONN_HOST = "localhost"
const CONN_PORT = "30000"

func main() {
	listener, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	checkError(err)
	defer listener.Close()

	for {
		// Abrimos conexión con un Cliente y comprobamos que todo esté correcto
		conn, err := listener.Accept()
		checkError(err)
		// Tiempo Límite que puede estar una conexión mantenida (1 hora en este caso)
		conn.SetDeadline(time.Now().Add(time.Hour))
		go thread(conn)
	}
}

/*
* AUTOR: Rafael Tolosana Calasanz
* EDITADO: Jorge Lisa y David Zandundo
* ASIGNATURA: 30221 Sistemas Distribuidos del Grado en Ingeniería Informática
*			  Escuela de Ingeniería y Arquitectura - Universidad de Zaragoza
* FECHA: septiembre de 2021
* FICHERO: client.go
* DESCRIPCIÓN: cliente completo para los cuatro escenarios de la práctica 1
 */
package main

import (
	"fmt"
	"main/com"
	"os"
)

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func codificarPeticion(request com.Request) (codigo [3]byte) {
	codigo[0] = byte(request.Id)
	codigo[1] = byte(request.Interval.A)
	codigo[2] = byte(request.Interval.B)
	return
}

func descodificarRespuesta(codigo []byte) (reply com.Reply) {
	reply.Id = int(codigo[0])

	reply.Primes = make([]int, len(codigo)-1)

	for i := 1; i < len(codigo); i++ {
		reply.Primes[i-1] = int(codigo[i])
	}
	return
}

func main() {
	//endpoint := "localhost:30000"

	// TODO: crear el intervalo solicitando dos números por teclado
	interval := com.TPInterval{1000, 700000}
	request := com.Request{1, interval}
	peticion := codificarPeticion(request)
	fmt.Println(peticion)

	codigo := []byte{0xF, 0x2, 0x3, 0x4, 0x5, 0x6, 0xA}
	respuesta := descodificarRespuesta(codigo)
	fmt.Println(respuesta)

	//tcpAddr, err := net.ResolveTCPAddr("tcp", endpoint)
	//checkError(err)

	//conn, err := net.DialTCP("tcp", nil, tcpAddr)
	//checkError(err)

}

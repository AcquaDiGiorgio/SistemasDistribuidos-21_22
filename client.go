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
	"encoding/gob"
	"fmt"
	"main/com"
	"net"
	"os"
	"time"
)

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func main() {
	endpoint := "localhost:30000"

	// TODO: crear el intervalo solicitando dos números por teclado
	interval := com.TPInterval{1, 10}
	request := com.Request{1, interval}

	tcpAddr, err := net.ResolveTCPAddr("tcp", endpoint)
	checkError(err)

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	checkError(err)
	defer conn.Close()
	/*
		Envío de mensajes
	*/
	enc := gob.NewEncoder(conn)
	start := time.Now()
	enc.Encode(request)

	/*
		Recepción de mensajes
	*/
	dec := gob.NewDecoder(conn)
	var respuesta com.Reply
	dec.Decode(&respuesta)
	end := time.Now()

	fmt.Println(respuesta)
	fmt.Println("\t", end.Sub(start))
}

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
	"math/big"
	"net"
	"os"
	"strconv"
)

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
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

func codificarPeticion(request com.Request) (codigo []byte) {
	codigo = append(codigo, int_to_byte(request.Id)...)
	codigo = append(codigo, int_to_byte(request.Interval.A)...)
	codigo = append(codigo, int_to_byte(request.Interval.B)...)
	return
}

func descodificarRespuesta(codigo []byte) (reply com.Reply) {
	reply.Id = byte_to_int(codigo[0:4])
	totPrimos := len(codigo)/4 - 1
	reply.Primes = make([]int, totPrimos)

	for i := 0; i < totPrimos; i++ {
		ini := 4 * (i + 1)
		fin := ini + 4
		reply.Primes[i] = byte_to_int(codigo[ini:fin])
	}

	return
}

const BUFF_SIZE = 524288        // 2^19 bytes -- Hasta 131.072 enteros de tamaño 8 bytes
const MAX_INT_SIZE = 4294967296 // entero de 8 bytes

func main() {

	args := os.Args[1:]

	if len(args) != 4 {
		fmt.Println("Número de parámetros incorrecto, ejecutar como:\n\tgo run client.go IP PORT intervalo_ini intervalo_fin")
		os.Exit(1)
	}

	endpoint := args[0] + ":" + args[1]

	ini, _ := strconv.Atoi(args[2])
	fin, _ := strconv.Atoi(args[3])

	if (ini > MAX_INT_SIZE) || (fin > MAX_INT_SIZE) || (ini < 0) || (fin < 0) {
		fmt.Println("El intervalo debe ser de números enteros positivos de tamaño menor o igual a 8 bytes")
		os.Exit(2)
	}

	if ini > fin {
		aux := ini
		ini = fin
		fin = aux
	}

	interval := com.TPInterval{ini, fin}
	request := com.Request{1, interval}
	peticion := codificarPeticion(request)

	tcpAddr, err := net.ResolveTCPAddr("tcp", endpoint)
	checkError(err)

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	checkError(err)

	conn.Write(peticion)

	var codigo [BUFF_SIZE]byte
	n, _ := conn.Read(codigo[:])

	respuesta := descodificarRespuesta(codigo[:n])
	fmt.Print("Primos encontrados: ")
	fmt.Println(respuesta.Primes)
}

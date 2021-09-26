package main

import (
	"bufio"
	"fmt"
	"net"
	"time"
)

func checkError(err error, str string) {
	if err != nil {
		fmt.Print("ERROR EN " + str)
	}
}

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	conn.SetDeadline(time.Now().Add(time.Hour))

	// Cierra la conexión cuando termina la ejecución del main
	defer conn.Close()
	checkError(err, "Dial")
	/*
		Envío de mensajes
	*/
	b := []byte("Hola\x00") // terminador x000 == \0 en Java o C
	trueW, _ := conn.Write(b)
	fmt.Print("Bytes Written - ", trueW, "\n")
	/*
		Recepción de mensajes
	*/
	read, _ := bufio.NewReader(conn).ReadBytes('\x00')

	fmt.Print("Mesaje Recibido: " + string(read) + "\n")
}

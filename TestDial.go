package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	conn.SetDeadline(time.Now().Add(time.Hour))

	defer conn.Close()

	if err != nil {
		// handle error
	}
	/*
		Envío de mensajes
	*/
	b := []byte("Hola")
	trueW, _ := conn.Write(b)
	fmt.Print("Bytes Written - ", trueW, "\n")
	/*
		Recepción de mensajes
	*/
	var read []byte
	trueR, _ := conn.Read(read)
	fmt.Print("Bytes Read - ", trueR)

	fmt.Print(string(read))
}

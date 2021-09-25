package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	ln, err := net.Listen("tcp", ":8080")
	defer ln.Close()

	if err != nil {
		fmt.Print("ERROR en LISTEN")
	}

	for {
		conn, err := ln.Accept()
		conn.SetDeadline(time.Now().Add(time.Hour))

		if err != nil {
			fmt.Print("ERROR en ACCEPT\n")
		}

		go func(c net.Conn) {
			defer c.Close()
			/*
				Recepción de mensajes
			*/
			var read []byte
			trueR, _ := c.Read(read)
			fmt.Print("Bytes Read - ", trueR, "\n")
			/*
				Envío de mensajes
			*/
			b := []byte("Adios")
			trueW, _ := c.Write(b)
			fmt.Print("Bytes Written - ", trueW, "\n")

			fmt.Print(string(read))
		}(conn)

	}
}

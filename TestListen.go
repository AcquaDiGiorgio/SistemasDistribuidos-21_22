package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"
)

func checkError(err error, str string) {
	if err != nil {
		fmt.Print("ERROR EN " + str)
	}
}

func main() {
	var wg sync.WaitGroup // Creamos un semaforo que espera a todos los clientes
	ln, err := net.Listen("tcp", ":8080")
	// Cierra la conexión cuando termina la ejecución del main (la escucha en el puerto 8080)
	defer ln.Close()
	defer fmt.Print("Cerramos Puerto\n")

	checkError(err, "Listen")

	for i := 0; i < 2; i++ {
		fmt.Print("Iter numero: " + strconv.Itoa(i) + "\n")
		conn, err := ln.Accept()
		checkError(err, "Accept")

		// Tiempo Límite que puede estar una conexión mantenida (1 hora en este caso)
		conn.SetDeadline(time.Now().Add(time.Hour))

		go func(c net.Conn) {
			wg.Add(1)
			// Cierra la conexión cuando termina la ejecución de la función integrada (la conexión con un cliente)
			defer c.Close()
			// Decimos al semaforo que hemos terminado cuando termiene el thread
			defer wg.Done()
			defer fmt.Print("Cerramos Conexión\n")
			/*
				Recepción de mensajes
			*/
			read, _ := bufio.NewReader(c).ReadBytes('\x00')
			fmt.Print("Mesaje Recibido: " + string(read) + "\n")
			/*
				Envío de mensajes
			*/
			b := []byte("Adios\x00") // terminador x000 == \0 en Java o C || También valdría \000
			trueW, _ := c.Write(b)
			fmt.Print("Bytes Written - ", trueW, "\n")
			c.Close()
		}(conn) // Parámetro que pasamos a la función embebida
	}
	// Esperamos a que terminen todos
	wg.Wait()
}

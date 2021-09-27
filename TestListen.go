package main

import (
	"fmt"
	"net"
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
	// Empezamos a ecuchar en el puerto y checkeamos error
	ln, err := net.Listen("tcp", ":8080")
	checkError(err, "Listen")
	// Cierra la conexión cuando termina la ejecución del main (la escucha en el puerto 8080)
	defer ln.Close()
	defer fmt.Print("Cerramos Puerto\n")

	// Aceptamos solo 2 Clientes
	for i := 0; i < 2; i++ {
		// Abrimos conexión con un Cliente y comprobamos que todo esté correcto
		conn, err := ln.Accept()
		checkError(err, "Accept")
		// Añadimos un proceso a la espera
		wg.Add(1)
		// Tiempo Límite que puede estar una conexión mantenida (1 hora en este caso)
		conn.SetDeadline(time.Now().Add(time.Hour))

		go func(c net.Conn) {
			// Cierra la conexión cuando termina la ejecución de la función integrada (la conexión con un cliente)
			defer c.Close()
			// Decimos al semaforo que hemos terminado cuando termiene el thread
			defer wg.Done()
			defer fmt.Print("Cerramos Conexión\n")
			/*
				Recepción de mensajes
			*/
			var lectura [512]byte                                       // buffer de max 512 bytes
			n, _ := conn.Read(lectura[:])                               // Leemos todo el buffer
			fmt.Print("Mesaje Recibido: " + string(lectura[:n]) + "\n") // Mostramos hasta el tam leido
			/*
				Envío de mensajes
			*/
			b := []byte("Adios")   // terminador x000 == \0 en Java o C || También valdría \000
			trueW, _ := c.Write(b) // Escribimos a través de la conexión
			fmt.Print("Bytes Escritos - ", trueW, "\n")
		}(conn) // Parámetro que pasamos a la función embebida
	}
	wg.Wait() // Esperamos a que terminen todos
}

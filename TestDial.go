package main

import (
	"fmt"
	"net"
	"time"
)

func checkError(err error) {
	if err != nil {
		fmt.Print("ERROR - " + err.Error())
	}
}

// No es necesario implementarla
/*
	Pre: Conexión con el Servidor
	Post: Cadena de enteros sin signo leidos

	Función especial que lee una muletilla y luego el resto del
	mensaje.
	La muletilla marca el tipo de máquina (Big endian o Little endian)
	que es el servidor y posteriormente trata el mensaje dependiendo
	de esta muletilla.
*/
func read(conn net.Conn) []uint16 {
	var buf [512]byte
	// Leemos primero el BOM
	bytesRead, err := conn.Read(buf[0:2])
	checkError(err)

	// Leemos el resto del Mensaje
	for true {
		m, err := conn.Read(buf[bytesRead:])
		if m == 0 || err != nil {
			break
		}
		bytesRead += m
	}

	// Creamos la respuesta de tamaño bytesRead/2 (de bytes a uint16: tamaño 8 a tamaño 16)
	var respuesta []uint16
	respuesta = make([]uint16, bytesRead/2)

	// Realizamos una cosa u otra dependiendo del timpo de máquina que es
	if buf[0] == 0xff && buf[1] == 0xfe { // Big Endian
		for i := 2; i < bytesRead; i += 2 {
			respuesta[i/2] = uint16(buf[i])<<8 + uint16(buf[i+1])
		}
	} else if buf[1] == 0xff && buf[0] == 0xfe { // Little Endian
		for i := 2; i < bytesRead; i += 2 {
			respuesta[i/2] = uint16(buf[i+1])<<8 + uint16(buf[i])
		}
	} else { // Orden no Identificado
		return nil
	}
	return respuesta
}

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	conn.SetDeadline(time.Now().Add(time.Hour))

	// Cierra la conexión cuando termina la ejecución del main
	defer conn.Close()
	checkError(err)
	/*
		Envío de mensajes
	*/
	b := []byte("Hola")
	trueW, _ := conn.Write(b)
	fmt.Print("Bytes Escritos - ", trueW, "\n")
	/*
		Recepción de mensajes
	*/
	var lectura [512]byte
	n, _ := conn.Read(lectura[:])

	fmt.Print("Mesaje Recibido: " + string(lectura[:n]) + "\n")
}

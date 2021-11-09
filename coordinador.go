package main

import (
	"encoding/gob"
	"log"
	"main/com"
	"net"
	"os"
)

type action struct {
	accion int
	args   []string
}

const (
	LANZAR_WORKER = 0
)

func lanzarWorker(worker string, ip string, usuario string, pass string) {
	//Creamos el ssh hacia la máquina en la que se encuentra el worker
	ssh, err := com.NewSshClient(
		usuario,
		worker,
		22,
		RSA,
		pass)
	if err != nil {
		log.Printf("SSH init error %v", err)
		os.Exit(1)
	}
	err = ssh.RunCommand(PATH + "worker " + ip)
	checkErrorCoord(err)
}

func main() {
	listener, err := net.Listen("", "")
	checkErrorCoord(err)

	for {
		conn, err := listener.Accept()
		checkErrorCoord(err)

		go func(net.Conn) {
			enc := gob.NewEncoder(conn)
			dec := gob.NewDecoder(conn)

			var msg action
			dec.Decode(msg)

			switch msg.accion {
			case LANZAR_WORKER:

				break

			}

		}(conn)
	}
}

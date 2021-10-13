/*
* AUTOR: Rafael Tolosana Calasanz
* EDITADO: Jorge Lisa y David Zandundo
* ASIGNATURA: 30221 Sistemas Distribuidos del Grado en Ingeniería Informática
*			  Escuela de Ingeniería y Arquitectura - Universidad de Zaragoza
* FECHA: septiembre de 2021
* FICHERO: server.go
* DESCRIPCIÓN: contiene la funcionalidad esencial para realizar los servidores
*			   correspondientes al trabajo 1
 */
package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"strings"
	"syscall"

	"main/com"

	"golang.org/x/term"
)

const PATH = "~/SSDD/Practica1/"
const RSA = "/home/a774248/.ssh/id_rsa"

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func siguiente(ini int) (fin int) {
	ini64 := float64(ini)
	fin = ini + int(26000-785*math.Pow(ini64, 0.3))
	return
}

func descomponerTarea(interval com.TPInterval) (intervalos []com.TPInterval) {
	ini := interval.A
	fin := siguiente(ini)

	for fin <= interval.B {
		intervalos = append(intervalos, com.TPInterval{ini, fin})
		ini = fin + 1
		fin = siguiente(ini)
	}
	if fin > interval.B {
		intervalos = append(intervalos, com.TPInterval{ini, interval.B})
	}
	return
}

func AtenderCliente(canal chan net.Conn, dirWorker string) {
	work, err := net.Dial("tcp", dirWorker)
	checkError(err)

	workEnvio := gob.NewEncoder(work)
	workRecepcion := gob.NewDecoder(work)

	//Lee de canal una conexion por iteracion y lo guarda en conn.
	for {
		fmt.Println("Nueva conexión!")
		conn := <-canal
		i := 1
		fallo := false

		clienteEnvio := gob.NewEncoder(conn)
		clienteRecepcion := gob.NewDecoder(conn)

		for !fallo {
			var peticion com.Request
			var respuesta com.Reply

			//Recibe del cliente la peticion con los datos

			err := clienteRecepcion.Decode(&peticion)
			if err != nil {
				fallo = true
				continue
			}

			fmt.Println("Atiendo petición", i)
			i++

			// Envío la petición y recibo la respuesta del worker
			workEnvio.Encode(peticion)
			workRecepcion.Decode(&respuesta)

			clienteEnvio.Encode(respuesta)
		}
	}
}

func LanzarWorker(worker string, ip string, usuario string, pass string) {
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

	err = ssh.RunCommand(PATH + "worker " + ip + " &")

	if err != nil {
		log.Printf("SSH run command error %v", err)
		os.Exit(2)
	}
}

func inicializacion(canal chan net.Conn) {
	var user string
	fmt.Print("Introduzca el usuario: ")
	fmt.Scanf("%s", &user)

	fmt.Print("Introduzca la Contraseña: ")
	pass, err := term.ReadPassword(int(syscall.Stdin))
	checkError(err)

	passStr := strings.TrimSpace(string(pass))

	for i := 0; i < POOL; i++ {
		LanzarWorker(com.HOSTS[i], com.IPs[i], user, passStr)
		fmt.Println("Worker", i, "en ejecución")
	}
}

const CONN_TYPE = "tcp"
const CONN_HOST = "155.210.154.210"
const CONN_PORT = "8000"
const POOL = 6

func main() {

	canal := make(chan net.Conn) //Canal que pasa las tareas a las gorutines

	listener, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	checkError(err)
	defer listener.Close()

	inicializacion(canal)
	fmt.Println("\nWorkers en ejecución")

	for i := 0; i < POOL; i++ {
		go AtenderCliente(canal, com.IPs[i])
	}

	for {
		conn, err := listener.Accept()
		fmt.Println("Accepto cleinte") //BORRAR
		checkError(err)
		canal <- conn //Manda la conexion por el canal hacia las gorutines
	}
}

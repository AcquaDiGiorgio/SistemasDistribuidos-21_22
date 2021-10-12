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

	"main/com"
)

const PATH = "~/SSDD/Practica1/"
const RSA = "/Users/jorge/.ssh/id_rsa"

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

	err = ssh.RunCommand(PATH + "worker " + ip)

	if err != nil {
		log.Printf("SSH run command error %v", err)
		os.Exit(2)
	}
}

func AtenderCliente(canal chan net.Conn, dirWorker string) {
	var peticion com.Request
	var respuesta com.Reply

	//Lee de canal una conexion por iteracion y lo guarda en conn.
	for {
		conn := <-canal
		for {
			clienteRecepcion := gob.NewDecoder(conn)
			clienteEnvio := gob.NewEncoder(conn)

			fmt.Println("Atiendo cliente")
			//Recibe del cliente la peticion con los datos

			err := clienteRecepcion.Decode(&peticion)
			if err != nil {
				continue
			}
			fmt.Println("Cliente me da: ", peticion.Interval)

			//Le envia al worker los datos
			work, err := net.Dial("tcp", dirWorker)

			checkError(err)

			workRecepcion := gob.NewEncoder(work)
			workEnvio := gob.NewDecoder(work)

			fmt.Println("Envío al worker")

			workRecepcion.Encode(peticion)
			workEnvio.Decode(&respuesta)

			fmt.Println("Recibo del worker: ", respuesta.Primes)

			//Enviar solucion al cliente
			clienteEnvio.Encode(respuesta)
		}
	}

}

const CONN_TYPE = "tcp"
const CONN_HOST = "localhost"
const CONN_PORT = "8005"
const POOL = 6

func main() {
	/*
		var user string
		fmt.Print("Introduzca el usuario: ")
		fmt.Scanf("%s", &user)

		fmt.Print("Introduzca la Contraseña: ")
		pass, err := term.ReadPassword(int(syscall.Stdin))
		passStr := strings.TrimSpace(string(pass))
	*/
	canal := make(chan net.Conn) //Canal que pasa las tareas a las gorutines

	listener, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	checkError(err)
	defer listener.Close()
	/*
		for i := 0; i < POOL; i++ {
			LanzarWorker(com.HOSTS[i], com.IPs[i], user, passStr)
			go AtenderCliente(canal, com.IPs[i])
		}
	*/
	//LanzarWorker(com.HOSTS[0], com.IPs[0], user, passStr)
	go AtenderCliente(canal, "localhost:8006")

	fmt.Println("\nWorkers en ejecución")

	for {
		conn, err := listener.Accept()
		fmt.Println("Accepto cleinte") //BORRAR
		checkError(err)
		canal <- conn //Manda la conexion por el canal hacia las gorutines
	}
}

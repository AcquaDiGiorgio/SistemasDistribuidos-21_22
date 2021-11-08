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
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"syscall"
	"time"

	"main/com"

	"golang.org/x/term"
)

const PATH = "/home/a774248/SSDD/Practica1/"
const RSA = "/home/a774248/.ssh/id_rsa"
const CONN_TYPE = "tcp"
const CONN_HOST = "155.210.154.200"

//Struct usado para realizar el envío de mensajes por canal.
//Consta de un encoder, para devolver el dato y la petición del cliente.
type Mensaje struct {
	reply     []int
	intervalo com.TPInterval
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

//Gorutina capaz de lanzar por ssh un worker y esperar a que entre por el canal de mensajes
//una petición del cliente
//Esta función recibe el host del worker, su ip, el usuario que hace el ssh, su contraseña y
//el canal por donde se recibirán las peticiones
func LanzarWorker(worker string, ip string, usuario string, pass string, canal chan Mensaje) {
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

	//Ejecutamos su archivo compilado
	err = ssh.RunCommand(PATH + "worker " + ip)
	if err != nil {
		log.Printf("SSH run command error %v", err)
		os.Exit(2)
	}

	//Esperamos para que asegurarnos de que el worker está preparado para escuchar
	time.Sleep(1 * time.Second)

	work, err := net.Dial("tcp", ip) // CAMBIAR A RPC
	checkError(err)

	fmt.Println("Worker", worker, "preparado")
	for {
		msj := <-canal

		//
	}
}

//función que recibiendo un canal de Mensajes y un puerto por donde escuchen
//los workers, inicializa los workers que se encuentran en el paquete com
func inicializacion(canal chan Mensaje) {
	var user string
	fmt.Print("Introduzca el usuario: ")
	fmt.Scanf("%s", &user)

	fmt.Print("Introduzca la Contraseña: ")
	pass, err := term.ReadPassword(int(syscall.Stdin))
	checkError(err)

	passStr := strings.TrimSpace(string(pass))

	for i := 0; i < com.POOL; i++ {
		go LanzarWorker(com.HOSTS[i], com.IPs[i], user, passStr, canal)
	}
}

func main() {

	args := os.Args[1:]
	if len(args) != 1 {
		os.Exit(1)
	}

	//Creamos un canal que pasa las tareas a las gorutines
	canal := make(chan Mensaje)

	listener, err := net.Listen(CONN_TYPE, CONN_HOST+":"+args[0])
	checkError(err)
	defer listener.Close()

	//Llama por ssh a los workers y los prepara para escuchar
	inicializacion(canal)

	//Aceptamos a un cliente
	for {
		conn, err := listener.Accept() // HACER POR RPC
		checkError(err)

		fallo := false
		//Mientras tenga algo que darnos y no haya cerrado conexión,
		//acptamos lo que nos dé
		for !fallo {

			//Enviamos su petición y el lugar por donde le tenemos que reponder
			//por el canal
			canal <- Mensaje{} // RELLENAR
		}
	}
}

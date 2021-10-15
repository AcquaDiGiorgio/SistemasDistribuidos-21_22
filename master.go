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
	"time"

	"main/com"

	"golang.org/x/term"
)

const PATH = "/home/a774248/SSDD/Practica1/"
const RSA = "/home/a774248/.ssh/id_rsa"

//Struct usado para realizar el envío de mensajes por canal.
//Consta de un encoder, para devolver el dato y la petición del cliente.
type Mensaje struct {
	encoder *gob.Encoder
	request com.Request
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

//Calcula el siguiente valor de un intrevalo partido
//usando la fórmula 26000-785*x^0.3
func siguiente(ini int) (fin int) {
	ini64 := float64(ini)
	fin = ini + int(26000-785*math.Pow(ini64, 0.3))
	return
}

//Función capaz de partir el intervalo en subintervalos con una carga de
//trabajo parcialmente equitativa
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

	work, err := net.Dial("tcp", ip)
	checkError(err)

	workEnvio := gob.NewEncoder(work)
	workRecepcion := gob.NewDecoder(work)

	fmt.Println("Worker", worker, "preparado")
	for {
		msj := <-canal

		enc := msj.encoder
		peticion := msj.request

		var respuesta com.Reply

		// Envío la petición y recibo la respuesta del worker
		workEnvio.Encode(peticion)
		workRecepcion.Decode(&respuesta)

		enc.Encode(respuesta)
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

const CONN_TYPE = "tcp"
const CONN_HOST = "155.210.154.200"
const POOL = 6

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
		conn, err := listener.Accept()
		checkError(err)

		dec := gob.NewDecoder(conn)
		enc := gob.NewEncoder(conn)

		var request com.Request

		fallo := false
		//Mientras tenga algo que darnos y no haya cerrado conexión,
		//acptamos lo que nos dé
		for !fallo {
			err = dec.Decode(&request)
			if err != nil {
				fallo = true
				continue
			}

			//Enviamos su petición y el lugar por donde le tenemos que reponder
			//por el canal
			canal <- Mensaje{enc, request}
		}
	}
}

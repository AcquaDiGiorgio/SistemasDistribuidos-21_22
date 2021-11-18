/*
* AUTORES: Jorge Lisa y David Zandundo
* ASIGNATURA: 30221 Sistemas Distribuidos del Grado en Ingeniería Informática
*			  Escuela de Ingeniería y Arquitectura - Universidad de Zaragoza
* FECHA: noviembre de 2021
* FICHERO: master.go
 */

package main

import (
	"fmt"
	"math"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"syscall"
	"time"

	"main/com"

	"golang.org/x/term"
)

const (
	MAX_TIME = time.Duration(2921 * time.Millisecond)
	PATH     = "/home/a774248/SSDD/Practica3/"
	RSA      = "/home/a774248/.ssh/id_rsa"
)

//Struct usado para realizar el envío de mensajes por canal.
//Consta de un encoder, para devolver el dato y la petición del cliente.
type Respuesta struct {
	reply []int
	err   error
}

type Mensaje struct {
	intervalo com.TPInterval
	resp      chan Respuesta
}

type Primes struct {
	canal chan Mensaje
	user  string
	pass  string
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func (p *Primes) lanzarWorker(id int) error {

	//Creamos el ssh hacia la máquina en la que se encuentra el worker
	fmt.Printf("LANZANDO WORKER %d A TRAVÉS DE SSH\n", id) // XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX

	ssh, err := com.NewSshClient(
		p.user,
		com.Workers[id].Host,
		22,
		RSA,
		p.pass)

	if err == nil {
		fmt.Printf("WORKER %d LANZANDO CORRECTAMENTE\n", id)
		err = ssh.RunCommand(PATH + "worker_configurable " + com.Workers[id].Ip + " 0 0 0")
	}

	return err
}

func trabajar(id int, primes *Primes, worker *rpc.Client) { // ELIMINAR ID
	callChan := make(chan *rpc.Call, 10)
	for {
		msj := <-primes.canal              // Recibimos la petición del master
		aprxDur := aproxThr(msj.intervalo) // Calculamos el coste aproximado

		fmt.Printf("INTERVALO %d -> %d RECIBIDO POR EL WORKER %d\n", msj.intervalo.A, msj.intervalo.B, id) // XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX

		var respuesta Respuesta

		worker.Go("PrimesImpl.FindPrimes", msj.intervalo, &respuesta.reply, callChan) // Enviamos al worker el trabajo

		select {
		case msg := <-callChan: // Recepción del mensaje a tiempo
			respuesta.err = msg.Error
			msj.resp <- respuesta
			if msg.Error != nil { // CRASH
				// Terminamos de Trabajar
				return
			}
			break

		case <-time.After(MAX_TIME + aprxDur): // Más retraso del permitido
			// Le pasamos el mensaje a otro worker y esperamos 0.250s
			primes.canal <- msj
			time.Sleep(250 * time.Millisecond)
			break
		}
	}
}

//Gorutina capaz de lanzar por ssh un worker y esperar a que entre por el canal de mensajes
//una petición del cliente
//Esta función recibe el host del worker, su ip, el usuario que hace el ssh y su contraseña
func ejecutarWorker(id int, primes *Primes) {
	for {
		worker, err := rpc.DialHTTP("tcp", com.Workers[id].Ip)
		if err != nil {
			continue
		}

		fmt.Printf("WORKER %d EN MARCHA\n", id) // XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
		trabajar(id, primes, worker)
	}
}

// Función RPC que conecta el cliente y el worker
func (p *Primes) FindPrimes(interval com.TPInterval, primeList *[]int) error {
	fmt.Printf("RECIBIDO INTERVALO %d -> %d DEL CLIENTE\n", interval.A, interval.B) // XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
	resp := make(chan Respuesta)

	p.canal <- Mensaje{interval, resp}         // Enviamos la petición a un worker
	respuesta := <-resp                        // Esperamos a la respuesta del worker
	fmt.Printf("RECIBO RESPUESTA DEL CANAL\n") // XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX

	if respuesta.err != nil {
		*primeList = respuesta.reply // Devolvemos la respuesta
	}
	return respuesta.err
}

func aproxThr(interval com.TPInterval) time.Duration {

	retVal := 0.0
	for j := interval.A; j <= interval.B; j += 1000 {
		retVal += 0.00164 * math.Pow(float64(j), 0.9055)
	}

	return time.Duration(int(retVal) * int(time.Millisecond))
}

func main() {

	// Creamos un canal que pasa las tareas a las gorutines
	primes := new(Primes)
	primes.canal = make(chan Mensaje)

	fmt.Print("Introduzca el usuario: ")
	fmt.Scanf("%s", &primes.user)

	fmt.Print("Introduzca la Contraseña: ")
	pass, err := term.ReadPassword(int(syscall.Stdin))
	checkError(err)

	primes.pass = string(pass)

	// Llama por ssh a los workers y los prepara para escuchar
	for i := 0; i < com.POOL; i++ {
		checkError(primes.lanzarWorker(i))
	}
	for i := 0; i < com.POOL; i++ {
		go ejecutarWorker(i, primes)
	}
	// Registro y Creación del RPC
	rpc.Register(primes)
	rpc.HandleHTTP()

	// Inicio Escucha
	listener, err := net.Listen("tcp", com.ENPOINT_MASTER)
	checkError(err)
	defer listener.Close()

	// Sirve petiticiones
	http.Serve(listener, nil)
}

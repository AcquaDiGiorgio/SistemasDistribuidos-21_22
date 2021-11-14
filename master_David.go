/*
* AUTOR: Jorge Lisa y David Zandundo
* ASIGNATURA: 30221 Sistemas Distribuidos del Grado en Ingeniería Informática
*			Escuela de Ingeniería y Arquitectura - Universidad de Zaragoza
* FECHA: octubre de 2021
* FICHERO: master.go
* DESCRIPCIÓN: contiene la funcionalidad esencial para realizar la practica 3
 */
package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net/rpc"
	"os"
	"practica3/com"
	"sync"
	"time"
)

const (
	NORMAL   = iota // NORMAL == 0
	DELAY    = iota // DELAY == 1
	CRASH    = iota // CRASH == 2
	OMISSION = iota // IOTA == 3

	WORKERS          = "workers.txt"
	MAX_INTENTOS     = 10                        //Numero de intentos
	MAX_TIMEOUT      = (time.Millisecond * 2000) //Debe calcularse para diferentes desplieges de workers
	TIMEOUT_OMISSION = (time.Second * 5)         //Si el worker no responde en TIMEOUT_OMISSION se entiende que ha ignorado
)

type PrimesImpl struct {
	delayMaxMilisegundos int
	delayMinMiliSegundos int
	behaviourPeriod      int
	behaviour            int
	i                    int
	mutex                sync.Mutex
}

type Tarea struct {
	datos     com.TPInterval //Intervalo para calcular los primos
	resultado *[]int         //Puntero al vector del resultado
	fin       chan bool      //Canal para comunicar que ya se ha terminado
}

var ColaTareas chan Tarea //Cola de tareas para que las cojan los proxys

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func isPrime(n int) (foundDivisor bool) {
	foundDivisor = false

	for i := 2; (i < n) && !foundDivisor; i++ {
		foundDivisor = (n%i == 0)
	}
	return !foundDivisor
}

// PRE: verdad
// POST: IsPrime devuelve verdad si n es primo y falso en caso contrario
func findPrimes(interval com.TPInterval) (primes []int) {
	for i := interval.A; i <= interval.B; i++ {
		if isPrime(i) {
			primes = append(primes, i)
		}
	}
	return primes
}
func leerWorkers(nombre string) ([]string, error) {
	fich, err := os.Open(nombre)
	if err != nil {
		return nil, err
	}
	defer fich.Close()

	var list []string
	scanner := bufio.NewScanner(fich)
	for scanner.Scan() {
		list = append(list, scanner.Text())
	}
	return list, scanner.Err()
}

func arrancaWorker(worker string) {
	//Arranca el worker especificado
	usuario := "a780500"
	pass := "xxxxxxxxx"
	PATH := "/home/a780500/SSDD/Practica3/"
	RSA := "/home/a780500/.ssh/id_rsa"
	IP := "localhost:30000"
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
	err = ssh.RunCommand(PATH + "worker " + IP)
	if err != nil {
		log.Printf("SSH run command error %v", err)
		os.Exit(2)
	}
	//Esperamos para que asegurarnos de que el worker está preparado para escuchar
	time.Sleep(1 * time.Second)
	//work, err := net.Dial("tcp", IP)
	//checkError(err)
}

//Proxys del master
func ProxyPrimes(worker string, colaTareas chan Tarea) {
	var reply []int
	c := make(chan error) //Canal de error, para informar del error

	go arrancaWorker(worker) //Arrancamos el worker asociado al proxy y le esperamos
	time.Sleep(5 * time.Second)

	workerCon, err := rpc.DialHTTP("tcp", worker) //Establecemos conexion con el worker asociado
	if err != nil {
		log.Fatal("dialing:", err)
	}
	fmt.Println(worker + " esta listo y conectado")

	opcion_omission := 1 //Selecciona la opcion para decidir que hacer cuando hay omision
	for {

		task := <-colaTareas

		go func() {
			c <- workerCon.Call("PrimesImpl.FindPrimes", task.datos, &reply)
		}()
		select {
		case fallo := <-c: //Responde a tiempo
			if fallo != nil { //Si ha habido algun error
				task.fin <- false //Avisa a FindPrimes que ha habido un error
				fmt.Println(worker + "ha habido CRASH")

				go arrancaWorker(worker) //Levantamos el worker de nuevo
				time.Sleep(3 * time.Second)
				fmt.Println(worker + "se ha levantado de nuevo")

				workerCon, err := rpc.DialHTTP("tcp", worker) //Establecemos conexion con el worker asociado
				if err != nil {
					log.Fatal("dialing:", err)
				}
				fmt.Println(worker + " esta otra vez listo y conectado")
			} else {
				copy(*task.resultado, reply) //Copia el resultado
				task.fin <- true             //Avisa que ha acabado correctamente
			}
		case <-time.After(MAX_TIMEOUT): //El worker ha tenido un fallo de tipo DELAY o OMISSION
			task.fin <- false //Avisa a FindPrimes que ha habido un error
			select {
			case <-c: //La respuesta se ha retrasado (DELAY)
				fmt.Println(worker + "ha habido DELAY")
			case <-time.After(TIMEOUT_OMISSION): // Se ha producido una omision (OMISSION)
				fmt.Println(worker + "ha habido OMISSION")

				if opcion_omission == 1 { //OPCION 1: MATAR AL WORKER Y LEVANTARLO DE NUEVO
					go func() {
						var i int
						workerCon.Call("PrimesImpl.Stop", 1, &i) //Intentamos matarlo
					}()
					time.Sleep(time.Second * 2)

					go arrancaWorker(worker) //Hay que volver a levantarlo
					time.Sleep(5 * time.Second)

					workerCon, err := rpc.DialHTTP("tcp", worker) //Establecemos conexion con el worker asociado
					if err != nil {
						log.Fatal("dialing:", err)
					}
					fmt.Println(worker + " esta otra vez listo y conectado")
				} else { //OPCION 2: ESPERAR A QUE RESPONDA
					<-c
					fmt.Println(worker + " recuperado")
				}
			}
		}
	}
}

func (p *PrimesImpl) FindPrimes(dato com.TPInterval, listaRes *[]int) error {
	heAcabado := make(chan bool)                                     //Se crea un canal para saber cuando se ha procesado
	tarea := Tarea{datos: dato, resultado: listaRes, fin: heAcabado} //Se crea la tarea para añadirla a la cola
	var done bool

	for i := 0; i < MAX_INTENTOS && done == false; i++ {
		ColaTareas <- tarea //Ponemos la tarea en la cola
		done = <-heAcabado  //Indica que ha acabado de calcularlo (o que ha habido un error)
	}

	if done {
		return nil
	} else {
		return errors.New("Servidor funciona mal, Problemas con los proxys")
	}
}

func main() {
	ColaTareas = (make(chan Tarea)) //Cola de tareas para que las cojan los proxys
	listWorkers, err := leerWorkers(WORKERS)
	if err != nil {
		log.Fatal("ERROR LECTURA FICHERO:", e)
	}

	for _, worker := range listWorkers {
		fmt.Println("SE HA LANZADO EL PROXY DEL WORKER: " + worker)
		go ProxyPrimes(worker, ColaTareas)
		time.Sleep(100 * time.Millisecond)
	}
}

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
	"fmt"
	"log"
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

	WORKERS = "workers.txt"
	MAX_INTENTOS = 10
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

//Lee los workers de un fichero y los guarda en un vector de strings
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

//Arranca el worker especificado
func arrancaWorker(worker string) {
	//Arranca el worker especificado
	/* TODO */
}

//Proxys del master
func ProxyPrimes(worker string, colaTareas chan Tarea) {

	var reply []int
	c := make(chan error) //Canal de error, para informar del error
	
	go arrancaWorker() //Arrancamos el worker asociado al proxy y le esperamos
	time.Sleep(5*time.Second)

	workerCon, err := rpc.DialHTTP("tcp", worker) //Establecemos conexion con el worker asociado
	if err != nil {
		log.Fatal("dialing:", err)
	}
	fmt.Println(workerDir + " esta listo y conectado")

	/* TODO */

}

//Funcion a la que se conecta el cliente
func (p *PrimesImpl) FindPrimes(dato com.TPInterval, lista *[]int) error {
	heAcabado = make(chan bol) //Se crea un canal para saber cuando se ha procesado
	tarea := Tarea{datos: dato,resultado: lista,fin: heAcabado} //Se crea la tarea para añadirla a la cola
	done bool = false

	for i := 0; i < MAX_INTENTOS && done == false; i++ {
		ColaTareas <- tarea //Ponemos la tarea en la cola
		done = <- heAcabado //Indica que ha acabado de calcularlo (o que ha habido un error)
	}

	if done {
		return nil
	}else{
		return error.New("Servidor funciona mal, Problemas con los proxys")
	}
}

//Funcion Main
func main() {
	if len(os.Args) == 1 {
		ColaTareas = (make(chan Tarea)) //Cola de tareas para que las cojan los proxys
		listWorkers, err := leerWorkers(WORKERS)
		if err != nil {
			log.Fatal("lectura fichero error:", e)
		}

		for _, worker := range listWorkers {
			go ProxyPrimes(worker, ColaTareas)
			fmt.Println("SE HA LANZADO EL PROXY DEL WORKER: " + worker)
			time.Sleep(100 * time.Millisecond)
		}

		primesImpl := new(PrimesImpl)
		primesImpl.delayMaxMilisegundos = 4000
		primesImpl.delayMinMiliSegundos = 2000
		primesImpl.behaviourPeriod = 4
		primesImpl.i = 1
		primesImpl.behaviour = NORMAL

		rpc.Register(primesImpl)
		rpc.HandleHTTP()
		l, e := net.Listen("tcp", os.Args[1])
		if e != nil {
			log.Fatal("listen error:", e)
		}
		http.Serve(l, nil)
	}else{
		fmt.Println("Usage: go run master.go <port>")
	}
}

package main

/*
* AUTOR: Rafael Tolosana Calasanz
* ASIGNATURA: 30221 Sistemas Distribuidos del Grado en Ingeniería Informática
*			Escuela de Ingeniería y Arquitectura - Universidad de Zaragoza
* FECHA: octubre de 2021
* FICHERO: worker.go
* DESCRIPCIÓN: contiene la funcionalidad esencial para realizar los servidores
*				correspondientes la practica 3
 */

import (
	"fmt"
	"log"
	"main/com"
	"math/rand"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	NORMAL   = iota // NORMAL == 0
	DELAY    = iota // DELAY == 1
	CRASH    = iota // CRASH == 2
	OMISSION = iota // IOTA == 3
)

type PrimesImpl struct {
	delayMaxMilisegundos int
	delayMinMiliSegundos int
	behaviourPeriod      int
	behaviour            int
	i                    int
	mutex                sync.Mutex
	can_crash            int
	can_ommit            int
	can_delay            int
}

func isPrime(n int) (foundDivisor bool) {
	foundDivisor = false

	for i := 2; (i < n) && !foundDivisor; i++ {
		foundDivisor = (n%i == 0)
	}
	return !foundDivisor
}

func (p *PrimesImpl) Stop(n int, result *int) error {
	os.Exit(n)
	return nil
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

// PRE: interval.A < interval.B
// POST: FindPrimes devuelve todos los números primos comprendidos en el
// 		intervalo [interval.A, interval.B]
func (p *PrimesImpl) FindPrimes(interval com.TPInterval, primeList *[]int) error {
	p.mutex.Lock()
	if p.i%p.behaviourPeriod == 0 {
		p.behaviourPeriod = rand.Intn(20-2) + 2
		options := rand.Intn(100)
		if options > 90 && p.can_crash == 1 {
			p.behaviour = CRASH
		} else if options > 60 && p.can_delay == 1 {
			p.behaviour = DELAY
		} else if options > 40 && p.can_ommit == 1 {
			p.behaviour = OMISSION
		} else {
			p.behaviour = NORMAL
		}
		p.i = 0
	}
	p.i++
	p.mutex.Unlock()
	switch p.behaviour {
	case DELAY:
		seconds := rand.Intn(p.delayMaxMilisegundos-p.delayMinMiliSegundos) + p.delayMinMiliSegundos
		fmt.Println(seconds)
		//time.Sleep(time.Duration(seconds) * time.Millisecond)
		*primeList = findPrimes(interval)
	case CRASH:
		//os.Exit(1)
	case OMISSION:
		option := rand.Intn(100)
		if option > 65 {
			//time.Sleep(time.Duration(10000) * time.Second)
			*primeList = findPrimes(interval)
		} else {
			*primeList = findPrimes(interval)
		}
	case NORMAL:
		*primeList = findPrimes(interval)
	default:
		*primeList = findPrimes(interval)
	}
	return nil
}

func main() {
	// args: [ip, can_crash(0|1), can_ommit(0|1), can_delay(0|1)]
	if len(os.Args) == 5 {
		time.Sleep(10 * time.Second)
		rand.Seed(time.Now().UnixNano())
		primesImpl := new(PrimesImpl)
		primesImpl.delayMaxMilisegundos = 4000
		primesImpl.delayMinMiliSegundos = 2000
		primesImpl.behaviourPeriod = 4
		primesImpl.i = 1
		primesImpl.behaviour = NORMAL
		primesImpl.can_crash, _ = strconv.Atoi(os.Args[2])
		primesImpl.can_ommit, _ = strconv.Atoi(os.Args[3])
		primesImpl.can_delay, _ = strconv.Atoi(os.Args[4])

		rand.Seed(time.Now().UnixNano())

		rpc.Register(primesImpl)
		rpc.HandleHTTP()
		l, e := net.Listen("tcp", os.Args[1])
		if e != nil {
			log.Fatal("listen error:", e)
		}
		http.Serve(l, nil)
	} else {
		fmt.Println("Usage: go run worker.go <ip:port>")
	}
}

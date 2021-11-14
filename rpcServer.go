package main

import (
	"fmt"
	"main/com"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"time"
)

type Mens struct {
	intervalo com.TPInterval
	reply     *[]int
	done      chan bool
}

type Test struct {
	canal chan Mens
	coord *rpc.Client
}

func checkErrorTest(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func (test *Test) workerTest() {
	for {
		msj := <-test.canal
		time.Sleep(3 * time.Second)
		msj.done <- true
	}
}

func (test *Test) ServerTest(interval com.TPInterval, primeList *[]int) error {
	donda := make(chan bool)
	test.canal <- Mens{interval, primeList, donda}
	<-donda
	*primeList = []int{0, 0, 0, 0}
	return nil
}

const where = "localhost:30000"

func main() {

	// Creamos un canal que pasa las tareas a las gorutines
	test := new(Test)
	test.canal = make(chan Mens)

	// Llama por ssh a los workers y los prepara para escuchar
	for i := 0; i < com.POOL; i++ {
		go test.workerTest()
	}

	// Registro y Creación del RPC
	rpc.Register(test)
	rpc.HandleHTTP()

	// Inicio Escucha
	listener, err := net.Listen("tcp", where)
	checkErrorTest(err)
	defer listener.Close()

	// Sirve petiticiones
	http.Serve(listener, nil)
}

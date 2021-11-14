package main

import (
	"fmt"
	"log"
	"main/com"
	"net/rpc"
	"sync"
)

const to = "localhost:30000"

func pedir(id int, client *rpc.Client, wg *sync.WaitGroup) {
	inter := com.TPInterval{1, 12}
	var reply []int
	client.Call("Test.ServerTest", inter, &reply)
	fmt.Println(id, reply)
	wg.Done()
}

// Comprobación de si RPC es secuencial en servidor
// Respuesta: No lo es, es concurrente
func main() {
	client, err := rpc.DialHTTP("tcp", to)

	if err != nil {
		log.Fatal("dialing:", err)
	}

	var wg *sync.WaitGroup = new(sync.WaitGroup)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go pedir(i, client, wg)
	}
	wg.Wait()
}

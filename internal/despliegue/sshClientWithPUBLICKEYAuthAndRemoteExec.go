// Implementacion de despliegue en ssh de multiples nodos
//
// Unica funcion exportada :
//		func ExecMutipleNodes(cmd string,
//							  hosts []string,
//							  results chan<- string,
//							  privKeyFile string)
//

package despliegue

import (
	"fmt"
	"os"
)

func ExecOneNode(user string, pass string, hostname string,
	rsaPath string, results chan<- string, cmd string) {

	client, err := NewSshClient(
		user,
		hostname,
		22,
		rsaPath,
		pass)

	if err != nil {
		panic(err)
	}

	err = client.RunCommand(cmd)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	results <- "output"
}

// Implementacion de despliegue en ssh de multiples nodos
//
// Unica funcion exportada :
//		func ExecMutipleNodes(cmd string,
//							  hosts []string,
//							  results chan<- string,
//							  privKeyFile string)
//

package despliegue

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

	output, err := client.RunCommand(cmd)
	if err != nil {
		panic(err)
	}

	results <- output
}

package test

import (
	"fmt"
	"raft/internal/comun/constants"
	"raft/internal/despliegue"
	"strconv"
	"testing"
	"time"
)

const (
	//puertos
	PUERTOREPLICA1 = 29001
	PUERTOREPLICA2 = 29002
	PUERTOREPLICA3 = 29003

	// paquete main de ejecutables relativos a PATH previo
	WORKPATH    = "/SSDD/Practica4/"
	EXECREPLICA = "cmd/srvraft/main "

	// comandos completo a ejecutar en máquinas remota con ssh. Ejemplo :
	// 				cd $HOME/raft; go run cmd/srvraft/main.go 127.0.0.1:29001

	// go run testcltvts/main.go 127.0.0.1:29003 127.0.0.1:29001 127.0.0.1:29000

	// Ubicar, en esta constante, nombre de fichero de vuestra clave privada local
	// emparejada con la clave pública en authorized_keys de máquinas remotas

	PRIVKEYFILE = "id_ed25519"
)

// PATH de los ejecutables de modulo golang de servicio de vistas

type testExecution struct {
	user            string
	pass            string
	rsaPath         string
	cmd             string
	canalResultados chan string
}

// TEST primer rango
func TestPrimerasPruebas(t *testing.T) { // (m *testing.M) {
	// <setup code>
	testExec := new(testExecution)

	/*
		fmt.Print("Introduzca el usuario: ")
		fmt.Scanf("%s", &testExec.user)

		fmt.Print("Introduzca la Contraseña: ")
		pass, _ := term.ReadPassword(int(syscall.Stdin))

		testExec.pass = string(pass)
	*/
	testExec.user = "a774248"
	testExec.pass = "Fsw5zw"

	testExec.rsaPath = "/home/" + testExec.user + "/.ssh/id_rsa"
	testExec.cmd = "/home/" + testExec.user + WORKPATH + EXECREPLICA

	// Crear canal de resultados de ejecuciones ssh en maquinas remotas
	testExec.canalResultados = make(chan string, 2000)

	// Run test sequence

	// Test1 : No debería haber ningun primario, si SV no ha recibido aún latidos
	t.Run("T1:ElegirPrimerLider",
		func(t *testing.T) { testExec.soloArranqueYparadaTest1(t) })

	// Test2 : No debería haber ningun primario, si SV no ha recibido aún latidos
	t.Run("T1:ElegirPrimerLider",
		func(t *testing.T) { testExec.elegirPrimerLiderTest2(t) })

	// Test3: tenemos el primer primario correcto
	t.Run("T2:FalloAnteriorElegirNuevoLider",
		func(t *testing.T) { testExec.falloAnteriorElegirNuevoLiderTest3(t) })

	// Test4: Primer nodo copia
	t.Run("T3:EscriturasConcurrentes",
		func(t *testing.T) { testExec.tresOperacionesComprometidasEstable(t) })

	// tear down code
	// eliminar procesos en máquinas remotas
	testExec.stop()
}

func (te *testExecution) stop() {
	// Leer las salidas obtenidos de los comandos ssh ejecutados
	for s := range te.canalResultados {
		fmt.Println(s)
	}
}

func (te *testExecution) startDistributedProcesses(option string) {

	for i, maquina := range constants.MachinesSSH {
		go despliegue.ExecOneNode(te.user, te.pass,
			maquina.Host, te.rsaPath, te.canalResultados, te.cmd+strconv.Itoa(i)+" "+option)

		// dar tiempo para se establezcan las replicas
		time.Sleep(1000 * time.Millisecond)
	}
}

func (te *testExecution) soloArranqueYparadaTest1(t *testing.T) {
	te.startDistributedProcesses("0")
}

func (te *testExecution) elegirPrimerLiderTest2(t *testing.T) {
	te.startDistributedProcesses("1")
}

func (te *testExecution) falloAnteriorElegirNuevoLiderTest3(t *testing.T) {
	te.startDistributedProcesses("2")
}

func (te *testExecution) tresOperacionesComprometidasEstable(t *testing.T) {
	te.startDistributedProcesses("3")
}

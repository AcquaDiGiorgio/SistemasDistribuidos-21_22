package test

import (
	"os"
	"path/filepath"
	"testing"
)

const (
	//hosts
	MAQUINA_LOCAL = "127.0.0.1"
	MAQUINA1      = "127.0.0.1"
	MAQUINA2      = "127.0.0.1"
	MAQUINA3      = "127.0.0.1"

	//puertos
	PUERTOREPLICA1 = "29001"
	PUERTOREPLICA2 = "29002"
	PUERTOREPLICA3 = "29003"

	//nodos replicas
	REPLICA1 = MAQUINA1 + ":" + PUERTOREPLICA1
	REPLICA2 = MAQUINA2 + ":" + PUERTOREPLICA2
	REPLICA3 = MAQUINA3 + ":" + PUERTOREPLICA3

	// paquete main de ejecutables relativos a PATH previo
	EXECREPLICA = "cmd/srvraft/main.go "

	// comandos completo a ejecutar en máquinas remota con ssh. Ejemplo :
	// 				cd $HOME/raft; go run cmd/srvraft/main.go 127.0.0.1:29001

	// go run testcltvts/main.go 127.0.0.1:29003 127.0.0.1:29001 127.0.0.1:29000

	// Ubicar, en esta constante, nombre de fichero de vuestra clave privada local
	// emparejada con la clave pública en authorized_keys de máquinas remotas

	PRIVKEYFILE = "id_ed25519"
)

// PATH de los ejecutables de modulo golang de servicio de vistas
var PATH = filepath.Join(os.Getenv("HOME"), "tmp", "P4", "raft")

var REPLICACMD = "cd " + PATH + "; go run " + EXECREPLICA

type CanalResultados chan string

// TEST primer rango
func TestPrimerasPruebas(t *testing.T) { // (m *testing.M) {
	// <setup code>
	// Crear canal de resultados de ejecuciones ssh en maquinas remotas
	cr := make(CanalResultados, 2000)

	// Run test sequence

	// Test1 : No debería haber ningun primario, si SV no ha recibido aún latidos
	t.Run("T1:ElegirPrimerLider",
		func(t *testing.T) { cr.soloArranqueYparadaTest1(t) })

	// Test2 : No debería haber ningun primario, si SV no ha recibido aún latidos
	t.Run("T1:ElegirPrimerLider",
		func(t *testing.T) { cr.elegirPrimerLiderTest2(t) })

	// Test3: tenemos el primer primario correcto
	t.Run("T2:FalloAnteriorElegirNuevoLider",
		func(t *testing.T) { cr.falloAnteriorElegirNuevoLiderTest3(t) })

	// Test4: Primer nodo copia
	t.Run("T3:EscriturasConcurrentes",
		func(t *testing.T) { cr.tresOperacionesComprometidasEstable(t) })

	// tear down code
	// eliminar procesos en máquinas remotas
	cr.stop()
}

func (cr *CanalResultados) soloArranqueYparadaTest1(t *testing.T) {

}

func (cr *CanalResultados) elegirPrimerLiderTest2(t *testing.T) {

}
func (cr *CanalResultados) falloAnteriorElegirNuevoLiderTest3(t *testing.T) {

}
func (cr *CanalResultados) tresOperacionesComprometidasEstable(t *testing.T) {

}

func (cr *CanalResultados) stop() {

}

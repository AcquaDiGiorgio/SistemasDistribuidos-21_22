package testintegracionraft1

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
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

	// PATH de los ejecutables de modulo golang de servicio de vistas
	PATH = filepath.Join(os.Getenv("HOME"), "tmp", "P4", "raft")

	// paquete main de ejecutables relativos a PATH previo
	EXECREPLICA = "cmd/srvraft/main.go "

	// comandos completo a ejecutar en máquinas remota con ssh. Ejemplo :
	// 				cd $HOME/raft; go run cmd/srvraft/main.go 127.0.0.1:29001

	// go run testcltvts/main.go 127.0.0.1:29003 127.0.0.1:29001 127.0.0.1:29000
	REPLICACMD = "cd " + PATH + "; go run " + EXECREPLICA

	// Ubicar, en esta constante, nombre de fichero de vuestra clave privada local
	// emparejada con la clave pública en authorized_keys de máquinas remotas

	PRIVKEYFILE = "id_ed25519"
)

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

// ---------------------------------------------------------------------
// 
// Canal de resultados de ejecución de comandos ssh remotos
type CanalResultados chan string

func (cr *CanalResultados) stop() {
	close(ts.cmdOutput)

	// Leer las salidas obtenidos de los comandos ssh ejecutados
	for s := range cr {
		fmt.Println(s)
	}
}

// start  gestor de vistas; mapa de replicas y maquinas donde ubicarlos;
// y lista clientes (host:puerto)
func (cr *CanalResultados) startDistributedProcesses(
										replicasMaquinas map[string]string) {

	for replica, maquina := range replicasMaquinas {
			despliegue.ExecMutipleHosts(
			REPLICACMD + " " + replica,
			[]string{maquina}, ts.cmdOutput, PRIVKEYFILE)

		// dar tiempo para se establezcan las replicas
		time.Sleep(1000 * time.Millisecond)
	}
}

//
func (cr *CanalResultados) stopDistributedProcesses(???) {

	// Parar procesos que han sido distribuidos con ssh ??

}

// --------------------------------------------------------------------------
// FUNCIONES DE SUBTESTS

// Se pone en marcha una replica ??
func (cr *CanalResultados) soloArranqueYparadaTest1(t *testing.T) {
	//t.Skip("SKIPPED soloArranqueYparadaTest1")

	fmt.Println(t.Name(), ".....................")

	// Poner en marcha replicas en remoto
	cr.startDistributedProcesses(map[string]string{REPLICA1: MAQUINA1})

	// Parar réplicas alamcenamiento en remoto
	cr.stopDistributedProcesses(??)

	fmt.Println(".............", t.Name(), "Superado")
}

// Primer lider en marcha
func (cr *CanalResultados) ElegirPrimerLiderTest2(t *testing.T) {
	//t.Skip("SKIPPED ElegirPrimerLiderTest2")

	fmt.Println(t.Name(), ".....................")

	// Poner en marcha  3 réplicas Raft
	replicasMaquinas :=
		map[string]string{REPLICA1: MAQUINA1, REPLICA2: MAQUINA2, REPLICA3: 																	MAQUINA3}
	cr.startDistributedProcesses(replicasMaquinas)


	// Se ha elegido lider ?
	fmt.Printf("Probando lider en curso\n")
	pruebaUnLider()


	// Parar réplicas alamcenamiento en remoto
	ts.stopDistributedProcesses(??)

	fmt.Println(".............", t.Name(), "Superado")
}

// Fallo de un primer lider y reeleccion de uno nuevo
func (cr *CanalResultados) FalloAnteriorElegirNuevoLiderTest3(t *testing.T) {
	//t.Skip("SKIPPED FalloAnteriorElegirNuevoLiderTest3")

	fmt.Println(t.Name(), ".....................")

	// Poner en marcha  3 réplicas Raft
	replicasMaquinas :=
		map[string]string{REPLICA1: MAQUINA1, REPLICA2: MAQUINA2, REPLICA3: 																	MAQUINA3}
	cr.startDistributedProcesses(replicasMaquinas)

	fmt.Printf("Lider inicial\n")
	pruebaUnLider()


	// Desconectar lider
	// ???

	fmt.Printf("Comprobar nuevo lider\n")
	pruebaUnLider()
	

	// Parar réplicas almacenamiento en remoto
	ts.stopDistributedProcesses(??)

	fmt.Println(".............", t.Name(), "Superado")
}

// 3 operaciones comprometidas con situacion estable y sin fallos
func (cr *CanalResultados) tresOperacionesComprometidasEstable(t *testing.T) {

	// A completar ???
}

// --------------------------------------------------------------------------
// FUNCIONES DE APOYO
// Comprobar que hay un solo lider
// probar varias veces si se necesitan reelecciones
func pruebaUnLider() int {
	for iters := 0; iters < 10; iters++ {
		time.Sleep(500 * time.Millisecond)
		mapaLideres := make(map[int][]int)
		for i := 0; i < cfg.n; i++ {
			if cfg.connected[i] {
				if _, t, lider := cfg.rafts[i].GetState(); lider {
					mapaLideres[t] = append(mapaLideres[t], i)
				}
			}
		}

		ultimoMandatoConLider := -1
		for t, lideres := range mapaLideres {
			if len(lideres) > 1 {
				cfg.t.Fatalf("mandato %d tiene %d (>1) lideres",
														t, len(lideres))
			}
			if t > ultimoMandatoConLider {
				ultimoMandatoConLider = t
			}
		}

		if len(mapaLideres) != 0 {
			
			return mapaLideres[ultimoMandatoConLider][0]  // Termina
			
		}
	}
	cfg.t.Fatalf("un lider esperado, ninguno obtenido")
	
	return -1   // Termina
}

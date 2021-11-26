// Escribir vuestro código de funcionalidad Raft en este fichero
//

package raft

//
// API
// ===
// Este es el API que vuestra implementación debe exportar
//
// nodoRaft = NuevoNodo(...)
//   Crear un nuevo servidor del grupo de elección.
//
// nodoRaft.Para()
//   Solicitar la parado de un servidor
//
// nodo.ObtenerEstado() (yo, mandato, esLider)
//   Solicitar a un nodo de elección por "yo", su mandato en curso,
//   y si piensa que es el msmo el lider
//
// nodoRaft.SometerOperacion(operacion interface()) (indice, mandato, esLider)

// type AplicaOperacion

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/rpc"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	//CONSTANTES DE LOS LOGS
	kEnableDebugLogs = true           //  false deshabilita los logs
	kLogToStdout     = true           // true: logs -> stdout | flase: logs -> kLogOutputDir
	kLogOutputDir    = "./logs_raft/" // Directorio de salida de los logs

	// CONSTANTES TEMPORALES
	pulseDelay = 100 * time.Millisecond // Delay de cada latido
)

// A medida que el nodo Raft conoce las operaciones de las  entradas de registro
// comprometidas, envía un AplicaOperacion, con cada una de ellas, al canal
// "canalAplicar" (funcion NuevoNodo) de la maquina de estados
type AplicaOperacion struct {
	indice    int // en la entrada de registro
	operacion interface{}
}

// Tipo de dato Go que representa un solo nodo (réplica) de raft
//
type NodoRaft struct {
	// Variables de control y depuración
	mux    sync.Mutex    // Mutex para proteger acceso a estado compartido
	nodos  []*rpc.Client // Conexiones RPC a todos los nodos (réplicas) Raft
	logger *log.Logger   // Logger para depuración

	// Variables que deberían ser iguales entre todos los nodos (sin tener
	// en cuenta posibles errores)
	candidaturaAnterior int
	candidaturaActual   int
	masterActual        int

	// Variables propias de cada Nodo
	yo                        int
	entradas                  []AplicaOperacion
	ultimaEntrada             int
	ultimaEntradaComprometida int

	// Variables únicas del Master actual (si el nodo actual no es master, se
	// deben ignorar sus valores)
	// NONE YET
}

func (nr *NodoRaft) inicializacion() {

	//Canal que indica que se ha terminado el timeout inicial
	ch := make(chan bool)

	//Funcion que le envia un true al nodo
	go func(ch chan bool) {
		time.Sleep(time.Duration(rand.Int()) * time.Millisecond)
		ch <- true
	}(ch)

	select {
		case <- ch: //Mi timeout ha acabado y tengo que presentarme como candidato y iniciar elecciones
			//Inicio elecciones y me presento como candidato 
		case <- //Me llega la candidatura de un candidato 
			//Le voto al candidato
	}
}

func (nr *NodoRaft) activarLogs() {
	nombreNodo := strconv.Itoa(nr.yo) // nodos[yo].String()
	logPrefix := fmt.Sprintf("%s ", nombreNodo)
	if kLogToStdout {
		nr.logger = log.New(os.Stdout, nombreNodo,
			log.Lmicroseconds|log.Lshortfile)
	} else {
		err := os.MkdirAll(kLogOutputDir, os.ModePerm)
		if err != nil {
			panic(err.Error())
		}
		logOutputFile, err := os.OpenFile(fmt.Sprintf("%s/%s.txt",
			kLogOutputDir, logPrefix), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			panic(err.Error())
		}
		nr.logger = log.New(logOutputFile, logPrefix,
			log.Lmicroseconds|log.Lshortfile)
	}
	nr.logger.Println("logger initialized")
}

// Creacion de un nuevo nodo de eleccion
//
// Tabla de <Direccion IP:puerto> de cada nodo incluido a si mismo.
//
// <Direccion IP:puerto> de este nodo esta en nodos[yo]
//
// Todos los arrays nodos[] de los nodos tienen el mismo orden

// canalAplicar es un canal donde, en la practica 5, se recogerán las
// operaciones a aplicar a la máquina de estados. Se puede asumir que
// este canal se consumira de forma continúa.
//
// NuevoNodo() debe devolver resultado rápido, por lo que se deberían
// poner en marcha Gorutinas para trabajos de larga duracion
func NuevoNodo(nodos []*rpc.Client, yo int,
	canalAplicar chan AplicaOperacion) *NodoRaft {

	nr := &NodoRaft{}
	nr.nodos = nodos
	nr.yo = yo

	if kEnableDebugLogs {
		nr.activarLogs()
	} else {
		nr.logger = log.New(ioutil.Discard, "", 0)
	}

	nr.inicializacion()

	return nr
}

// Metodo Para() utilizado cuando no se necesita mas al nodo
//
// Quizas interesante desactivar la salida de depuracion
// de este nodo
//
func (nr *NodoRaft) Para() {

	// Vuestro codigo aqui

}

// Devuelve "yo", mandato en curso y si este nodo cree ser lider
//
func (nr *NodoRaft) ObtenerEstado() (int, int, bool) {
	// No sé si hace lo que pide
	nr.mux.Lock()
	yo := nr.yo
	mandato := nr.candidaturaActual
	esLider := nr.masterActual == nr.yo
	nr.mux.Lock()

	return yo, mandato, esLider
}

// El servicio que utilice Raft (base de datos clave/valor, por ejemplo)
// Quiere buscar un acuerdo de posicion en registro para siguiente operacion
// solicitada por cliente.

// Si el nodo no es el lider, devolver falso
// Sino, comenzar la operacion de consenso sobre la operacion y devolver con
// rapidez
//
// No hay garantia que esta operacion consiga comprometerse n una entrada de
// de registro, dado que el lider puede fallar y la entrada ser reemplazada
// en el futuro.
// Primer valor devuelto es el indice del registro donde se va a colocar
// la operacion si consigue comprometerse.
// El segundo valor es el mandato en curso
// El tercer valor es true si el nodo cree ser el lider
func (nr *NodoRaft) SometerOperacion(operacion interface{}) (int, int, bool) {
	indice := -1
	mandato := -1
	EsLider := true

	for id := range nr.nodos {
		// TODO: Función que introduzca una operación a los seguidores por RPC
		nr.nodos[id].Call("", nil, nil)
	}

	return indice, mandato, EsLider
}

//
// ArgsPeticionVoto
// ===============
// Estructura de argumentos de RPC PedirVoto.
//
type ArgsPeticionVoto struct {
	CandidaturaActual int
	Candidato         int
	ultimaEntrada     int
	ultimaCandidatura int
}

//
// RespuestaPeticionVoto
// ================
// Struct respuesta de RPC PedirVoto
//
//
type RespuestaPeticionVoto struct {
}

//
// PedirVoto
// ===========
//
// Metodo para RPC PedirVoto
//
func (nr *NodoRaft) PedirVoto(args *ArgsPeticionVoto, reply *RespuestaPeticionVoto) {
	// Vuestro codigo aqui
}

// Ejemplo de código enviarPeticionVoto
//
// nodo int -- indice del servidor destino en nr.nodos[]
//
// args *RequestVoteArgs -- argumetnos par la llamada RPC
//
// reply *RequestVoteReply -- respuesta RPC
//
// Los tipos de argumentos y respuesta pasados a CallTimeout deben ser
// los mismos que los argumentos declarados en el metodo de tratamiento
// de la llamada (incluido si son punteros
//
// Si en la llamada RPC, la respuesta llega en un intervalo de tiempo,
// la funcion devuelve true, sino devuelve false
//
// la llamada RPC deberia tener un timout adecuado.
//
// Un resultado falso podria ser causado por una replica caida,
// un servidor vivo que no es alcanzable (por problemas de red ?),
// una petiión perdida, o una respuesta perdida
//
// Para problemas con funcionamiento de RPC, comprobar que la primera letra
// del nombre  todo los campos de la estructura (y sus subestructuras)
// pasadas como parametros en las llamadas RPC es una mayuscula,
// Y que la estructura de recuperacion de resultado sea un puntero a estructura
// y no la estructura misma.
//
func (nr *NodoRaft) enviarPeticionVoto(nodo int, args *ArgsPeticionVoto,
	reply *RespuestaPeticionVoto) (ok bool) {

	return ok
}

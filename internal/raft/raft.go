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
	"net"
	"net/http"
	"net/rpc"
	"os"
	"raft/internal/comun/constants"
	"strconv"
	"sync"
	"time"
)

const (
	//CONSTANTES DE LOS LOGS
	kEnableDebugLogs = false          //  false deshabilita los logs
	kLogToStdout     = true           // true: logs -> stdout | flase: logs -> kLogOutputDir
	kLogOutputDir    = "./logs_raft/" // Directorio de salida de los logs

	// CONSTANTES TEMPORALES
	pulseDelay = 100 * time.Millisecond // Delay de cada latido (10 veces / seg)
)

type EmptyValue struct{}

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
	mux         sync.Mutex    // Mutex para proteger acceso a estado compartido
	nodos       []*rpc.Client // Conexiones RPC a todos los nodos (réplicas) Raft
	logger      *log.Logger   // Logger para depuración
	canalLatido chan bool
	endChan     chan bool
	canalMaster chan bool

	// Variables que deberían ser iguales entre todos los nodos (sin tener
	// en cuenta posibles errores)
	candidaturaAnterior int
	candidaturaActual   int
	masterActual        int

	// Variables propias de cada Nodo
	yo                        int
	votosCandidaturaActual    int
	entradas                  []AplicaOperacion
	ultimaEntrada             int
	ultimaEntradaComprometida int
	periodoLatido             time.Duration
	periodoCandidatura        time.Duration

	// Variables únicas del Master actual (si el nodo actual no es master, se
	// deben ignorar sus valores)
	totalCompromisosUltima int
}

func (nr *NodoRaft) registrarNodo() {

	// Registro y Creación del RPC
	rpc.Register(nr)
	rpc.HandleHTTP()

	// Inicio Escucha
	listener, err := net.Listen("tcp", constants.HOSTS[nr.yo])
	if err != nil {
		os.Exit(1)
	}
	defer listener.Close()

	// Sirve petiticiones
	http.Serve(listener, nil)
}

func (nr *NodoRaft) contactarNodos() {
	for i := 0; i < constants.USERS; i++ {
		if i != nr.yo { // No contacto conmigo
			nodo, err := rpc.DialHTTP("tcp", constants.HOSTS[i])
			if err == nil { // No ha habido error
				nr.nodos = append(nr.nodos, nodo)
				fmt.Println("Nodo Contactado")
			}
		}
	}
}

func (nr *NodoRaft) iniciarComunicacion() {

	nr.contactarNodos()
	time.Sleep(1 * time.Second)

	for {
		if nr.masterActual == nr.yo { // Yo soy el master
			fmt.Println("------------Ejecuto Latidos------------")
			nr.comunicarLatidos()
			time.Sleep(pulseDelay)
		} else {
			select {
			case <-nr.canalLatido: // El master ha respondido a tiempo
				fmt.Println("!¡!¡!¡!¡!¡!¡ ME LLEGA UN LATIDO !¡!¡!¡!¡!¡!¡")
				break

			case <-time.After(nr.periodoLatido): // El master ha tardado mucho
				fmt.Println("¿?¿?¿?¿?¿?¿? EL master no contesta, empiezo candidatura ¿?¿?¿?¿?¿?¿?")
				nr.prepararCandidatura()
				break

			case <-nr.endChan:
				fmt.Println("************* Termino la Ejecucion *************")
				os.Exit(0)
			}
		}
	}
}

func (nr *NodoRaft) prepararCandidatura() {
	select {
	case <-nr.canalMaster:
		break
	case <-time.After(nr.periodoCandidatura):
		args := &ArgsPeticionVoto{
			nr.candidaturaActual, nr.yo, nr.ultimaEntrada, nr.candidaturaAnterior}

		for id := range nr.nodos {
			var respuesta RespuestaPeticionVoto
			ok := nr.enviarPeticionVoto(id, args, &respuesta)
			if ok {
				if !respuesta.VotoGrantizado {
					if nr.candidaturaActual < respuesta.Candidatura {
						nr.candidaturaActual = respuesta.Candidatura
					}
				} else {
					nr.votosCandidaturaActual++
					if nr.votosCandidaturaActual >= constants.USERS/2+1 {
						nr.inicializarMaster()
						nr.candidaturaActual++
						nr.masterActual = nr.yo
					}
				}
			}
		}
	}
}

func (nr *NodoRaft) YaHayMaster(master Estado, emptyReply *EmptyValue) error {
	nr.canalMaster <- true
	nr.masterActual = master.Yo
	nr.candidaturaActual = master.Mandato
	return nil
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
func NuevoNodo(yo int, canalAplicar chan AplicaOperacion) *NodoRaft {

	nr := new(NodoRaft)
	nr.yo = yo
	milis := 1000 + rand.Int()%1500 // Tiempo aleatorio entre 1000 y 2500 ms
	nr.periodoLatido = time.Duration(milis * int(time.Millisecond))
	nr.periodoCandidatura = time.Duration(3 * time.Second)

	if kEnableDebugLogs {
		nr.activarLogs()
	} else {
		nr.logger = log.New(ioutil.Discard, "", 0)
	}

	go nr.registrarNodo()
	nr.iniciarComunicacion()

	return nr
}

// Metodo Para() utilizado cuando no se necesita mas al nodo
//
// Quizas interesante desactivar la salida de depuracion
// de este nodo
//
func (nr *NodoRaft) Para(emptyArgs EmptyValue, emptyReply *EmptyValue) error {
	nr.endChan <- true
	return nil
}

type Estado struct {
	Yo      int
	Mandato int
	EsLider bool
}

//
// Devuelve "yo", mandato en curso y si este nodo cree ser lider
//
func (nr *NodoRaft) ObtenerEstado(emptyArgs EmptyValue, estado *Estado) error {

	nr.mux.Lock()
	estado.Yo = nr.yo
	estado.Mandato = nr.candidaturaActual
	estado.EsLider = nr.masterActual == nr.yo
	nr.mux.Lock()

	return nil
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

type OpASometer struct {
	Indice  int
	Mandato int
	EsLider bool
}

func (nr *NodoRaft) SometerOperacion(operacion *interface{}, oas *OpASometer) error {
	nr.mux.Lock()

	nr.ultimaEntrada++
	esLider := nr.yo == nr.masterActual

	oas.Indice = nr.ultimaEntrada
	oas.Mandato = nr.candidaturaActual
	oas.EsLider = esLider

	nr.mux.Unlock()

	if esLider {
		for id := range nr.nodos {
			var success bool
			nr.nodos[id].Call("NodoRaft.AppendEntries", &operacion, &success)
		}
	}

	return nil
}

func (nr *NodoRaft) AppendEntries(operacion *interface{}, correct *bool) error {
	if nr.yo != nr.masterActual {
		nr.canalLatido <- true // El master sigue vivo

		nr.mux.Lock()
		indice := nr.ultimaEntrada + 1
		nr.entradas[indice] = AplicaOperacion{indice, *operacion}
		nr.ultimaEntrada++
		nr.mux.Unlock()

		*correct = true
	} else {
		*correct = false
	}

	return nil
}

//
// Funciones relacionadas con los latidos entre
// el master y las réplicas
//

func (nr *NodoRaft) RecibirLatido(emptyArgs EmptyValue, ultimaEntrada *AplicaOperacion) error {
	fmt.Println("Fución Recibir latido activada")
	nr.canalLatido <- true
	*ultimaEntrada = nr.entradas[nr.ultimaEntrada]
	return nil
}

func (nr *NodoRaft) comunicarLatidos() {
	var empty EmptyValue
	for id := range nr.nodos {
		var ultimaEntrada AplicaOperacion
		var ch chan *rpc.Call
		nr.nodos[id].Go("NodoRaft.RecibirLatido", empty, &ultimaEntrada, ch)

		if ultimaEntrada.indice > 0 { // Hay alguna entrada
			if nr.entradas[ultimaEntrada.indice].operacion == ultimaEntrada.operacion {
				nr.totalCompromisosUltima++
				if nr.totalCompromisosUltima > (constants.USERS)/2 {
					nr.totalCompromisosUltima = 0
					nr.ultimaEntradaComprometida++
				}
			}
		}
	}
}

func (nr *NodoRaft) inicializarMaster() {

	estado := Estado{nr.yo, nr.candidaturaActual, true}
	var empty EmptyValue
	for id := range nr.nodos {
		nr.nodos[id].Call("NodoRaft.YaHayMaster", estado, &empty)
	}

	nr.totalCompromisosUltima = 0
	// Variables de master a inicializar
}

//
// ArgsPeticionVoto
// ===============
// Estructura de argumentos de RPC PedirVoto.
//
type ArgsPeticionVoto struct {
	CandidaturaActual int
	Candidato         int
	UltimaEntrada     int
	UltimaCandidatura int
}

//
// RespuestaPeticionVoto
// ================
// Struct respuesta de RPC PedirVoto
//
//
type RespuestaPeticionVoto struct {
	Candidatura    int
	VotoGrantizado bool
}

//
// PedirVoto
// ===========
// Metodo para RPC PedirVoto
//
func (nr *NodoRaft) PedirVoto(args *ArgsPeticionVoto, reply *RespuestaPeticionVoto) error {
	acceso := false

	// Aún no ha habido un master
	if nr.candidaturaAnterior == -1 {
		acceso = true

		// El candidato tiene por lo menos con las entradas de este nodo
	} else if args.UltimaCandidatura >= nr.candidaturaAnterior && args.UltimaEntrada >= nr.ultimaEntrada {
		acceso = true
	}

	reply = &RespuestaPeticionVoto{nr.candidaturaActual, acceso}

	return nil
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

	err := nr.nodos[nodo].Call("NodoRaft.PedirVoto", args, &reply)
	return err == nil
}

//
// Función que activa los logs de este nodo
//

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

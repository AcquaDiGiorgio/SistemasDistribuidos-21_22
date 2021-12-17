package raft

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
	"raft/internal/comun/rpctimeout"
	"strconv"
	"sync"
	"time"
)

const (
	//CONSTANTES DE LOS LOGS
	kEnableDebugLogs = true  // false deshabilita los logs
	kLogToStdout     = false // true: logs -> stdout
	// flase: logs -> kLogOutputDir
	kLogOutputDir = "../logs/" // Directorio de salida de los logs

	// CONSTANTES TEMPORALES
	pulseDelay = 100 * time.Millisecond // Delay de cada latido (10 veces / seg)
)

type EmptyValue struct{}

// A medida que el nodo Raft conoce las operaciones de las  entradas de registro
// comprometidas, envía un AplicaOperacion, con cada una de ellas, al canal
// "canalAplicar" (funcion NuevoNodo) de la maquina de estados
type AplicaOperacion struct {
	Indice    int // en la entrada de registro
	Operacion string
}

type entradaIntroducir struct {
	Entrada string
	Quien   int
}

//
// Tipo de dato Go que representa un solo nodo (réplica) de raft
//
type NodoRaft struct {
	// Variables de control y depuración
	mux          sync.Mutex    // Mutex para proteger acceso a estado compartido
	nodos        []*rpc.Client // Conexiones RPC a todos los nodos (réplicas) Raft
	logger       *log.Logger   // Logger para depuración
	canalLatido  chan bool
	endChan      chan bool
	canalAplicar chan entradaIntroducir

	// Variables que deberían ser iguales entre todos los nodos (sin tener
	// en cuenta posibles errores)
	candidaturaActual int
	masterActual      int

	// Varaibles de cada candidatura
	estamosEnCandidatura bool
	soyCandidato         bool
	heVotadoA            int

	// Variables propias de cada Nodo
	yo                        int
	entradas                  []string
	ultimaEntrada             int
	ultimaEntradaComprometida int
	periodoLatido             time.Duration
	periodoCandidatura        time.Duration

	// Variables únicas del Master actual (si el nodo actual no es master, se
	// deben ignorar sus valores)
	indiceUltimaEntradaNodo []int
	comprometiendo          []bool
}

// Creacion de un nuevo nodo de eleccion
//
func NuevoNodo(yo int) *NodoRaft {

	nr := new(NodoRaft)
	nr.yo = yo
	nr.candidaturaActual = -1
	nr.masterActual = -1
	nr.ultimaEntrada = -1
	nr.ultimaEntradaComprometida = -1
	nr.heVotadoA = -1
	nr.soyCandidato = false

	nr.entradas = make([]string, 1000)

	// Tiempo aleatorio entre 500 y 1000 ms
	milis := time.Duration(500 + rand.Int()%6*100)
	nr.periodoLatido = time.Duration(milis * time.Millisecond)

	// Tiempo aleatorio entre 1000 y 3000 ms
	milis = time.Duration(1000 + rand.Int()%21*100)
	nr.periodoCandidatura = time.Duration(milis * time.Millisecond)

	nr.canalLatido = make(chan bool)
	nr.endChan = make(chan bool)

	if kEnableDebugLogs {
		nr.activarLogs()
	} else {
		nr.logger = log.New(ioutil.Discard, "", 0)
	}

	go nr.registrarNodo()
	go nr.iniciarComunicacion()

	return nr
}

//
// Función que prepara a un nodo para la recepción de llamadas
// RPC
//
// Debe ejecutarse como gorutina o el sistema se queda bloqueado
//
func (nr *NodoRaft) registrarNodo() {

	// Registro y Creación del RPC
	rpc.Register(nr)
	rpc.HandleHTTP()

	// Inicio Escucha
	listener, err := net.Listen("tcp", constants.MachinesSSH[nr.yo].Ip)
	if err != nil {
		os.Exit(1)
	}

	// Sirve petiticiones
	http.Serve(listener, nil)
}

//
// Inicializa el vector que guarda los usuarios del sistema
//
// TODOS los nodos deben haber sido registrados antes de que esta función
// se ejecute
//
func (nr *NodoRaft) contactarNodos() {
	nr.mux.Lock()
	for i := 0; i < constants.USERS; i++ {
		if i != nr.yo { // No contacto conmigo
			nodo, err := rpc.DialHTTP("tcp", constants.MachinesSSH[i].Ip)
			if err == nil { // No ha habido error
				nr.nodos = append(nr.nodos, nodo)
				nr.indiceUltimaEntradaNodo =
					append(nr.indiceUltimaEntradaNodo, -1)

				nr.comprometiendo = append(nr.comprometiendo, false)
				nr.logger.Println("Contacto con el nodo", i)
			} else {
				nr.logger.Panicln("ERROR CONTACTO: ", err.Error())
			}
		}
	}
	nr.mux.Unlock()
}

//
// Inicia la ecucha permanente de latidos del master o, por contra,
// si el nodo es master, envia latidos
//
// También incia el periodo de candidatura si no recibe un latido
// a tiempo
//
// Si se van a hacer operaciones externas, se debe ejecutar con una
// gorutina
//
func (nr *NodoRaft) iniciarComunicacion() {

	time.Sleep(2 * time.Second) // Esperamos que todas las réplicas estén registradas
	nr.contactarNodos()         // Contactamos con ellas
	time.Sleep(1 * time.Second)

	for {
		select {
		case <-nr.endChan:
			return

		case <-time.After(10 * time.Millisecond):
			if nr.masterActual == nr.yo { // Yo soy el master
				nr.comunicarLatidos()
				time.Sleep(pulseDelay)
			} else {
				select {
				case <-nr.canalLatido: // El master ha respondido a tiempo
					break

				case <-time.After(nr.periodoLatido): // El master ha tardado mucho
					nr.logger.Println("El master no me contesta, empiezo candidatura")
					nr.prepararCandidatura()
					break
				}
			}
		}
	}
}

//
// Periodo de candidatura de un nodo.
// Se ejecuta tras no recibir un latido del master a tiempo,
//
func (nr *NodoRaft) prepararCandidatura() {
	nr.mux.Lock()
	nr.estamosEnCandidatura = true
	nr.candidaturaActual++
	nr.mux.Unlock()
	for { // Periodo de candidatura
		nr.mux.Lock()
		votosCandidaturaActual := 1
		nr.soyCandidato = false
		nr.mux.Unlock()

		select {
		case <-nr.endChan:
			nr.estamosEnCandidatura = false
			return

		case <-nr.canalLatido: // Alguien se ha convertido en master
			nr.estamosEnCandidatura = false
			nr.heVotadoA = -1
			nr.logger.Println("El nodo", nr.masterActual,
				"se ha hecho master del mandato", nr.candidaturaActual)
			return

		case <-time.After(nr.periodoCandidatura):
			nr.mux.Lock()

			args := ArgsPeticionVoto{
				nr.candidaturaActual,
				nr.yo,
				nr.ultimaEntrada,
				nr.ultimaEntradaComprometida}

			nr.soyCandidato = true
			nr.heVotadoA = -1
			nr.mux.Unlock()

			var sum int //DEBUG
			for id := range nr.nodos {
				var respuesta RespuestaPeticionVoto
				ok := nr.enviarPeticionVoto(id, args, &respuesta)

				if id == nr.yo { //DEBUG
					sum++
				}

				if ok && nr.heVotadoA == -1 { // El nodo no está caído
					nr.logger.Println("El nodo", id+sum,
						"me ha dicho", respuesta.VotoGrantizado)

					if !respuesta.VotoGrantizado { // No me da el voto
						if nr.candidaturaActual < respuesta.Candidatura {
							nr.mux.Lock()
							nr.candidaturaActual = respuesta.Candidatura
							nr.mux.Unlock()
						}
					} else { // Me da el voto
						votosCandidaturaActual++
						if votosCandidaturaActual > constants.USERS/2 {
							nr.inicializarMaster()
							return
						}
					}
				} else {
					nr.logger.Println("El nodo", id+sum,
						"no ha contestado a tiempo")
				}
			}

		}
	}
}

func (nr *NodoRaft) inicializarMaster() {
	nr.masterActual = nr.yo

	nr.logger.Println("Me convierto en master")

	nr.canalAplicar = make(chan entradaIntroducir, 20)
	go nr.introducirEntradas()
}

//
// Metodo que termina la ejecución de un nodo
//
func (nr *NodoRaft) Para(emptyArgs EmptyValue, emptyReply *EmptyValue) error {
	nr.endChan <- true
	return nil
}

type Estado struct {
	Yo                        int
	CandidaturaActual         int
	EsLider                   bool
	MasterActual              int
	Entradas                  []string
	UltimaEntrada             int
	UltimaEntradaComprometida int
	EstamosEnCandidatura      bool
}

//
// Devuelve "yo", mandato en curso y si este nodo cree ser lider
//
func (nr *NodoRaft) ObtenerEstado(emptyArgs EmptyValue, estado *Estado) error {

	nr.mux.Lock()

	estado.Yo = nr.yo
	estado.CandidaturaActual = nr.candidaturaActual
	estado.EsLider = nr.masterActual == nr.yo
	estado.MasterActual = nr.masterActual
	estado.UltimaEntrada = nr.ultimaEntrada
	estado.UltimaEntradaComprometida = nr.ultimaEntradaComprometida
	estado.Entradas = nr.entradas
	estado.EstamosEnCandidatura = nr.estamosEnCandidatura

	nr.mux.Unlock()

	return nil
}

//
// El servicio que utilice Raft (base de datos clave/valor, por ejemplo)
// Quiere buscar un acuerdo de posicion en registro para siguiente operacion
// solicitada por cliente.
//
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
//
type OpASometer struct {
	Indice  int
	Mandato int
	EsLider bool
}

func (nr *NodoRaft) SometerOperacion(operacion string, oas *OpASometer) error {

	nr.logger.Println("Intento someter: ", operacion)
	nr.mux.Lock()

	esLider := nr.yo == nr.masterActual

	if !esLider {
		nr.mux.Unlock()
		return fmt.Errorf("el nodo actual no puede someter operaciones")
	}

	nr.ultimaEntrada++
	oas.Indice = nr.ultimaEntrada
	oas.Mandato = nr.candidaturaActual
	oas.EsLider = esLider

	nr.mux.Unlock()

	if esLider {
		nr.mux.Lock()
		nr.entradas[nr.ultimaEntrada] = operacion
		nr.mux.Unlock()
	}

	return nil
}

//
// Función RPC que recibe un nodo cuando un master le pide introducir
// una entrada
//
// Recibe la operación a someter
// Devuelve si la ha introducido
//
func (nr *NodoRaft) AppendEntries(operacion AplicaOperacion,
	correct *bool) error {

	nr.logger.Println("El master, me pide meter entradas")
	if nr.yo != nr.masterActual {
		nr.canalLatido <- true // El master sigue vivo

		nr.mux.Lock()

		nr.ultimaEntrada++
		nr.logger.Println("Op Recibida:\n\tentrada[", operacion.Indice,
			"]", operacion.Operacion)
		nr.entradas[operacion.Indice] = operacion.Operacion

		nr.mux.Unlock()

		*correct = true
	} else {
		*correct = false
	}

	return nil
}

type ArgsLatido struct {
	MasterActual  int
	MandatoActual int
	Comprometidas int
}

//
// Funciones relacionadas con los latidos entre
// el master y las réplicas
//
func (nr *NodoRaft) RecibirLatido(args ArgsLatido, ultimaEntrada *int) error {
	nr.canalLatido <- true

	nr.mux.Lock()
	if args.MandatoActual >= nr.candidaturaActual {
		nr.candidaturaActual = args.MandatoActual
		nr.masterActual = args.MasterActual
		nr.ultimaEntradaComprometida = args.Comprometidas
		*ultimaEntrada = nr.ultimaEntrada
	}
	nr.mux.Unlock()

	return nil
}

func (nr *NodoRaft) comunicarLatidos() {
	args := ArgsLatido{nr.yo, nr.candidaturaActual,
		nr.ultimaEntradaComprometida}

	var respuesta int

	for id := range nr.nodos {
		err := rpctimeout.CallTimeout(nr.nodos[id],
			"NodoRaft.RecibirLatido", args, &respuesta, time.Second)

		if err == nil {
			nr.indiceUltimaEntradaNodo[id] = respuesta
		}

		nr.comprobarCompromiso()

		// Hay alguna entrada en mi nodo
		if nr.ultimaEntrada != -1 {
			// El nodo con quien contacto no tiene las mismas entradas que yo
			if nr.indiceUltimaEntradaNodo[id] != nr.ultimaEntrada {
				nr.logger.Println("El nodo", id, "tiene",
					nr.indiceUltimaEntradaNodo[id],
					"entradas y vamos por la", nr.ultimaEntrada)

				indice := nr.indiceUltimaEntradaNodo[id] + 1

				nr.mux.Lock()
				comprometiendoEntrada := nr.comprometiendo[id]
				nr.mux.Unlock()

				if !comprometiendoEntrada {
					entrada := entradaIntroducir{nr.entradas[indice], id}
					nr.canalAplicar <- entrada
				}
			}
		}
	}
}

func (nr *NodoRaft) introducirEntradas() {
	for {
		entradaIntroducir := <-nr.canalAplicar
		var correcto bool

		nr.mux.Lock()

		nr.comprometiendo[entradaIntroducir.Quien] = true

		operacion := AplicaOperacion{
			nr.indiceUltimaEntradaNodo[entradaIntroducir.Quien] + 1,
			entradaIntroducir.Entrada}

		nr.mux.Unlock()

		rpctimeout.CallTimeout(nr.nodos[entradaIntroducir.Quien],
			"NodoRaft.AppendEntries", operacion, &correcto, time.Second)

		nr.mux.Lock()
		nr.comprometiendo[entradaIntroducir.Quien] = false
		nr.mux.Unlock()
	}
}

func (nr *NodoRaft) comprobarCompromiso() {
	nr.mux.Lock()
	nextEntry := nr.ultimaEntradaComprometida + 1
	nr.mux.Unlock()

	graterOrEqual := 1

	for _, val := range nr.indiceUltimaEntradaNodo {
		if val >= nextEntry {
			graterOrEqual++
			if graterOrEqual > constants.USERS/2 {
				nr.logger.Println("Se compromete la entrada número", val)
				nr.mux.Lock()
				nr.ultimaEntradaComprometida++
				nr.mux.Unlock()
				return
			}
		}
	}
}

//
// ArgsPeticionVoto
// ===============
// Estructura de argumentos de RPC PedirVoto.
//
type ArgsPeticionVoto struct {
	CandidaturaActual         int
	Candidato                 int
	UltimaEntrada             int
	UltimaEntradaComprometida int
}

//
// RespuestaPeticionVoto
// ================
// Struct respuesta de RPC PedirVoto
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
func (nr *NodoRaft) PedirVoto(args ArgsPeticionVoto,
	reply *RespuestaPeticionVoto) error {

	acceso := false

	// Si soy candidato o ya he votado, no doy el voto
	if nr.soyCandidato || nr.heVotadoA != -1 {
		*reply = RespuestaPeticionVoto{nr.candidaturaActual, acceso}
		return nil
	}

	// Aún no ha habido un master
	if nr.candidaturaActual == -1 {
		nr.mux.Lock()
		nr.heVotadoA = args.Candidato
		nr.mux.Unlock()
		acceso = true

		// El candidato tiene por lo menos la misma cantidad de entradas
		// y de entradas comprometidas que este nodo
	} else if args.CandidaturaActual >= nr.candidaturaActual &&
		args.UltimaEntrada >= nr.ultimaEntrada &&
		args.UltimaEntradaComprometida >= nr.ultimaEntradaComprometida {

		nr.mux.Lock()
		nr.heVotadoA = args.Candidato
		nr.mux.Unlock()
		acceso = true
	}

	*reply = RespuestaPeticionVoto{nr.candidaturaActual, acceso}

	return nil
}

//
// nodo int 				-- indice del servidor destino en nr.nodos[]
// args *RequestVoteArgs 	-- argumetnos par la llamada RPC
// reply *RequestVoteReply 	-- respuesta RPC
//
// Si en la llamada RPC, la respuesta llega en un intervalo de tiempo,
// la funcion devuelve true, sino devuelve false
//
// Un resultado falso podria ser causado por una replica caida,
// un servidor vivo que no es alcanzable,
// una petición perdida, o una respuesta perdida
//
func (nr *NodoRaft) enviarPeticionVoto(nodo int, args ArgsPeticionVoto,
	reply *RespuestaPeticionVoto) (ok bool) {

	err := rpctimeout.CallTimeout(nr.nodos[nodo], "NodoRaft.PedirVoto",
		args, &reply, time.Second)

	return err == nil
}

//
// Función que activa los logs de este nodo
//
func (nr *NodoRaft) activarLogs() {
	nombreNodo := strconv.Itoa(nr.yo)
	logPrefix := fmt.Sprintf("Nodo_%s ", nombreNodo)
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

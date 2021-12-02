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
	kEnableDebugLogs = false      //  false deshabilita los logs
	kLogToStdout     = true       // true: logs -> stdout | flase: logs -> kLogOutputDir
	kLogOutputDir    = "../logs/" // Directorio de salida de los logs

	// CONSTANTES TEMPORALES
	pulseDelay = 100 * time.Millisecond // Delay de cada latido (10 veces / seg)
)

type EmptyValue struct{}

// A medida que el nodo Raft conoce las operaciones de las  entradas de registro
// comprometidas, envía un AplicaOperacion, con cada una de ellas, al canal
// "canalAplicar" (funcion NuevoNodo) de la maquina de estados
type AplicaOperacion struct {
	Indice    int // en la entrada de registro
	Operacion interface{}
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

	// Variables que deberían ser iguales entre todos los nodos (sin tener
	// en cuenta posibles errores)
	candidaturaAnterior int
	candidaturaActual   int
	masterActual        int
	soyCandidato        bool

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
	nr.candidaturaActual = -1
	nr.masterActual = -1
	nr.candidaturaAnterior = -1
	nr.totalCompromisosUltima = -1
	nr.ultimaEntrada = -1
	nr.ultimaEntradaComprometida = -1
	nr.votosCandidaturaActual = 0
	nr.soyCandidato = false

	milis := time.Duration(1000 + rand.Int()%11*100) // Tiempo aleatorio entre 1000 y 2500 ms
	nr.periodoLatido = time.Duration(milis * time.Millisecond)

	milis = time.Duration(1000 + rand.Int()%11*100) // Tiempo aleatorio entre 1000 y 2500 ms
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
	listener, err := net.Listen("tcp", constants.MachinesLocal[nr.yo].Ip)
	if err != nil {
		os.Exit(1)
	}
	defer listener.Close()

	// Sirve petiticiones
	http.Serve(listener, nil)
}

//
// Inicializa el vector que guarda los usuarios del sistema
//
// Los nodos deben haber sido registrados antes de que esta función
// se ejecute
//
func (nr *NodoRaft) contactarNodos() {
	for i := 0; i < constants.USERS; i++ {
		if i != nr.yo { // No contacto conmigo
			nodo, err := rpc.DialHTTP("tcp", constants.MachinesLocal[i].Ip)
			if err == nil { // No ha habido error
				nr.nodos = append(nr.nodos, nodo)
				nr.logger.Println("Contacto con el nodo", i)
			}
		}
	}
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

	nr.contactarNodos()
	time.Sleep(1 * time.Second)

	go func() {
		<-nr.endChan
		nr.logger.Println("Termino la ejecución")
		os.Exit(0)
	}()

	for {
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

//
// Periodo de candidatura de un nodo.
// Se ejecuta tras no recibir un latido del master a tiempo,
//
func (nr *NodoRaft) prepararCandidatura() {
	nr.mux.Lock()
	nr.candidaturaActual++
	nr.mux.Unlock()
	for {
		nr.mux.Lock()
		nr.votosCandidaturaActual = 1
		nr.soyCandidato = false
		nr.mux.Unlock()
		select {
		case <-nr.canalLatido: // Alguien se ha convertido en master
			nr.logger.Println("Termina la candidatura", nr.candidaturaActual)
			return

		case <-time.After(nr.periodoCandidatura):
			nr.mux.Lock()
			args := ArgsPeticionVoto{
				nr.candidaturaActual, nr.yo, nr.ultimaEntrada, nr.candidaturaAnterior}
			nr.soyCandidato = true
			nr.mux.Unlock()

			var sum int //DEBUG
			for id := range nr.nodos {
				var respuesta RespuestaPeticionVoto
				ok := nr.enviarPeticionVoto(id, args, &respuesta)

				if id == nr.yo { //DEBUG
					sum++
				}

				if ok { // El nodo no está caído
					nr.logger.Println("El nodo", id+sum, "me ha dicho", respuesta.VotoGrantizado)

					if !respuesta.VotoGrantizado { // No me da el voto
						if nr.candidaturaActual < respuesta.Candidatura {
							nr.mux.Lock()
							nr.candidaturaActual = respuesta.Candidatura
							nr.mux.Unlock()
						}
					} else { // Me da el voto
						nr.votosCandidaturaActual++
						if nr.votosCandidaturaActual >= constants.USERS/2+1 {
							nr.inicializarMaster()
							return
						}
					}
				} else {
					nr.logger.Println("El nodo", id+sum, "ha dado error de conexion")
				}
			}
		}
	}
}

func (nr *NodoRaft) inicializarMaster() {
	nr.logger.Println("Me convierto en master")

	nr.mux.Lock()
	nr.masterActual = nr.yo
	nr.mux.Unlock()

	var empty EmptyValue
	var estado Estado

	nr.ObtenerEstado(empty, &estado)

	for id := range nr.nodos {
		nr.nodos[id].Call("NodoRaft.YaHayMaster", estado, &empty)
	}

	nr.totalCompromisosUltima = 0
	// Variables de master a inicializar
}

//
// Avisa que en el periodo de candidatura actual, ya se ha elegido a alguien como
// master.
//
// La información de quién es va guardada en master.
//
func (nr *NodoRaft) YaHayMaster(master Estado, emptyReply *EmptyValue) error {
	nr.canalLatido <- true

	nr.mux.Lock()
	nr.masterActual = master.Yo
	nr.candidaturaActual = master.Mandato
	nr.mux.Unlock()

	nr.logger.Println("El nodo", master.Yo, "se ha hecho master del mandato", master.Mandato)
	return nil
}

//
// Metodo que termina la ejecución de un nodo
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

	nr.mux.Unlock()

	return nil
}

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

	ao := AplicaOperacion{nr.ultimaEntrada, operacion}

	if esLider {
		nr.mux.Lock()
		nr.entradas = append(nr.entradas, ao)
		nr.mux.Unlock()
		for id := range nr.nodos {
			var success bool
			nr.nodos[id].Call("NodoRaft.AppendEntries", operacion, &success)
			nr.logger.Println("Recibo: ", success)
		}
	}

	return nil
}

func (nr *NodoRaft) AppendEntries(operacion string, correct *bool) error {
	nr.logger.Println("Me piden meter entradas", nr.yo, nr.masterActual)
	if nr.yo != nr.masterActual {
		nr.canalLatido <- true // El master sigue vivo

		nr.mux.Lock()

		nr.ultimaEntrada++
		nr.logger.Println("Op Recibida:\n\tentrada[", nr.ultimaEntrada, "]", operacion)
		nr.entradas = append(nr.entradas, AplicaOperacion{nr.ultimaEntrada, operacion})

		nr.mux.Unlock()

		*correct = true
	} else {
		*correct = false
	}

	return nil
}

type ArgsLatido struct {
	MasterActual              int
	MandatoActual             int
	ultimaEntradaComprometida int
}

//
// Funciones relacionadas con los latidos entre
// el master y las réplicas
//
func (nr *NodoRaft) RecibirLatido(args ArgsLatido, ultimaEntrada *AplicaOperacion) error {
	nr.canalLatido <- true

	if args.MandatoActual >= nr.candidaturaActual {
		nr.candidaturaAnterior = nr.candidaturaActual
		nr.candidaturaActual = args.MandatoActual
	}

	nr.ultimaEntradaComprometida = args.ultimaEntradaComprometida

	if nr.ultimaEntrada != -1 {
		*ultimaEntrada = nr.entradas[nr.ultimaEntrada]
	} else {
		*ultimaEntrada = AplicaOperacion{-1, nil}
	}

	return nil
}

func (nr *NodoRaft) comunicarLatidos() {
	args := ArgsLatido{nr.yo, nr.candidaturaActual, nr.ultimaEntradaComprometida}

	for id := range nr.nodos {
		var ultimaEntrada AplicaOperacion
		nr.nodos[id].Call("NodoRaft.RecibirLatido", args, &ultimaEntrada)

		if nr.ultimaEntrada != -1 { // Hay alguna entrada en mi nodo
			if ultimaEntrada.Indice != -1 { // El nodo con quien contacto tiene alguna entrada
				if nr.entradas[ultimaEntrada.Indice].Operacion == ultimaEntrada.Operacion {
					nr.totalCompromisosUltima++
					if nr.totalCompromisosUltima > (constants.USERS)/2 {
						nr.totalCompromisosUltima = 0
						nr.ultimaEntradaComprometida++
					}
				}
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
type RespuestaPeticionVoto struct {
	Candidatura    int
	VotoGrantizado bool
}

//
// PedirVoto
// ===========
// Metodo para RPC PedirVoto
//
func (nr *NodoRaft) PedirVoto(args ArgsPeticionVoto, reply *RespuestaPeticionVoto) error {
	acceso := false

	if nr.soyCandidato {
		*reply = RespuestaPeticionVoto{nr.candidaturaActual, acceso}
		return nil
	}

	// Aún no ha habido un master
	if nr.candidaturaAnterior == -1 {
		acceso = true

		// El candidato tiene por lo menos con las entradas de este nodo
	} else if args.UltimaCandidatura >= nr.candidaturaAnterior && args.UltimaEntrada >= nr.ultimaEntrada {
		acceso = true
	}

	*reply = RespuestaPeticionVoto{nr.candidaturaActual, acceso}

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
// una petición perdida, o una respuesta perdida
//
// Para problemas con funcionamiento de RPC, comprobar que la primera letra
// del nombre  todo los campos de la estructura (y sus subestructuras)
// pasadas como parametros en las llamadas RPC es una mayuscula,
// Y que la estructura de recuperacion de resultado sea un puntero a estructura
// y no la estructura misma.
//
func (nr *NodoRaft) enviarPeticionVoto(nodo int, args ArgsPeticionVoto,
	reply *RespuestaPeticionVoto) (ok bool) {

	err := nr.nodos[nodo].Call("NodoRaft.PedirVoto", args, &reply)
	return err == nil
}

//
// Función que activa los logs de este nodo
//
func (nr *NodoRaft) activarLogs() {
	nombreNodo := strconv.Itoa(nr.yo) // nodos[yo].String()
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

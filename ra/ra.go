/*
* AUTOR: Rafael Tolosana Calasanz
* ASIGNATURA: 30221 Sistemas Distribuidos del Grado en Ingeniería Informática
*			Escuela de Ingeniería y Arquitectura - Universidad de Zaragoza
* FECHA: septiembre de 2021
* FICHERO: ricart-agrawala.go
* DESCRIPCIÓN: Implementación del algoritmo de Ricart-Agrawala Generalizado en Go
 */
package ra

import (
	"fmt"
	"main/fm"
	"main/ms"
	"strconv"
	"sync"

	"github.com/DistributedClocks/GoVector/govec"
)

type Request struct {
	Clock      int
	Pid        int
	Actor      string
	packedData []byte
}

type Reply struct {
	packedData []byte
}

type RASharedDB struct { // TODO: Completar
	//CONSTANT
	Me    int
	Actor string
	//INTEGER
	OurSeqNum int
	HigSeqNum int
	OutRepCnt int
	//BOOLEAN
	ReqCS   bool
	RepDefd [N]bool
	//BINARY SEMAPHORE
	Mutex sync.Mutex
	//OTHER
	ms     *ms.MessageSystem
	done   chan bool
	chrep  chan bool
	logger *govec.GoLog
}

const N = 2

// logger.PrepareSend("Sending Message", messagePayload, govec.GetDefaultLogOptions())
// logger.UnpackReceive("Receiving Message", vectorClockMessage, &messagePayload, govec.GetDefaultLogOptions())
// logger.LogLocalEvent("Example Complete", govec.GetDefaultLogOptions())

func New(me int, usersFile string, actor string) *RASharedDB {
	messageTypes := []ms.Message{Request{}, Reply{}}
	msgs := ms.New(me, usersFile, messageTypes)

	logger := govec.InitGoVector(actor+"-"+strconv.Itoa(me), "./Log"+strconv.Itoa(me), govec.GetDefaultConfig())

	ra := RASharedDB{me, actor, 0, 0, N, false, [N]bool{}, sync.Mutex{}, &msgs, make(chan bool), make(chan bool), logger}

	for i := 0; i < N; i++ {
		ra.RepDefd[i] = false
	}

	go func() {
		for {
			select {
			case <-ra.done:
				return
			default:
				dato := ra.ms.Receive()
				println("Recibo Mensaje")
				switch dato.(type) {
				case Request: // Recibo una petición de acceso
					println("Es un request")
					pet := dato.(Request)
					ra.HigSeqNum = intMax(ra.HigSeqNum, pet.Clock)
					var reciboPeticion = []byte("alguien pide la SC")
					var envioRespuesta = []byte("permito el acceso a la SC")
					ra.logger.UnpackReceive("Recibo Peticion Acceso", pet.packedData, &reciboPeticion, govec.GetDefaultLogOptions())

					ra.Mutex.Lock()
					Defer_It := ra.ReqCS &&
						(pet.Clock > ra.OurSeqNum ||
							(pet.Clock == ra.OurSeqNum && pet.Pid > ra.Me) ||
							(pet.Clock > ra.OurSeqNum && exclude(ra.Actor, pet.Actor)) ||
							(pet.Clock == ra.OurSeqNum && pet.Pid > ra.Me && exclude(ra.Actor, pet.Actor)))
					ra.Mutex.Unlock()

					if Defer_It {
						ra.RepDefd[pet.Pid-1] = true
					} else {
						pD := ra.logger.PrepareSend("Permito Acceso Def", envioRespuesta, govec.GetDefaultLogOptions())
						ra.ms.Send(pet.Pid, Reply{pD}) // ESTO DA PROBLEMAS
					}
					fmt.Println("LOG: ", pet.Pid, pet.Clock, pet.Actor, Defer_It, ra.OurSeqNum, ra.HigSeqNum)

				case Reply: // Recibo una respuesta
					println("Es un reply")
					rep := dato.(Reply)
					var reciboRespuesta = []byte("recibo el acceso a la SC")
					ra.logger.UnpackReceive("Recibo Permiso", rep.packedData, &reciboRespuesta, govec.GetDefaultLogOptions())
					if ra.ReqCS {
						ra.OutRepCnt--
						if ra.OutRepCnt == 0 {
							ra.chrep <- true
						}
					}
				default: // Comorl, que esh lo que é rechibido
					fmt.Printf("WTF %T\n", dato)
					return
				}
			}
		}
	}()

	return &ra
}

//Pre: Verdad
//Post: Realiza  el  PreProtocol  para el  algoritmo de
//      Ricart-Agrawala Generalizado
func (ra *RASharedDB) PreProtocol() {
	ra.Mutex.Lock()
	ra.ReqCS = true
	ra.OurSeqNum++
	ra.Mutex.Unlock()
	ra.OutRepCnt--

	for j := 1; j <= N; j++ {
		if j == ra.Me {
			continue
		}
		var envioPeticion = []byte("pido SC")
		pD := ra.logger.PrepareSend("Envio Peticion Acceso", envioPeticion, govec.GetDefaultLogOptions())
		ra.ms.Send(j, Request{ra.OurSeqNum, ra.Me, ra.Actor, pD})
	}

	<-ra.chrep
}

//Pre: Verdad
//Post: Realiza  el  PostProtocol  para el  algoritmo de
//      Ricart-Agrawala Generalizado
func (ra *RASharedDB) PostProtocol() {
	ra.ReqCS = false
	for j := 1; j <= N; j++ {
		if ra.RepDefd[j-1] {
			ra.RepDefd[j-1] = false
			var envioRespuesta = []byte("permito el acceso a la SC")
			pD := ra.logger.PrepareSend("Envio Peticion Acceso PP", envioRespuesta, govec.GetDefaultLogOptions())
			ra.ms.Send(j, Reply{pD})
		}
	}
	ra.OutRepCnt = N
}

func (ra *RASharedDB) Stop() {
	ra.ms.Stop()
	ra.done <- true
}

func (ra *RASharedDB) AccedoSC(entrada string) {
	ra.logger.LogLocalEvent("Accedo a la SC", govec.GetDefaultLogOptions())
	if ra.Actor == "lector" {
		fm.LeerFichero()
	} else {
		fm.EscribirFichero()
	}
}

// Funciones que no tienen que ver con el algoritmo

func intMax(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func exclude(Actor1 string, Actor2 string) bool {
	return !(Actor1 == "lector" && Actor2 == "lector")
}

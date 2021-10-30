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
	"main/ms"
	"sync"

	"github.com/DistributedClocks/GoVector/govec"
)

type Request struct {
	Clock      int
	Pid        int
	Actor      string
	PackedData []byte
}

type Reply struct {
	PackedData []byte
}

type RASharedDB struct {
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
	Logger *govec.GoLog
}

const N = 4

func New(me int, usersFile string, actor string, logger *govec.GoLog) *RASharedDB {
	messageTypes := []ms.Message{Request{}, Reply{}}
	msgs := ms.New(me, usersFile, messageTypes)

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
				fmt.Println("LOG: Recibo Mensaje")
				switch msg := dato.(type) {
				case Request: // Recibo una petición de acceso

					ra.HigSeqNum = intMax(ra.HigSeqNum, msg.Clock)

					var reciboPeticion []byte
					ra.Logger.UnpackReceive("Recibo Peticion Acceso", msg.PackedData, &reciboPeticion, govec.GetDefaultLogOptions())

					ra.Mutex.Lock()
					Defer_It := ra.ReqCS &&
						(msg.Clock > ra.OurSeqNum ||
							(msg.Clock == ra.OurSeqNum && msg.Pid > ra.Me) ||
							(msg.Clock > ra.OurSeqNum && exclude(ra.Actor, msg.Actor)) ||
							(msg.Clock == ra.OurSeqNum && msg.Pid > ra.Me && exclude(ra.Actor, msg.Actor)))
					ra.Mutex.Unlock()

					if Defer_It {
						ra.RepDefd[msg.Pid-1] = true
					} else {
						var envioRespuesta = []byte("Permito el acceso a la SC")
						pd := ra.Logger.PrepareSend("Envio Permiso de Acceso", envioRespuesta, govec.GetDefaultLogOptions())
						ra.ms.Send(msg.Pid, Reply{pd})
					}

				case Reply: // Recibo una respuesta

					var reciboRespuesta []byte
					ra.Logger.UnpackReceive("Recibo Permiso de Acceso", msg.PackedData, &reciboRespuesta, govec.GetDefaultLogOptions())

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
	ra.OutRepCnt--
	ra.Mutex.Unlock()

	for j := 1; j <= N; j++ {
		if j == ra.Me {
			continue
		}

		var envioPeticion = []byte("pido SC")
		pD := ra.Logger.PrepareSend("Envio Peticion Acceso", envioPeticion, govec.GetDefaultLogOptions())
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

			var envioRespuesta = []byte("Permito el acceso a la SC")
			pD := ra.Logger.PrepareSend("Envio Permiso de Acceso", envioRespuesta, govec.GetDefaultLogOptions())
			ra.ms.Send(j, Reply{pD})
		}
	}
	ra.OutRepCnt = N
}

func (ra *RASharedDB) Stop() {
	ra.ms.Stop()
	ra.done <- true
}

func intMax(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func exclude(Actor1 string, Actor2 string) bool {
	return !(Actor1 == "lector" && Actor2 == "lector")
}

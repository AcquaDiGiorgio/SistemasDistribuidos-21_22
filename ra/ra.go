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
	PackedData []byte // Información del proceso que ha enviado la petición para poder crear el dibujo con ShiViz
}

type Reply struct {
	PackedData []byte // Información del proceso que ha enviado la respuesta para poder crear el dibujo con ShiViz
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
			case <-ra.done: // Ya hemos terminado el algoritomo
				return
			default: // Aún estamos a la espera de mensajes
				dato := ra.ms.Receive()
				fmt.Println("LOG: Recibo Mensaje")
				switch msg := dato.(type) {
				case Request: // Recibo una petición de acceso

					ra.HigSeqNum = intMax(ra.HigSeqNum, msg.Clock)

					var reciboPeticion []byte
					ra.Logger.UnpackReceive("Recibo Peticion Acceso", msg.PackedData, &reciboPeticion, govec.GetDefaultLogOptions())

					// Accedo a la lectura de las variables de forma atómica
					ra.Mutex.Lock()
					Defer_It := ra.ReqCS &&
						(msg.Clock > ra.OurSeqNum ||
							(msg.Clock == ra.OurSeqNum && msg.Pid > ra.Me) ||
							(msg.Clock > ra.OurSeqNum && exclude(ra.Actor, msg.Actor)) ||
							(msg.Clock == ra.OurSeqNum && msg.Pid > ra.Me && exclude(ra.Actor, msg.Actor)))
					ra.Mutex.Unlock()

					if Defer_It { // Si yo tengo prioridad
						// Le defiero la respuesta para cuando yo ya haya terminado
						ra.RepDefd[msg.Pid-1] = true
					} else { // Si él tiene prioridad
						// Le permito el acceso
						var envioRespuesta = []byte("Permito el acceso a la SC")
						pd := ra.Logger.PrepareSend("Envio Permiso de Acceso", envioRespuesta, govec.GetDefaultLogOptions())
						ra.ms.Send(msg.Pid, Reply{pd})
					}

				case Reply: // Recibo una respuesta

					var reciboRespuesta []byte
					ra.Logger.UnpackReceive("Recibo Permiso de Acceso", msg.PackedData, &reciboRespuesta, govec.GetDefaultLogOptions())

					// Si quiero entrar a la SC (Checkeo por seguridad)
					if ra.ReqCS {
						ra.OutRepCnt--         // 1 persona menos a la que esperar respuesta
						if ra.OutRepCnt == 0 { // Si no hay a quien esperar
							ra.chrep <- true // Puedo acceder a la SC
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
	// Escritura en las variables en exclusión mutua
	ra.Mutex.Lock()
	ra.ReqCS = true
	ra.OurSeqNum++
	ra.OutRepCnt--
	ra.Mutex.Unlock()

	// Para todos los usuarios del sistema de mensajes
	for j := 1; j <= N; j++ {
		// Si soy yo, no hago nada
		if j == ra.Me {
			continue
		}

		// Si no soy yo, pido acceso a la SC
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
	// Ya no deseo entrar a la SC
	ra.ReqCS = false
	// A todos los clientes del Sistema de Mensajes
	for j := 1; j <= N; j++ {
		// Si los he diferido
		if ra.RepDefd[j-1] {
			ra.RepDefd[j-1] = false

			// Les permito el acceso a la SC
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

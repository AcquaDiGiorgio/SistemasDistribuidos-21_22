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
	"main/ms"
	"sync"
)

type Request struct {
	Clock  int
	Pid    int
	Accion string
}

type Reply struct{}

type RASharedDB struct { // TODO: Completar
	//CONSTANT
	Me     int
	Accion string
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
	ms    *ms.MessageSystem
	done  chan bool
	chrep chan bool
}

const N = 2

func New(me int, usersFile string, accion string) *RASharedDB {
	messageTypes := []ms.Message{Request{}, Reply{}}
	msgs := ms.New(me, usersFile, messageTypes)
	ra := RASharedDB{me, accion, 0, 0, 0, false, [N]bool{}, sync.Mutex{}, &msgs, make(chan bool), make(chan bool)}

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
				switch dato.(type) {
				case Request: // Recibo una petición de acceso
					pet := dato.(Request)
					ra.HigSeqNum = intMax(ra.HigSeqNum, pet.Clock)

					ra.Mutex.Lock()
					Defer_It := ra.ReqCS &&
						(pet.Clock > ra.OurSeqNum ||
							(pet.Clock == ra.OurSeqNum && pet.Pid > ra.Me) ||
							exclude(ra.Accion, pet.Accion))
					ra.Mutex.Unlock()

					if Defer_It {
						ra.RepDefd[pet.Pid-1] = true
					} else {
						ra.ms.Send(pet.Pid, Reply{})
					}

				case Reply: // Recibo una respuesta
					if ra.ReqCS {
						ra.OutRepCnt--
						if ra.OutRepCnt == 0 {
							ra.chrep <- true
						}
					}
				default: // Comorl, que esh lo que é rechibido
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
		ra.ms.Send(j, Request{ra.OurSeqNum, j, ra.Accion})
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
			ra.ms.Send(j, Reply{})
		}
	}
	ra.OutRepCnt = N
}

func (ra *RASharedDB) Stop() {
	ra.ms.Stop()
	ra.done <- true
}

// Funciones que no tienen que ver con el algoritmo

func intMax(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func exclude(miAccion string, suAccion string) bool {
	return !(miAccion == "leer" && suAccion == "leer")
}

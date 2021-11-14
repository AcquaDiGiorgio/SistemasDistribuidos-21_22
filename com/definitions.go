/*
* AUTOR: Rafael Tolosana Calasanz
* ASIGNATURA: 30221 Sistemas Distribuidos del Grado en Ingeniería Informática
*			Escuela de Ingeniería y Arquitectura - Universidad de Zaragoza
* FECHA: septiembre de 2021
* FICHERO: definitions.go
* DESCRIPCIÓN: contiene las definiciones de estructuras de datos necesarias para
*			la práctica 3
 */
package com

type TPInterval struct {
	A int
	B int
}

type Action struct {
	Accion int
	Args   []string
}

type Worker struct {
	Ip   string
	Host string
}

type Salida struct {
	Id       int
	Interval TPInterval
}

const POOL = 6

var Workers = [...]Worker{
	{"lab102-200.cps.unizar.es", "155.210.154.200:30000"},
	{"lab102-200.cps.unizar.es", "155.210.154.200:30001"},
	{"lab102-200.cps.unizar.es", "155.210.154.200:30002"},
	{"lab102-200.cps.unizar.es", "155.210.154.200:30003"},
	{"lab102-200.cps.unizar.es", "155.210.154.200:30004"},
	{"lab102-199.cps.unizar.es", "155.210.154.200:30000"}}

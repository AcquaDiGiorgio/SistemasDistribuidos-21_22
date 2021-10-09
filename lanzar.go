package main

import (
	"fmt"
	"log"
	"main/com"
	"os"
	"strconv"
)

const GO = "/usr/local/go/bin/go"
const PATH = "~/SSDD/"
const RSA = "/Users/jorge/.ssh/id_rsa"

func lanzar(ssh *com.SshClient, file string) {
	output, err := ssh.RunCommand(PATH + file)
	fmt.Println(output)
	if err != nil {
		log.Printf("SSH run command error %v", err)
		return
	}

}

func main() {
	args := os.Args[1:]

	if len(args) != 4 {
		fmt.Println("Número de parámetros incorrecto, ejecutar como:\n\tgo run lanzar.go user num_host_server num_host_client port")
		os.Exit(1)
	}

	hostServ := "lab102-" + args[1] + ".cps.unizar.es"
	ip := "155.210.154." + args[1]
	hostClie := "lab102-" + args[2] + ".cps.unizar.es"

	//fmt.Print("Introduzca la Contraseña: ")
	//pass, err := terminal.ReadPassword(0)

	sshServ, err := com.NewSshClient(
		args[0],
		hostServ,
		22,
		RSA,
		"Fsw5zw")
	if err != nil {
		log.Printf("SSH init error %v", err)
		return
	}

	sshClie, err := com.NewSshClient(
		args[0],
		hostClie,
		22,
		RSA,
		"Fsw5zw")

	if err != nil {
		log.Printf("SSH init error %v", err)

	} else {
		var ini int
		var fin int

		fmt.Print("Introduzca el pincipio del Intervalo: ")
		fmt.Scanln(&ini)
		fmt.Print("Introduzca el final del Intervalo: ")
		fmt.Scanln(&fin)

		argsServ := ip + " " + args[3]
		argsClie := ip + " " + args[3] + " " + strconv.Itoa(ini) + " " + strconv.Itoa(fin)

		fmt.Println(argsClie)

		go lanzar(sshServ, "server "+argsServ)
		lanzar(sshClie, "client "+argsClie)
	}
}

package main

import (
	"fmt"
	"log"
	"main/com"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/term"
)

const GO = "/usr/local/go/bin/go run "
const PATH = "~/SSDD/Trabajo1/"
const RSA = "/Users/jorge/.ssh/id_rsa"

func lanzar(ssh *com.SshClient, file string, salida bool) {
	output, err := ssh.RunCommand(PATH + file)
	if err != nil {
		log.Printf("SSH run command error %v", err)
		return
	}
	if salida {
		fmt.Println(output)
	}
}

func crearSSH(usuario string, host string, pass string) (ssh *com.SshClient) {
	ssh, err := com.NewSshClient(
		usuario,
		host,
		22,
		RSA,
		strings.TrimSpace(pass))
	if err != nil {
		log.Printf("SSH init error %v", err)
		os.Exit(1)
	}
	return
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

	fmt.Print("Introduzca la Contraseña: ")
	pass, err := term.ReadPassword(int(syscall.Stdin))

	if err != nil {
		fmt.Fprintf(os.Stderr, "Password error: %s", err.Error())
		os.Exit(1)
	}

	sshServ := crearSSH(args[0], hostServ, string(pass))
	sshClie := crearSSH(args[0], hostClie, string(pass))

	var ini int
	var fin int

	fmt.Print("\nIntroduzca el pincipio del Intervalo: ")
	fmt.Scanln(&ini)
	fmt.Print("Introduzca el final del Intervalo: ")
	fmt.Scanln(&fin)

	argsServ := ip + " " + args[3]
	argsClie := ip + " " + args[3] + " " + strconv.Itoa(ini) + " " + strconv.Itoa(fin)

	go lanzar(sshServ, "server "+argsServ, false)
	time.Sleep(1 * time.Second)
	lanzar(sshClie, "client "+argsClie, true)

}

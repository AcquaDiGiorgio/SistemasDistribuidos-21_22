package main

import (
	"fmt"
	"log"
	"main/com"
	"os"
)

func main() {
	args := os.Args[1:]

	if len(args) != 2 {
		fmt.Println("Número de parámetros incorrecto, ejecutar como:\n\tgo run lanzar.go user password")
		os.Exit(1)
	}

	ssh, err := com.NewSshClient(
		args[0],
		"hendrix-ssh.cps.unizar.es",
		22,
		"/Users/jorge/.ssh/id_rsa",
		args[1])

	if err != nil {
		log.Printf("SSH init error %v", err)
	} else {
		output, err := ssh.RunCommand("ls")
		fmt.Println(output)
		if err != nil {
			log.Printf("SSH run command error %v", err)
		}
	}
}

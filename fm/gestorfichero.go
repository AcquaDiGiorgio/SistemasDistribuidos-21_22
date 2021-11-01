package fm // file manager

import (
	"bufio"
	"fmt"
	"os"
)

const PATH = "file.txt"

func checkError(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func LeerFichero() string {

	file, err := os.OpenFile("file.txt", os.O_APPEND|os.O_RDONLY, 0644)
	checkError(err)
	defer file.Close()

	reader := bufio.NewReader(file)

	var read string
	var readError error = nil
	var str string

	for readError == nil {
		read += str
		str, readError = reader.ReadString('\n')
	}

	return read
}

func EscribirFichero(datos string) {
	file, err := os.OpenFile("file.txt", os.O_APPEND|os.O_RDWR, 0644)
	checkError(err)
	defer file.Close()

	writter := bufio.NewWriter(file)
	_, err = writter.WriteString(datos)
	checkError(err)

	writter.Flush()
}

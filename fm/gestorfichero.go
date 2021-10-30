package fm // file manager

import (
	"fmt"
	"os"
	"time"
)

const PATH = "file.txt"

func checkError(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func LeerFichero() string {

	time.Sleep(500 * time.Millisecond)
	return "Leido jaja"
	/*
		var file, err = os.Open(PATH)
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
	*/
}

func EscribirFichero(datos string) {
	time.Sleep(1 * time.Second)
	/*
		file, err := os.Open(PATH)
		checkError(err)
		defer file.Close()

		writter := bufio.NewWriter(file)
		_, err = writter.WriteString(datos)
		checkError(err)

		writter.Flush()
	*/
}

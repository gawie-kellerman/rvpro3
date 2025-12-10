package main

import (
	"fmt"
	"io"
	"log"

	"github.com/angli232/serial"
)

func main() {
	var connection *serial.Port
	var err error
	var buf [1999]byte
	var readBytes int

	config := &serial.Config{
		BaudRate:    115200,
		DataBits:    8,
		Parity:      0,
		StopBits:    0,
		FlowControl: 0,
	}

	if connection, err = serial.Open("/dev/ttymxc2", config); err != nil {
		goto errorLabel
	}

	if err = connection.SetTimeout(100, 200); err != nil {
		goto errorLabel
	}

	for n := 1; n <= 400; n++ {
		readBytes, err = connection.Read(buf[:])
		if err != nil && err != io.EOF {
			goto errorLabel
		}

		fmt.Println(readBytes)
	}

	return

errorLabel:
	log.Fatal(err)
}

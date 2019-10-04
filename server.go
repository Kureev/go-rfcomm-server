package main

import (
	"log"
	"fmt"
	"bytes"
	"golang.org/x/sys/unix"
)

func main() {
	// Open a bluetooth stream socket using RFCOMM
	fd, err := unix.Socket(unix.AF_BLUETOOTH, unix.SOCK_STREAM, unix.BTPROTO_RFCOMM)
	if err != nil {
		log.Fatalf("Can't open a bluetooth socket: %v\n", err)
	}

	// Bind it to the local address
	err = unix.Bind(fd, &unix.SockaddrRFCOMM{Channel: 1, Addr: [6]uint8{0, 0, 0, 0, 0, 0}})
	if err != nil {
		log.Fatalf("Can't bind: %v\n", err)
	}
	
	// Change socket mode to "listening"
	err = unix.Listen(fd, 1)
	if err != nil {
		log.Fatalf("Can't start listening to a socket: %v\n", err)
	}

	// Prevent server from exiting by creating a goroutine that never closes
	finished := make(chan bool)
	go runServer(fd)
	<- finished
}

// 4Kb is a maximum msg length
const MAX_MSG_SIZE = 4 * 1024

func runServer(fd int) {
	// Start accepting messages
	nfd, _, err := unix.Accept(fd)
	if err != nil {
		log.Fatalf("Can't start accept connection: %v\n", err)
	}

	// Create an accumulator buffer
	acc := make([]byte, MAX_MSG_SIZE)
	for {
		// Temporary message buffer size
		buffer := make([]byte, 1024)
	
		// Read message to buffer
		_, err := unix.Read(nfd, buffer)
		if err != nil {
			log.Fatal(err)
		}

		// Add buffer to accumulator as we might not receive the full message yet
		acc = append(acc, buffer...)

		// If we received terminating bytes, show the message and clean the accumulator
		if bytes.Contains(buffer, []byte{13, 10}) {
			fmt.Println("Finished message: ", string(acc))
			acc = make([]byte, MAX_MSG_SIZE)
		}
	}
}

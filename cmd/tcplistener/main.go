package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"httpfromtcp/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		fmt.Printf("Error creating listener: %v\n", err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Println("Server listenting on :42069")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v\n", err)
			continue
		}

		fmt.Println("Connection accepted at port 42069")

		r, err := request.RequestFromReader(conn)
		if err != nil {
			log.Println("Error reading from connection: ", err)
			continue
		}
		fmt.Println("Request line:")
		fmt.Printf("- Method: %v\n- Target: %v\n- Version: %v\n", r.RequestLine.Method, r.RequestLine.RequestTarget, r.RequestLine.HttpVersion)
		fmt.Println("Headers:")
		for key, value := range r.Headers {
			fmt.Printf("- %v: %v\n", key, value)
		}

		fmt.Println("The connection has been terminated")
	}
}

package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func getLinesChannel(conn net.Conn) <-chan string {
	ch := make(chan string)
	go func() {
		defer conn.Close()
		defer close(ch)

		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			ch <- scanner.Text()
		}

		if err := scanner.Err(); err != nil {
			log.Printf("Error reading from connection: %v", err)
		}
	}()
	return ch
}

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
		ch := getLinesChannel(conn)

		for elem := range ch {
			fmt.Println("read:", elem)
		}

		fmt.Println("The connection has been terminated")
	}
}

package main

import (
	"bufio"
	"log"
	"net"
	"os"
	"fmt"
)

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", ":42069")
	if err != nil {
		log.Println("Error resolving address: ", err)
		os.Exit(1)
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Println("Error preparing UDP connection: ", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Eror reading line from connection: ", err)
		}
		_, err = conn.Write([]byte(line))
		if err != nil {
			log.Println("Error writing to connection: ", err)
		}
	}
}

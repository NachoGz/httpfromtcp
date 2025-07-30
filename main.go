package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

 func main() {
	 file, err := os.Open("./messages.txt")
	 if err != nil {
	 	fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	 }

	var contents [8]byte
	bytesRead, err := file.Read(contents[:])

	var currentLine string
	currentLine += string(contents[:bytesRead])

	for bytesRead != 0 {
		parts := strings.Split(currentLine, "\n")
		if len(parts) > 1 {
			for idx, part := range parts {
				if idx != len(parts) - 1 {
					fmt.Printf("read: %s\n", part)
					currentLine = ""
				} else {
					currentLine += part
				}
			}
		}
		if err == io.EOF {
			if len(currentLine) > 0 {
				fmt.Printf("read: %s\n", currentLine)
			}
			os.Exit(0)
		} else if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			os.Exit(1)
		}
		bytesRead, err = file.Read(contents[:])
		currentLine += string(contents[:bytesRead])
	}
 }

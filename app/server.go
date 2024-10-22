package main

import (
	"fmt"
	"io"
	"net"
	"os"
)


func main() {
	ln,err := net.Listen("tcp",":6379")
	if err!=nil{
		fmt.Printf("Error: %v",err)
		os.Exit(1)
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil{
			fmt.Printf("Error: %v",err)
			os.Exit(1)

		}
		go handleConnection(conn,[]byte("+PONG\r\n"))
	}
}

func handleConnection(conn net.Conn, message []byte){
	defer conn.Close()

	for {
		buffer := make([]byte,1024)
		_,err := conn.Read(buffer)

		if err == io.EOF{
			return
		}
		_, err = conn.Write(message)
		if err!=nil{
			fmt.Printf("Error: %v",err)
		}
	}
}

// select response string according to client message



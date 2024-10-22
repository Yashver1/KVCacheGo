package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

const (
	simpleStrings byte = '+'
	errorString byte = '-'
	bulkStrings byte = '$'
	arrays byte = '*'
	integers byte = ':'
)


var Commands  = map[string]struct{}{
	"ECHO":{},
	"PING":{},
}

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
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn){
	defer conn.Close()

	for {

		buffer := make([]byte,1024)
		lengthOfData, err := conn.Read(buffer)
		if err == io.EOF{
			return
		}
		reader := bytes.NewReader(buffer[:lengthOfData])
		result, err := parseRESP(reader)
		if err != nil{
			fmt.Print("Error, unable to parse string")
			return
		}

		var message []byte
		if array, ok := result.([]interface{}); ok{
			command, ok := array[0].([]byte)
			if !ok{
				fmt.Printf("Error command is not a bytestring")
			}

			if strings.ToUpper(string(command)) == "ECHO"{
				message = array[1].([]byte)
				message = []byte("+" + string(message) + "\r\n")

			} else if strings.ToUpper(string(command)) == "PING"{
				message = []byte("+PONG\r\n")
			}

		} else if _, ok := result.([]byte); ok{

			message = []byte("+OK\r\n")
		}

		_, err = conn.Write(message)
		if err!=nil{
			fmt.Printf("Error: %v",err)
		}
	}
}

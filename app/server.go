package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
)

const (
	simpleStrings byte = '+'
	errorString byte = '-'
	bulkStrings byte = '$'
	arrays byte = '*'
	integers byte = ':'
)

type KVStore struct {
	*sync.RWMutex
	store map[string][]byte
}

var kvstore = KVStore{
	RWMutex: &sync.RWMutex{},
	store: make(map[string][]byte),
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

		message, err := selectReply(reader)
		if err!=nil{
			fmt.Printf("Error from parsing Message")
			return 
		}
		_, err = conn.Write(message)
		if err!=nil{
			fmt.Printf("Error: %v",err)
		}
	}
}


func selectReply(reader *bytes.Reader) ([]byte,error){
	clientMessage, err := parseRESP(reader)
	if err!=nil{
		return nil, err
	}

	var message []byte

	// For commands
	if messageArray , ok := clientMessage.([]interface{}); ok{
		fmt.Printf("Recieved Message: %s",messageArray)
		// get command

		command, ok := messageArray[0].([]byte)

		if !ok{
			return nil, fmt.Errorf("expected []byte for command")
		}

		switch strings.ToUpper(string(command)){
		case "ECHO":

			messageContent, ok := messageArray[1].([]byte)
			if !ok {
				return nil, fmt.Errorf("expected []byte for message content")
			}
			message = []byte("+" + string(messageContent) + "\r\n")

		case "PING":
			message = []byte("+PONG\r\n")

		case "SET":

			key, ok := messageArray[1].([]byte)
			if !ok {
				return nil, fmt.Errorf("expected []byte for key")
			}
			value, ok := messageArray[2].([]byte)
			if !ok {
				return nil, fmt.Errorf("expected []byte for value")
			}

			kvstore.Lock()
			kvstore.store[string(key)] = value
			kvstore.Unlock()

			message = []byte("+OK\r\n")

		case "GET":

			key , ok := messageArray[1].([]byte)
			if !ok{
				return nil, fmt.Errorf("expected []byte for key")
			}

			kvstore.RLock()
			value := kvstore.store[string(key)]
			kvstore.RUnlock()
			valueLength := len(string(value))
			message = []byte(fmt.Sprintf("$%v\r\n%s\r\n",valueLength,string(value)))


		default:
			return nil, fmt.Errorf("error:%v",err)
		}

	  //For strings, ints etc
	}  else if messageBytes, ok := clientMessage.([]byte);  ok {
		fmt.Printf("Recieved Message: %v",string(messageBytes))
		message = []byte("+OK\r\n")

	}

	return message, nil
}

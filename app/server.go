package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"
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
	store map[string]*Entry
}

type Entry struct {
	entry []byte
	creationTime time.Time
	expiryTime time.Time
	
}


var kvstore = KVStore{
	RWMutex: &sync.RWMutex{},
	store: make(map[string]*Entry),
}

func main() {
	ln,err := net.Listen("tcp",":6379")
	if err!=nil{
		fmt.Printf("Error: %v",err)
		os.Exit(1)
	}
	defer ln.Close()
	
	ticker := time.NewTicker(5*time.Minute)
	defer ticker.Stop()

	//Cleaner Loop
	go func(){
		for range ticker.C{
			removeExpired()
		}
	}()

	//Handler Loop
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
			fmt.Printf("Error from parsing Message: %v",err)
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

	// Check if the client message is an array of interfaces
	if tempArray, ok := clientMessage.([]interface{}); ok {
		fmt.Printf("Received Message: %s\n", tempArray)

		var messageArray [][]byte

		// Convert interface array to array of bytes
		for _, msg := range tempArray {
			checkedMsg, ok := msg.([]byte)
			if !ok {
				return nil, fmt.Errorf("expected []byte for elements in messageArray")
			}
			messageArray = append(messageArray, checkedMsg)
		}

		// Get command
		command := messageArray[0]
		switch strings.ToUpper(string(command)){

		case "ECHO":
			message = handleECHO(messageArray)
		case "PING":
			message = handlePING()
		case "SET":
			message, err = handleSET(messageArray) 
			if err!=nil{
				return nil,err
			}
		case "GET":
			message = handleGET(messageArray)
		default:
			return nil, fmt.Errorf("error:%v",err)
		}

	  //For strings, ints etc
	}  else if messageBytes, ok := clientMessage.([]byte);  ok {
		fmt.Printf("Recieved Message: %v\n",string(messageBytes))
		message = []byte("+OK\r\n")

	}

	return message, nil
}


func removeExpired(){
	kvstore.RLock()
	defer kvstore.RUnlock()
	for key, entry := range kvstore.store{
		if entry.expiryTime.Before(time.Now()){
			delete(kvstore.store, key)
		}
	}
}
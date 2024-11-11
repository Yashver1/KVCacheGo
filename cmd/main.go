package main

import (
	"bytes"
	//"errors"
	//"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"
	//"path/filepath"
	parser "github.com/Yashver1/KVCacheGo/pkg/parser"
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

type Config struct {
	dir string
	dbFileName string
}

var config Config

// func init(){
// 	var dir string
// 	var dbFileName string
// 	flag.StringVar(&dir,"dir","","The path to the RDB file store")
// 	flag.StringVar(&dbFileName,"dbfilename","","The name of the RDB" )
// 	flag.Parse()

// 	if _, err := os.Stat(filepath.Join(dir,dbFileName)); errors.Is(err, os.ErrNotExist){
// 		dir = ""
// 	}

// 	config = Config{
// 		dir: dir,
// 		dbFileName: dbFileName,
// 	}
// }

func main() {


	fmt.Printf("server config: %v",config)
	kvstore := KVStore{
		RWMutex: &sync.RWMutex{},
		store: make(map[string]*Entry),
	}

	ln,err := net.Listen("tcp",":6379")
	if err!=nil{
		fmt.Printf("Error: %v",err)
		os.Exit(1)
	}
	defer ln.Close()

	ticker := time.NewTicker(time.Minute*5)
	defer ticker.Stop()

	//expiry clear loop
	go func(){
		for range ticker.C{
			kvstore.removeExpired()
		}
	}()



	//Main server handle loop
	for {
		conn, err := ln.Accept()
		if err != nil{
			fmt.Printf("Error: %v",err)
			os.Exit(1)

		}
		go handleConnection(conn,&kvstore)
	}
}

func handleConnection(conn net.Conn, kvstore *KVStore){
	defer conn.Close()

	for {
		buffer := make([]byte,1024)
		lengthOfData, err := conn.Read(buffer)
		if err == io.EOF{
			return
		}
		reader := bytes.NewReader(buffer[:lengthOfData])

		message, err := selectReply(reader, kvstore)
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


func selectReply(reader *bytes.Reader, kvstore *KVStore) ([]byte,error){

	clientMessage, err := parser.ParseRESP(reader)
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

		case "CONFIG":
			if string(messageArray[1]) == "GET"{
				message = handleCONFIGGET(messageArray)
			}

		case "ECHO":
			message = handleECHO(messageArray)
		case "PING":
			message = handlePING()
		case "SET":
			message, err = kvstore.handleSET(messageArray) 
			if err!=nil{
				return nil,err
			}
		case "GET":
			message = kvstore.handleGET(messageArray)
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

func (kvstore *KVStore) removeExpired(){
	kvstore.Lock()
	defer kvstore.Unlock()
	for key, entry := range kvstore.store {
		if (entry.expiryTime.Before(time.Now())){
			delete(kvstore.store, key)
		}
	}
}
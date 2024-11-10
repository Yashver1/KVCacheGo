package main

import (
	"bytes"
	"fmt"
	"strconv"
	"time"
)

func handlePING() []byte {
	return []byte("+PONG\r\n")
}

func handleECHO(messages [][]byte) []byte {
	var result []byte
	echoValue:= messages[1]
	result = append(result, '+')
	result = append(result, echoValue...)
	result = append(result, []byte("\r\n")...)

	return result

}


func handleCONFIGGET(messages [][]byte) []byte{
	var result []byte
	requestedFlag := string(messages[2])


	switch requestedFlag{
	case "dir":
		flagValue := config.dir
		flagValueLength := strconv.Itoa(len(flagValue))
		result = []byte(fmt.Sprintf("*2\r\n$3\r\ndir\r\n$%s\r\n%s\r\n",flagValueLength,flagValue))

	case "dbfilename":
		flagValue := config.dbFileName
		flagValueLength := strconv.Itoa(len(flagValue))
		result = []byte(fmt.Sprintf("*2\r\n$10\r\ndbfilename\r\n$%s\r\n%s\r\n",flagValueLength,flagValue))
	}

	return result
}


func (kvstore *KVStore) handleGET(messages [][]byte) []byte{
	var results []byte
	key := messages[1]
	expiredFlag := false

	kvstore.RLock()
	value := kvstore.store[string(key)] 
	if value.expiryTime.Before(time.Now()){
		expiredFlag = true
		delete(kvstore.store,string(key))
	}
	kvstore.RUnlock()

	if expiredFlag{

		return []byte("$-1\r\n")
	}



	valueLength := strconv.Itoa(len(string(value.entry)))

	results = append(results, '$')
	results = append(results, []byte(valueLength)...)
	results = append(results, []byte("\r\n")...)
	results = append(results, []byte(value.entry)...)
	results = append(results, []byte("\r\n")...)


	return results
}

func (kvstore *KVStore) handleSET(messages [][]byte) ([]byte, error){
	key := messages[1]
	value := messages[2]
	var liveDuration int
	currentTime := time.Now()
	expiryTime := currentTime.Add(time.Duration(int(time.Hour)*10000))

	if (len(messages) > 4 && bytes.Equal(messages[3],[]byte("px"))){
		var err error
		liveDuration, err = strconv.Atoi(string(messages[4]))
		if err!=nil{
			return nil,err
		}
		expiryTime = currentTime.Add(time.Duration(int(time.Millisecond)*liveDuration))
	}

	
	kvstore.Lock()
	kvstore.store[string(key)] = &Entry{
		entry: value,
		creationTime: currentTime,
		expiryTime: expiryTime,
	}
	kvstore.Unlock()

	return []byte("+OK\r\n"),nil

}
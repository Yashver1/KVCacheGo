package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"
)


type LengthEncodedValue struct {
	isInt bool 
	value []byte
}


func readRdbFile(reader *bytes.Reader) (map[string]string, error){


	fileStartIndicator := make([]byte,5)
	if _,err := reader.Read(fileStartIndicator); err!=nil{
		return nil, err
	}

	if string(fileStartIndicator) != "REDIS"{
		return nil,fmt.Errorf("file is not a real RDB file")
	}

	redisVersionNumber := make([]byte,4)
	if _,err := reader.Read(redisVersionNumber); err!=nil{
		return nil, err
	}

	redisVersionConverted, err := strconv.Atoi(string(redisVersionNumber))
	if err!=nil{
		return nil, err
	}

	if redisVersionConverted > 12 ||  redisVersionConverted < 0{
		return nil, fmt.Errorf("redis version is not 12")
	}
	result := make(map[string]string)	
	result["rdbFileStart"] = string(fileStartIndicator) + string(redisVersionNumber)


	//check checksum before hand

	for {

		opCode , err := reader.ReadByte()
		if err!=nil{
			return nil, err
		}

		switch opCode {
		case 0xFA:
			key , err := readLengthEncodedString(reader)
			if err!=nil{
				return nil, err
			}
			value, err := readLengthEncodedString(reader)
			if err!=nil{
				return nil, err
			}

			var parsedValue string

			if value.isInt {
				parsedValue = strconv.Itoa(int(binary.BigEndian.Uint32(value.value)))
			} else {
				parsedValue = string(value.value)
			}

			result[string(key.value)] = string(parsedValue)

		case 0xFB:
		case 0xFC:
		case 0xFD:
		case 0xFE:
		case 0xFF:
			return result , nil
		default:
			return nil, fmt.Errorf("invalid Op Code")
		}
	}

}

// func readRdbKeyValuePairs(reader *bytes.Reader)([]byte, error){
// 	return []byte("test"), nil
// }

func readLengthEncodedString(reader *bytes.Reader)(LengthEncodedValue, error){
	initByte, err := reader.ReadByte()
	if err!=nil{
		return LengthEncodedValue{}, err
	}

	bits := ( initByte >> 6 ) & 0x3

	switch bits{
	case 0x00:
		lengthInBits := (initByte & 0x3f)
		buffer := make([]byte,int(lengthInBits))
		if _,err := reader.Read(buffer); err!=nil{
			return LengthEncodedValue{}, err
		}
		return LengthEncodedValue{false,buffer}, nil


	case 0x01:
		firstHalf := (initByte & 0x3f)
		secondHalf, err := reader.ReadByte()
		if err!=nil{
			return LengthEncodedValue{}, err
		}

		length := binary.BigEndian.Uint16([]byte{firstHalf, secondHalf})
		buffer := make([]byte, int(length))
		if _,err :=reader.Read(buffer); err!=nil{
			return LengthEncodedValue{}, err
		}
		return LengthEncodedValue{false, buffer},nil

	case 0x02:
		buffer := make([]byte,4)
		if _, err := reader.Read(buffer); err!=nil{
			return LengthEncodedValue{}, err
		}

		length := binary.BigEndian.Uint32(buffer)
		wordBuffer := make([]byte,length)
		if _, err := reader.Read(wordBuffer); err!=nil{
			return LengthEncodedValue{}, err
		}
		return LengthEncodedValue{false, wordBuffer},nil

	case 0x03:
		remainingSixBits := (initByte & 0x3f)

		switch int(remainingSixBits){
		case 0:
			buffer := make([]byte,1)
			if _, err := reader.Read(buffer); err!=nil{
				return LengthEncodedValue{}, err
			}
			buffer2 := make([]byte,3)
			buffer2 = append(buffer2, buffer...)
			return LengthEncodedValue{true,buffer2}, nil
		case 1:
			buffer := make([]byte,2)
			if _, err := reader.Read(buffer); err!=nil{
				return LengthEncodedValue{}, err
			}
			buffer2 := make([]byte,2)
			buffer2 = append(buffer2, buffer...)
			return LengthEncodedValue{true,buffer2}, nil
		case 2:
			buffer := make([]byte,4)
			if _, err := reader.Read(buffer); err !=nil {
				return LengthEncodedValue{},err
			}
			return LengthEncodedValue{true,buffer},nil
		//case 3: ignore for now as using LZF encoding
		default:
			return LengthEncodedValue{}, fmt.Errorf("invalid Special Format after 0x11")
			
		}

	default:
		return LengthEncodedValue{}, fmt.Errorf("invalid inital length encoding bits")

	}
	
}

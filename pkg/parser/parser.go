package parser

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
)

const (
	simpleStrings byte = '+'
	errorString byte = '-'
	bulkStrings byte = '$'
	arrays byte = '*'
	integers byte = ':'
)

func readTilEndOfType(reader *bytes.Reader, delimiter byte) ([]byte, error){
	var data []byte
	for {
		curr, err := reader.ReadByte()
		if err!=nil{
			return nil, err
		}
		if curr == delimiter{
			break

		}
		data = append(data, curr)
	}

	//clear \r \ n
	clearEndOfByte(reader)
	return data, nil
}


func clearEndOfByte(reader *bytes.Reader) error{
	data, err := reader.ReadByte()
	if err!=nil{
		return err
	}
	if data != '\r' {
		return fmt.Errorf("error: /r not found")
	}

	data, err = reader.ReadByte()
	if err!=nil{
		return err
	}
	if data != '\n' {
		return fmt.Errorf("error: /n not found")
	}

	return nil
}

// returns either slice of bytes OR returns an array of slice of bytes
func ParseRESP(reader *bytes.Reader) (interface{}, error)  {
	data,err := reader.ReadByte()
	if err!=nil{
		return nil, err
	}
	switch data{


	case simpleStrings:
		return parseSimpleStrings(reader)

	case bulkStrings:
		length, err := readTilEndOfType(reader,'\r')
		if err!=nil{
			return nil, err
		}
		bulkLength, err := strconv.Atoi(string(length))
		if err!=nil{
			return nil,err
		}
		return parseBulkStrings(reader,bulkLength)

	case arrays:
		length , err := reader.ReadByte()
		if err!=nil{
			return nil, err
		}
		arrayLength, err := strconv.Atoi(string(length))
		if err!=nil{
			return nil, err
		}
		clearEndOfByte(reader)
		return parseArray(reader, arrayLength)
//	case integers:
		// handle integers
//	case errorString:
		// handle errors
	default:
		return nil,fmt.Errorf("error: dataType not valid,%s",string(data))
	}

}


func parseArray(reader *bytes.Reader, length int)([]interface{}, error){
	var result []interface{}
	for i := 0 ; i < length ; i++{
		curr, err := ParseRESP(reader)
		if err!=nil{
			return nil, err
		}
		result = append(result, curr)
	}
	return result, nil
}

func parseSimpleStrings(reader *bytes.Reader) ([]byte, error){
	return readTilEndOfType(reader, '\r')
}

func parseBulkStrings(reader *bytes.Reader, length int)([]byte, error){
	bulkStringBuffer := make([]byte, length)
    _, err := io.ReadFull(reader,bulkStringBuffer)
	if err!=nil{
		return nil, err
	}

	clearEndOfByte(reader)
	return bulkStringBuffer, nil
}

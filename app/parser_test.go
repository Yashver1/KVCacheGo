package main

import (
	"bytes"
	"reflect"
	"testing"
)

func TestReadTilEndOfType(t *testing.T) {
	tests := []struct {
		input    string
		delimiter byte
		expected []byte
		expectErr bool
	}{
		{"hello\r\n", '\r', []byte("hello"), false},
		{"test123\r\n", '\r', []byte("test123"), false},
		{"no delimiter", '\r', nil, true},
	}

	for _, test := range tests {
		reader := bytes.NewReader([]byte(test.input))
		result, err := readTilEndOfType(reader, test.delimiter)

		if test.expectErr && err == nil {
			t.Errorf("Expected error, got nil")
		}

		if !test.expectErr && err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("Expected %v, got %v", test.expected, result)
		}
	}
}

func TestClearEndOfByte(t *testing.T) {
	tests := []struct {
		input     string
		expectErr bool
	}{
		{"\r\n", false},
		{"\r", true},
		{"\n", true},
		{"", true},
	}

	for _, test := range tests {
		reader := bytes.NewReader([]byte(test.input))
		err := clearEndOfByte(reader)

		if test.expectErr && err == nil {
			t.Errorf("Expected error, got nil")
		}

		if !test.expectErr && err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}
}

func TestParseRESP(t *testing.T) {
	// Define simpleStrings, bulkStrings, and arrays constants
	tests := []struct {
		input    string
		expected interface{}
		expectErr bool
	}{
		{"+hello\r\n", []byte("hello"), false},
		{"$5\r\nhello\r\n", []byte("hello"), false},
		{"*2\r\n+hello\r\n+world\r\n", []interface{}{[]byte("hello"), []byte("world")}, false},
		{"invalid", nil, true},
	}

	for _, test := range tests {
		reader := bytes.NewReader([]byte(test.input))
		result, err := parseRESP(reader)

		if test.expectErr && err == nil {
			t.Errorf("Expected error, got nil")
		}

		if !test.expectErr && err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("Expected %v, got %v", test.expected, result)
		}
	}
}

func TestParseArray(t *testing.T) {
	input := "*2\r\n+hello\r\n+world\r\n"
	expected := []interface{}{[]byte("hello"), []byte("world")}

	reader := bytes.NewReader([]byte(input))
	// Skip the '*' and '2' characters
	reader.ReadByte()
	reader.ReadByte()
	clearEndOfByte(reader)

	result, err := parseArray(reader, 2)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestParseSimpleStrings(t *testing.T) {
	tests := []struct {
		input    string
		expected []byte
		expectErr bool
	}{
		{"hello\r\n", []byte("hello"), false},
		{"test123\r\n", []byte("test123"), false},
		{"no delimiter", nil, true},
	}

	for _, test := range tests {
		reader := bytes.NewReader([]byte(test.input))
		result, err := parseSimpleStrings(reader)

		if test.expectErr && err == nil {
			t.Errorf("Expected error, got nil")
		}

		if !test.expectErr && err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("Expected %v, got %v", test.expected, result)
		}
	}
}

func TestParseBulkStrings(t *testing.T) {
	tests := []struct {
		input    string
		length   int
		expected []byte
		expectErr bool
	}{
		{"hello\r\n", 5, []byte("hello"), false},
		{"test123\r\n", 7, []byte("test123"), false},
		{"short", 10, nil, true},
	}

	for _, test := range tests {
		reader := bytes.NewReader([]byte(test.input))
		result, err := parseBulkStrings(reader, test.length)

		if test.expectErr && err == nil {
			t.Errorf("Expected error, got nil")
		}

		if !test.expectErr && err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("Expected %v, got %v", test.expected, result)
		}
	}
}
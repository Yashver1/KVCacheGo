package main

import (
	"bytes"
	"reflect"
	"testing"
)

func TestReadRdbFile(t *testing.T) {
	tests := []struct {
		input    string
		expected map[string]string
		expectErr bool
	}{
		{"REDIS0000\r\n", map[string]string{"rdbFileStart": "REDIS0000"}, false},
		{"NOTREDIS0000\r\n", nil, true},
		{"REDIS0001\r\n", map[string]string{"rdbFileStart": "REDIS0001"}, false},
	}

	for _, test := range tests {
		reader := bytes.NewReader([]byte(test.input))
		result, err := readRdbFile(reader)

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

// TestParseRESP checks that the parseRESP function works correctly by
// providing a number of test cases.
//
// The test cases cover the following scenarios:
// - A simple string
// - A bulk string
// - An array of bulk strings
// - An invalid input
// - A new test case added to test multi bulk strings
func TestParseRESP(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
		expectErr bool
	}{
		{"+hello\r\n", []byte("hello"), false},
		{"$5\r\nhello\r\n", []byte("hello"), false},
		{"*2\r\n+hello\r\n+world\r\n", []interface{}{[]byte("hello"), []byte("world")}, false},
		{"invalid", nil, true},
		// New test case
		{"*3\r\n$3\r\nSET\r\n$10\r\nstrawberry\r\n$9\r\nraspberry\r\n", 
			[]interface{}{[]byte("SET"), []byte("strawberry"), []byte("raspberry")}, 
			false},
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
		{"strawberry\r\n", 10, []byte("strawberry"), false},
		{"SET\r\n", 3, []byte("SET"), false},
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
			t.Errorf("Expected %s, got %s", string(test.expected), string(result))
		}
	}
}

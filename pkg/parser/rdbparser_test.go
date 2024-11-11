package parser

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"
	"testing"
	"os"
	"io"
	
)

func TestReadRdbFileFromActualFile(t *testing.T) {
	// Open the file
	file, err := os.Open("../../cmd/dump.rdb")
	if err !=nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	// Read the file content into a byte slice
	fileContent, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	// Create a bytes.Reader from the file content
	reader := bytes.NewReader(fileContent)

	// Call the readRdbFile function
	result, err := readRdbFile(reader)
	if err != nil {
		t.Fatalf("Failed to read RDB file: %v", err)
	}

	// Print the result
	t.Logf("RDB File Result: %v", result)

	// Add your assertions here to validate the result
	// Example:
	if len(result) == 0 {
		t.Errorf("Expected non-empty result, got: %v", result)
	}
}

// TestReadLengthEncodedString tests the readLengthEncodedString function
// according to the length encoding rules described.
func TestReadLengthEncodedString(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  string
		isInt bool
	}{
		{
			name:  "00 - 6 bits length",
			input: []byte{0x05, 'h', 'e', 'l', 'l', 'o'}, // 00000101
			want:  "hello",
			isInt: false,
		},
		{
			name:  "01 - 14 bits length",
			input: []byte{0x40, 0x05, 'h', 'e', 'l', 'l', 'o'}, // 0100000 00000101
			want:  "hello",
			isInt: false,
		},
		{
			name:  "10 - 4 bytes length",
			input: []byte{0x80, 0x00, 0x00, 0x00, 0x05, 'h', 'e', 'l', 'l', 'o'}, // 10000000 00000000 00000000 00000000 00000101
			want:  "hello",
			isInt: false,
		},
		{
			name:  "11 - special format - 8 bit integer",
			input: []byte{0xc0, 0x05}, // 11000000 00000101
			want:  "5",
			isInt: true,
		},
		{
			name:  "11 - special format - 16 bit integer",
			input: []byte{0xc0 | 0x01, 0x00, 0x05}, // 11000001 00000000 00000101
			want:  "5",
			isInt: true,
		},
		{
			name:  "11 - special format - 32 bit integer",
			input: []byte{0xc0 | 0x02, 0x00, 0x00, 0x00, 0x05}, // 11000010 00000000 00000000 00000000 00000101
			want:  "5",
			isInt: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Printf("test")
			reader := bytes.NewReader(tt.input)
			got, err := readLengthEncodedString(reader)
			if err != nil {
				t.Errorf("readLengthEncodedString() error = %v", err)
				return
			}
			var value string
			if got.isInt {
				value = strconv.Itoa(int(binary.BigEndian.Uint32(got.value)))
			} else {
				value = string(got.value)
			}
			if value != tt.want {
				t.Errorf("readLengthEncodedString() = %v, want %v", value, tt.want)
			}
		})
	}
}
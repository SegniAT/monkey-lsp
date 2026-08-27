package rpc_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/SegniAT/monkey-lsp/rpc"
)

func TestReadMessage(t *testing.T) {

	tests := map[string]struct {
		input         io.Reader
		expected      []byte
		expectedError error
	}{
		"Valid String": {
			input:    strings.NewReader("Content-Length: 5\r\n\r\nhello"),
			expected: []byte("hello"),
		},
		"Zero-length body": {
			input:    strings.NewReader("Content-Length: 0\r\n\r\n"),
			expected: []byte(""),
		},
		"Missing separator": {
			input:         strings.NewReader("Content-Length: 5\r\nhello"),
			expected:      nil,
			expectedError: rpc.ErrSeparatorNotFound,
		},
		"Missing header Content-Length": {
			input:         strings.NewReader("Some-Other-Header: 777\r\n\r\nhello"),
			expected:      nil,
			expectedError: rpc.ErrMissingHeaderContentLength,
		},
		"Invalid header Content-Length": {
			input:         strings.NewReader("Content-Length: abc\r\n\r\nhello"),
			expected:      nil,
			expectedError: rpc.ErrInvalidContentLength,
		},
		"Incomplete body": {
			input:         strings.NewReader("Content-Length: 10\r\n\r\nhello"),
			expected:      nil,
			expectedError: rpc.ErrIncompleteMessage,
		},
		"Partial reads (OneByteReader)": {
			input:         iotest.OneByteReader(strings.NewReader("Content-Length: 17\r\n\r\n{\"method\":\"test\"}")),
			expected:      []byte("{\"method\":\"test\"}"),
			expectedError: nil,
		},
		"Network error (TimeoutReader)": {
			// Wrapped the string reader in a OneByteReader first because we use bufio.Reader under the hood, which defaults to reading data in 4KB chunks.
			input:         iotest.TimeoutReader(iotest.OneByteReader(strings.NewReader("Content-Length: 5\r\n\r\nhello"))),
			expected:      nil,
			expectedError: iotest.ErrTimeout,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			reader := rpc.NewReader(test.input)
			got, err := reader.ReadMessage()

			if test.expectedError != nil {
				if err == nil {
					t.Fatalf("expected error '%s' but got none, output: %s", test.expectedError.Error(), got)
				}
				if !errors.Is(err, test.expectedError) {
					t.Fatalf("expected error: %s\ngot error: %s", test.expectedError.Error(), err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !bytes.Equal(got, test.expected) {
				t.Errorf("expected: %s\ngot: %s", string(test.expected), got)
			}
		})
	}
}

func TestReadMultipleMessages(t *testing.T) {
	input := "Content-Length: 5\r\n\r\nhelloContent-Length: 5\r\n\r\nworld"
	reader := rpc.NewReader(strings.NewReader(input))

	got1, err := reader.ReadMessage()
	if err != nil {
		t.Fatalf("unexpected error on first read: %v", err)
	}
	if !bytes.Equal(got1, []byte("hello")) {
		t.Errorf("expected 'hello', got %s", got1)
	}

	got2, err := reader.ReadMessage()
	if err != nil {
		t.Fatalf("unexpected error on second read: %v", err)
	}
	if !bytes.Equal(got2, []byte("world")) {
		t.Errorf("expected 'world', got %s", got2)
	}

	_, err = reader.ReadMessage()
	if !errors.Is(err, rpc.ErrSeparatorNotFound) {
		t.Fatalf("expected stream to be empty/ErrSeparatorNotFound, got: %v", err)
	}
}

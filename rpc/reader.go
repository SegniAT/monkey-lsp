package rpc

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Reader struct {
	rd *bufio.Reader
}

func NewReader(reader io.Reader) *Reader {
	return &Reader{
		rd: bufio.NewReader(reader),
	}
}

func (r *Reader) ReadMessage() ([]byte, error) {
	contentLength, err := r.readHeaders()
	if err != nil {
		return nil, err
	}

	content := make([]byte, contentLength)

	_, err = io.ReadFull(r.rd, content)
	if err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, ErrIncompleteMessage
		}

		return nil, err
	}

	return content, nil
}

func (r *Reader) readHeaders() (int, error) {
	const prefix = "Content-Length: "
	var contentLength int

	found := false
	for {
		line, err := r.rd.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				return 0, ErrSeparatorNotFound // If we hit EOF before finding the \r\n separator
			}
			return 0, err
		}

		if line == "\r\n" {
			if !found {
				return 0, ErrMissingHeaderContentLength
			}
			return contentLength, nil
		}

		if !found && strings.HasPrefix(line, prefix) {
			value := strings.TrimSuffix(line, "\r\n")

			contentLength, err = strconv.Atoi(value[len(prefix):])
			if err != nil {
				return 0, fmt.Errorf("%w: %w", ErrInvalidContentLength, err)
			}

			found = true
		}
	}
}

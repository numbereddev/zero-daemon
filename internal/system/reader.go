package system

import (
	"bufio"
	"bytes"
	"errors"
	"io"
)

type ByteOrString interface {
	~string | ~[]byte
}

func ReadLines[T ByteOrString](r io.Reader, logFunc func(T)) error {
	reader := bufio.NewReader(r)
	if logFunc == nil {
		return errors.New("log func is required")
	}

	for {
		line, err := reader.ReadBytes('\n')

		if len(line) > 0 {
			line := bytes.TrimRight(line, "\r\n")
			logFunc(T(line))
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			// Streaming error
			return err
		}
	}
}

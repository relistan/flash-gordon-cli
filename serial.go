package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// A IOWrapper is a wrapper that gives an io.ReadWriter interface to
// stdin and stdout.
type IOWrapper struct {
	Input  *bufio.Reader
	Output io.Writer
}

func (s *IOWrapper) Read(p []byte) (n int, err error) {
	return s.Input.Read(p)
}

func (s *IOWrapper) ReadLine() (line []byte, isPrefix bool, err error) {
	return s.Input.ReadLine()
}

func (s *IOWrapper) Write(p []byte) (n int, err error) {
	return s.Output.Write(p)
}

// serialWriteLn writes a line of data to the serial port, terminating it with a newline
func serialWriteLn(port io.ReadWriter, data string, delay time.Duration) {
	var sent int

	for {
		n, err := port.Write([]byte(data[sent:]))
		sent += n
		if err != nil {
			log.Errorf("Error writing to port: %s", err)
			return
		}
		if sent >= len(data) {
			port.Write([]byte("\n"))
			break
		}
	}

	time.Sleep(delay)
}

// readAllData loops through anything pending in the serial buffer and prints
// it.
func readAllData(port *IOWrapper) {
	readAllDataToWriter(port, os.Stdout)
}

// readAllDataToIO loops through anything in the serial buffer and writes it
// out to a file.
func readAllDataToWriter(port *IOWrapper, out io.Writer) {
	buf := make([]byte, BytesPerLine)

	// Trigger at least _some_ output even if it's just sitting at the prompt
	serialWriteLn(port, "", 0)

	for {
		n, err := port.Read(buf)
		if err != nil && err != io.EOF {
			log.Errorf("Error reading from serial port: %s", err)
			break
		}

		if n <= 0 {
			break
		}

		if pos := bytes.Index(buf[:n], []byte("Enter Command")); pos != -1 {
			out.Write(buf[:pos])
			break
		}
		out.Write(buf[:n])
		time.Sleep(6 * time.Millisecond)
	}
}

func serialReadOutput(port *IOWrapper) (string, error) {
	data, _, _ := port.ReadLine()
	if strings.Contains(string(data), "ERROR") {
		return string(data), fmt.Errorf("%s", data)
	}

	return string(data), nil
}

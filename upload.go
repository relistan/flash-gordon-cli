package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	progressbar "github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"
)

// getFileSize is a little wrapper to get the size of a file in bytes
func getFileSize(file *os.File) int64 {
	info, err := file.Stat()
	if err != nil {
		log.Warnf("Error calling Stat() on file! %s", err)
		return -1
	}
	return info.Size()
}

// maybeEmitSegmentRecord handles more than 16 bits by emitting a new segment
// record when we have an address bigger than 16bits.
func maybeEmitSegmentRecord(outputPort io.Writer, addr, segment int) int {
	if addr > twoToThe16th-1 {
		addr -= twoToThe16th - 1
		segment++

		segmentBytes := make([]byte, 2)
		binary.BigEndian.PutUint16(segmentBytes, uint16(segment))
		serialWriteLn(outputPort,
			formatRecord(0x00, 0x02, segmentBytes), 10*time.Millisecond,
		)
	}
	return segment
}

// uploadCleanUp just does all the closing down of the setup
func uploadCleanUp(port *IOWrapper, addr int) error {
	// EOF Record
	log.Infof("Completed sending: %d bytes", addr)
	out, err := serialReadOutput(port)
	if err != nil {
		return err
	}
	println(out)
	readAllData(port)

	return nil
}

// uploadFile uploads a file by encoding it as an Intel hex file and streaming
// it over the serial port as it goes.
func uploadFile(port *IOWrapper, config *Config, file *os.File) error {
	var (
		addr       int = *config.BaseAddr
		segment    int = addr % twoToThe16th
		shouldStop bool
		buf        []byte = make([]byte, BytesPerLine)
	)

	readAllData(port)

	log.Info("Sending hex file...")

	// Begin the upload and tell it what kind of chip we're using
	switch {
	case *config.Flash32Pin:
		port.Write([]byte("uf"))
	case *config.EEPROM28Pin:
		port.Write([]byte("ue"))
	default:
		return errors.New("Unable to determine chip type!")
	}

	out, err := serialReadOutput(port)
	if err != nil {
		return err
	}
	println(out)

	// Starting address record
	serialWriteLn(port,
		formatRecord(0x0, RecTypeExtAddr, []byte{0x0, 0x0}),
		10*time.Millisecond,
	)
	if _, err := serialReadOutput(port); err != nil {
		return err
	}

	// Set up the progress bar
	size := getFileSize(file)
	progress := progressbar.DefaultBytes(
		int64(float64(size)*RatioBinToHex),
		"uploading",
	)

	outputPort := io.MultiWriter(port, progress)

	// Run until told otherwise
	for !shouldStop {
		readLen, err := io.ReadFull(file, buf)

		// This is expected when we hit the end of the file
		switch err {
		case nil:
			break
		case io.ErrUnexpectedEOF:
			shouldStop = true
		case io.EOF:
			shouldStop = true
		default:
			return fmt.Errorf("Error on read: %s", err)
		}

		serialWriteLn(outputPort,
			formatRecord(addr, RecTypeData, buf[:readLen]),
			10*time.Millisecond,
		)
		addr += readLen

		segment = maybeEmitSegmentRecord(outputPort, addr, segment)

		_, err = serialReadOutput(port)
		if err != nil {
			return err
		}
	}

	serialWriteLn(outputPort, ":00000001FF\n", 10*time.Millisecond)
	if err := uploadCleanUp(port, addr); err != nil {
		return err
	}
	file.Close()

	return nil
}

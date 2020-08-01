package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"time"

	log "github.com/sirupsen/logrus"
)

// uploadFile uploads a file by encoding it as an Intel hex file and streaming
// it over the serial port as it goes.
func uploadFile(port *IOWrapper, config *Config, file io.Reader) error {
	var (
		addr       int = *config.BaseAddr
		segment    int = addr % twoToThe16th
		shouldStop bool
		buf        []byte = make([]byte, BytesPerLine)
	)

	readAllData(port)

	log.Info("Sending hex file...")

	// Begin the upload
	port.Write([]byte("u"))
	serialReadOutput(port)

	// Starting address record
	serialWriteLn(port, formatRecord(0x0, 0x04, []byte{0x0, 0x0}), 10*time.Millisecond)
	if err := serialReadOutput(port); err != nil {
		return err
	}

	for !shouldStop {
		readLen, err := io.ReadFull(file, buf)
		// This is expected when we hit the end of the file
		if err == io.ErrUnexpectedEOF || err == io.EOF {
			shouldStop = true
		} else if err != nil {
			return fmt.Errorf("Error on read: %s", err)
		}

		serialWriteLn(port, formatRecord(addr, 0x0, buf[:readLen]), 10*time.Millisecond)
		addr += readLen

		// Handle more than 16 bits by emitting a new segment record when
		// we have an address bigger than 16bits.
		if addr > twoToThe16th-1 {
			addr -= twoToThe16th - 1
			segment++

			segmentBytes := make([]byte, 2)
			binary.BigEndian.PutUint16(segmentBytes, uint16(segment))
			serialWriteLn(port, formatRecord(0x00, 0x02, segmentBytes), 10*time.Millisecond)
		}
		if err := serialReadOutput(port); err != nil {
			return err
		}
	}

	// EOF Record
	serialWriteLn(port, ":00000001FF\n", 10*time.Millisecond)
	log.Infof("Completed sending: %d bytes", addr)
	if err := serialReadOutput(port); err != nil {
		log.Error(err)
		return err
	}
	if err := serialReadOutput(port); err != nil {
		return err
	}
	readAllData(port)

	return nil
}

// dumpFile dumps the contents of the Flash to the terminal in a human readable format,
// starting from 0 and running to the end of the Flash.
func dumpFlash(port *IOWrapper) {
	readAllData(port)

	log.Info("Dumping Flash contents")

	// Begin the upload
	port.Write([]byte("d"))
	readAllData(port)

	log.Info("Completed dump")
}

func eraseFlash(port *IOWrapper) {
	readAllData(port)

	log.Info("Performing chip erase")

	// Begin the upload
	port.Write([]byte("e"))
	readAllData(port)

	log.Info("Completed erase")
}

func eraseFlashSector(port *IOWrapper, config *Config) {
	readAllData(port)

	log.Infof("Performing sector erase for sector %d", *config.Sector)
	port.Write([]byte("s"))
	readAllData(port)

	log.Info("Completed sector erase")
}

func main() {
	config := parseConfig()
	file := configureFile(config)
	port := configureSerial(config)
	logConfig(config)

	switch config.Command {
	case "upload":
		err := uploadFile(port, config, file)
		if err != nil {
			log.Error(err.Error())
		}
	case "dump":
		dumpFlash(port)
	case "erase":
		eraseFlash(port)
	}
}

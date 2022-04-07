package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	// twoToThe16th is used in checking address lengths
	twoToThe16th = 65536
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

	// Tell it what kind of chip we're using
	if *config.Flash32Pin {
		port.Write([]byte("f"))
	} else if *config.EEPROM28Pin {
		port.Write([]byte("e"))
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

	for !shouldStop {
		readLen, err := io.ReadFull(file, buf)
		// This is expected when we hit the end of the file
		if err == io.ErrUnexpectedEOF || err == io.EOF {
			shouldStop = true
		} else if err != nil {
			return fmt.Errorf("Error on read: %s", err)
		}

		serialWriteLn(port,
			formatRecord(addr, RecTypeData, buf[:readLen]),
			10*time.Millisecond,
		)
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
		out, err := serialReadOutput(port)
		if err != nil {
			return err
		}
		println(out)
	}

	// EOF Record
	serialWriteLn(port, ":00000001FF\n", 10*time.Millisecond)
	log.Infof("Completed sending: %d bytes", addr)
	out, err = serialReadOutput(port)
	if err != nil {
		log.Error(err)
		return err
	}
	println(out)

	out, err = serialReadOutput(port)
	if err != nil {
		return err
	}
	println(out)
	readAllData(port)

	return nil
}

// dumpFlash dumps the contents of the Flash to the terminal in a human
// readable format, starting from `start` and running to the end of the Flash.
func dumpFlash(port *IOWrapper, config *Config) {
	readAllData(port)

	log.Infof(
		"Dumping %d bytes from Flash contents starting from %d",
		*config.Length, *config.BaseAddr,
	)

	outFile, err := os.OpenFile(
		*config.OutputFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644,
	)
	if err != nil {
		log.Fatalf("Unable to create output file! '%s'", err)
	}

	// Begin the dump - command is 'd' plus 32bit hex base and length
	addrStr := fmt.Sprintf("d%08X%08X", *config.BaseAddr, *config.Length)
	log.Infof("AddrStr: %s", addrStr)
	port.Write([]byte(addrStr))
	readAllDataToWriter(port, outFile)
	outFile.Close()

	log.Info("Completed dump")
}

func eraseFlash(port *IOWrapper) {
	readAllData(port)

	log.Info("Performing chip erase")

	// Begin the erase
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

	readAllData(port)

	switch config.Command {
	case "upload":
		err := uploadFile(port, config, file)
		if err != nil {
			log.Error(err.Error())
		}
	case "dump":
		dumpFlash(port, config)
	case "erase":
		eraseFlash(port)
	}
}

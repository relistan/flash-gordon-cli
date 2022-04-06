package main

import (
	"bufio"
	"os"

	log "github.com/sirupsen/logrus"
	"go.bug.st/serial"
	"gopkg.in/alecthomas/kingpin.v2"
)

type Config struct {
	SerialPort  *string
	BaudRate    *int
	UseSerial   *bool
	BaseAddr    *int
	Flash32Pin  *bool
	EEPROM28Pin *bool

	Command string

	// Upload
	InputFile *string

	// Dump
	OutputFile *string
	Length     *int

	// Erase
	Sector *int
}

// parseConfig parse the command line using Kingpin and returns a Config
// struct that has been populated.
func parseConfig() *Config {
	uploadCommand := kingpin.Command("upload", "Upload a file to Flash Gordon")
	dumpCommand := kingpin.Command("dump", "Dump the contents of the flash chip")
	eraseCommand := kingpin.Command("erase", "Erase the whole flash chip contents")

	config := &Config{
		SerialPort:  kingpin.Flag("serial-port", "The Serial port name/path to use").Default("/dev/cu.usbserial-FTDOMLSO").String(),
		BaudRate:    kingpin.Flag("baud-rate", "The baud rate of the serial port").Default("57600").Int(),
		UseSerial:   kingpin.Flag("use-serial", "Serial or stdout?").Default("true").Bool(),
		Flash32Pin:  kingpin.Flag("32pin-flash", "Is this a 32 pin flash chip?").Bool(),
		EEPROM28Pin: kingpin.Flag("28pin-eeprom", "Is this a 28 pin EEPROM?").Bool(),
		BaseAddr:    kingpin.Flag("base-addr", "Base Address to use as starting address").Default("0").Int(),

		// UPLOAD
		InputFile: uploadCommand.Arg("input-file", "The file to take input from").Required().String(),

		// DUMP
		OutputFile: dumpCommand.Arg("output-file", "The file to write to locally").Required().String(),
		Length:     dumpCommand.Flag("length", "The number of bytes to dump").Default("1024").Int(),

		// ERASE
		Sector: eraseCommand.Flag("sector", "The number of the sector to erase").Int(),
	}

	command := kingpin.Parse()

	if *config.EEPROM28Pin && *config.Flash32Pin {
		log.Fatal("Must choose either a flash chip or an EEPROM, not both")
	} else if !*config.EEPROM28Pin && !*config.Flash32Pin {
		log.Fatal("Must choose at least one of: 32pin-flash or 28pin-eeprom")
	}

	switch command {
	case uploadCommand.FullCommand():
		config.Command = "upload"
	case dumpCommand.FullCommand():
		config.Command = "dump"
	case eraseCommand.FullCommand():
		config.Command = "erase"
	}

	return config
}

// configureFile returns either stdin or the open file named in the config
func configureFile(config *Config) *os.File {
	// If we got a filename use it. Otherwise use stdin.
	if *config.InputFile == "" {
		return os.Stdin
	}

	file, err := os.Open(*config.InputFile)
	if err != nil {
		log.Fatalf("Unable to open %s: %s", *config.InputFile, err)
	}

	return file
}
func logConfig(config *Config) {
	log.Info("Flash Gordon starting up")
	log.Info("--------------------------------------------")

	if *config.UseSerial {
		log.Infof("Serial Port: %s", *config.SerialPort)
		log.Infof("Baud Rate:   %d", *config.BaudRate)
	} else {
		log.Info("Local:        Using stdin/stdout")
	}
	log.Info("--------------------------------------------")
}

// configureSerial will configure the port if we use one. Otherwise it will
// use stdout.
func configureSerial(config *Config) *IOWrapper {
	if *config.UseSerial {
		m := &serial.Mode{
			BaudRate: *config.BaudRate,
			Parity:   serial.NoParity,
			DataBits: 8,
			StopBits: serial.OneStopBit,
		}

		s, err := serial.Open(*config.SerialPort, m)
		if err != nil {
			log.Fatalf("Unable to open serial port '%s': %s", *config.SerialPort, err)
		}

		s.ResetInputBuffer()

		return &IOWrapper{Input: bufio.NewReader(s), Output: s}
	}

	return &IOWrapper{Input: bufio.NewReader(os.Stdin), Output: os.Stdin}
}

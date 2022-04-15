package main

import (
	"fmt"
	"io"
	"os"

	progressbar "github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"
)

// Hex bytes are 2.375 times the count of original file bytes
const RatioBinToHex = 2.375

var (
	// twoToThe16th is used in checking address lengths
	twoToThe16th = 65536
)

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

	progress := progressbar.DefaultBytes(
		int64(float64(*config.Length)*RatioBinToHex),
		"uploading",
	)

	outFileWithProgress := io.MultiWriter(outFile, progress)

	// Begin the dump - command is 'd' plus 32bit hex base and length
	addrStr := fmt.Sprintf("d%08X%08X", *config.BaseAddr, *config.Length)
	log.Infof("AddrStr: %s", addrStr)
	port.Write([]byte(addrStr))
	readAllDataToWriter(port, outFileWithProgress)
	outFile.Close()

	println()
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

package main

import (
	"bufio"
	"bytes"
	log "github.com/sirupsen/logrus"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_AllMethods(t *testing.T) {
	Convey("All methods talking to the board", t, func() {
		logCapture := &bytes.Buffer{}
		log.SetOutput(logCapture)

		sentText := []byte("The folk-kings’ former fame we have heard of," +
			"How princes displayed then their prowess-in-battle.",
		)

		expectedText := []byte(`:020000040000FA
:2000000054686520666F6C6B2D6B696E6773E2809920666F726D65722066616D652077654F
:200020002068617665206865617264206F662C486F77207072696E63657320646973706C39
:2000400061796564207468656E2074686569722070726F776573732D696E2D626174746C77
:02006000652E0B
:00000001FF


`)
		inputBuf := bytes.NewBuffer([]byte("Enter Command:"))
		infile := bytes.NewBuffer(sentText)

		outputBuf := &bytes.Buffer{}
		port := &IOWrapper{
			Input:  bufio.NewReader(inputBuf),
			Output: outputBuf,
		}


		Convey("uploadFile() sends the correct data after receiving 'Enter Command'", func() {
			zero := 0
			config := &Config{
				BaseAddr: &zero,
			}

			err := uploadFile(port, config, infile)
			So(err, ShouldBeNil)
			// Make sure we sent the right command
			So(outputBuf.Bytes()[0:2], ShouldResemble, []byte("\nu"))
			// Make sure we sent the right encoding
			So(outputBuf.Bytes()[2:], ShouldResemble, expectedText)
			// Validate the error logs
			So(logCapture.String(), ShouldNotContainSubstring, "error")
		})

		Convey("eraseFlash() sends the correct command and waits for reply", func() {
			eraseFlash(port)
			So(outputBuf.Bytes()[0:2], ShouldResemble, []byte("\ne"))
			So(logCapture.String(), ShouldContainSubstring, "Completed")
			// Validate the error logs
			So(logCapture.String(), ShouldNotContainSubstring, "error")
		})

		Convey("eraseFlashSector() sends the correct command and waits for reply", func() {
			five := 5;
			config := &Config{
				Sector: &five,
			}

			eraseFlashSector(port, config)
			So(outputBuf.Bytes()[0:2], ShouldResemble, []byte("\ns"))
			So(logCapture.String(), ShouldContainSubstring, "Completed")
			// Validate the error logs
			So(logCapture.String(), ShouldNotContainSubstring, "error")
			So(logCapture.String(), ShouldContainSubstring, "for sector 5")
		})
	})
}


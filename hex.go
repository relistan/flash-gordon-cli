package main

import (
	"encoding/binary"
	"fmt"
)

const (
	// BytesPerLine is the number of binary bytes we'll encode per line of the
	// hex file.
	BytesPerLine = 32

	RecTypeData    = 0 // normal data record
	RecTypeEOF     = 1 // end-of-file record
	RecTypeExtSeg  = 2 // extended segment address record
	RecTypeExtAddr = 4 // extended linear address record (for 32bit records)
)

// checksumFor calculates the checksum byte for the byteslice we pass in
func checksumFor(record []byte) byte {
	var sum byte
	for _, j := range record {
		sum += j
	}

	return (sum ^ 0xFF) + 1
}

// formatRecord prints out a record in the correctly encoded format.
func formatRecord(addr int, recType byte, rec []byte) string {
	// Get first byte of rec len (it's BytesPerLine or less) and pad with space for
	// the 16bit addr as well. We'll overwrite those 0x0s with addr. Follow
	// with the rest of the record.
	allBytes := append([]byte{byte(len(rec)), 0x0, 0x0, recType}, rec...)

	binary.BigEndian.PutUint16(allBytes[1:], uint16(addr))
	checkSum := checksumFor(allBytes)

	return fmt.Sprintf(":%02X%02X", allBytes, checkSum)
}

// Ruby implementation of checksum, for validation:
/// (([the_hex_str].pack('H*').each_byte.inject(0) { |memo, x| memo = (memo & 0xFF) + x } ^ 0xFF) + 1).to_s(16)

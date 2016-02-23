package pn532

import (
	"bytes"
	"errors"

	"github.com/zyxar/berry/core"
)

var (
	ErrPageOutOfRange = errors.New("page out of range")
	ErrNoAckReceived  = errors.New("no ack received")
	ErrInvalidDataLen = errors.New("invalid length of data")
)

// Tries to read an entire 4-byte page at the specified address
// TAG Type       PAGES   USER START    USER STOP
// --------       -----   ----------    ---------
// NTAG 203       42      4             39
// NTAG 213       45      4             39
// NTAG 215       135     4             129
// NTAG 216       231     4             225
func NtagReadPage(device Device, page uint8) ([]byte, error) {
	if page >= 231 {
		return nil, ErrPageOutOfRange
	}
	p := []byte{
		COMMAND_INDATAEXCHANGE,
		1,
		MIFARE_CMD_READ,
		page,
	}
	if !SendCommandCheckAck(device, p, defaultTimeoutMs) {
		return nil, ErrNoAckReceived
	}
	p = make([]byte, 26)
	if p[7] != 0x00 {
		return nil, errors.New("unexpected response")
	}
	// Copy the 4 data bytes to the output buffer
	// Block content starts at byte 9 of a valid response
	// Note that the command actually reads 16 byte or 4
	// pages at a time ... simply discard the last 12 bytes
	return p[8:14], nil
}

// Tries to write an entire 4-byte page at the specified block address
// data should be exactly 4 bytes long
// TAG Type       PAGES   USER START    USER STOP
// --------       -----   ----------    ---------
// NTAG 203       42      4             39
// NTAG 213       45      4             39
// NTAG 215       135     4             129
// NTAG 216       231     4             225
func NtagWritePage(device Device, page uint8, data []byte) error {
	if page < 4 || page > 225 {
		return ErrPageOutOfRange
	}
	if len(data) != 4 {
		return ErrInvalidDataLen
	}
	p := []byte{
		COMMAND_INDATAEXCHANGE,
		1,
		MIFARE_ULTRALIGHT_CMD_WRITE,
		page,
		data[0],
		data[1],
		data[2],
		data[3],
	}
	if !SendCommandCheckAck(device, p, defaultTimeoutMs) {
		return ErrNoAckReceived
	}
	core.Delay(10)
	p = make([]byte, 26)
	device.ReadData(p)
	return nil
}

/*
  Writes an NDEF URI Record starting at the specified page (4..nn)
  Note that this function assumes that the NTAG2xx card is
  already formatted to work as an "NFC Forum Tag".
  The id code (0 = none, 0x01 = "http://www.", etc.)
*/
func NtagWriteNDEFURI(device Device, id uint8, url []byte) (err error) {
	length := byte(len(url))
	if length < 1 || length+1 > 256-12 {
		err = ErrInvalidDataLen
		return
	}
	head := []byte{ // NDEF Lock Control TLV (must be first and always present)
		0x01, // Tag Field (0x01 = Lock Control TLV)
		0x03, // Payload Length (always 3)
		0xA0, // The position inside the tag of the lock bytes (upper 4 = page address, lower 4 = byte offset)
		0x10, // Size in bits of the lock area
		0x44, // Size in bytes of a page and the number of bytes each lock bit can lock (4 bit + 4 bits)
		// NDEF Message TLV - URI Record
		0x03,       // Tag Field (0x03 = NDEF Message)
		length + 5, // Payload Length (not including 0xFE trailer)
		0xD1,       // NDEF Record Header (TNF=0x1: Well known record + SR + ME + MB)
		0x01,       // Type Length for the record type indicator
		length + 1, // Payload length
		0x55,       // Record Type Indicator (0x55 or 'U' = URI Record)
		id,         // URI Prefix (ex. 0x01 = "http://www.")
	}
	if err = NtagWritePage(device, 4, head[:4]); err != nil {
		return err
	}
	if err = NtagWritePage(device, 5, head[4:8]); err != nil {
		return err
	}
	if err = NtagWritePage(device, 6, head[8:12]); err != nil {
		return err
	}
	currentPage := byte(7)
	buf := make([]byte, 4)
	b := bytes.NewBuffer(buf)
	for length > 0 {
		if length < 4 {
			b.Reset()
			b.Write(url[:length])
			b.WriteByte(0xFE)
			err = NtagWritePage(device, currentPage, b.Bytes())
			return
		} else if length == 4 {
			b.Reset()
			b.Write(url[:4])
			if err = NtagWritePage(device, currentPage, b.Bytes()); err != nil {
				return
			}
			b.Reset()
			b.WriteByte(0xFE)
			currentPage++
			err = NtagWritePage(device, currentPage, b.Bytes())
			return
		} else {
			b.Reset()
			b.Write(url[:4])
			if err = NtagWritePage(device, currentPage, b.Bytes()); err != nil {
				return
			}
			currentPage++
			url = url[4:]
			length -= 4
		}
	}
	return nil
}

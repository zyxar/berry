package pn532

import (
	"bytes"
	"errors"

	"github.com/zyxar/berry/core"
)

type deviceSPI struct {
	ss, clk, mosi, miso core.Pin
	tag                 byte    // number of inlisted tag
	key                 [6]byte // Mifare Classic key
	uid                 [8]byte // [len1:uid7] ISO14443A uid
}

var (
	ErrBufferOverflow = errors.New("buffer overflow")
	ErrDeviceNotReady = errors.New("device not ready")
)

func openDeviceSPI(clk, miso, mosi, ss uint8) (device Device, err error) {
	d := &deviceSPI{
		ss:   core.Pin(ss),
		clk:  core.Pin(clk),
		miso: core.Pin(miso),
		mosi: core.Pin(mosi),
	}
	d.ss.Output()
	d.clk.Output()
	d.mosi.Output()
	d.miso.Input()
	d.ss.DigitalWrite(core.LOW)
	// core.Delay(1000)
	if !SendCommandCheckAck(d, []byte{COMMAND_GETFIRMWAREVERSION}, defaultTimeoutMs) {
		err = ErrDeviceNotReady
		return
	}
	d.ss.DigitalWrite(core.HIGH)
	device = d
	return
}

// Writes a command to the PN532, automatically inserting the preamble and required frame details (checksum, len, etc.)
func (id *deviceSPI) WriteCommand(p []byte) {
	length := len(p) + 1
	var checksum byte = PREAMBLE + PREAMBLE + STARTCODE2
	id.ss.DigitalWrite(core.LOW)
	core.Delay(2)
	id.write([]byte{
		PREAMBLE,
		PREAMBLE,
		STARTCODE2,
		byte(length),
		byte((^length) + 1),
		HOSTTOPN532})
	checksum += HOSTTOPN532
	id.write(p)
	for i := 0; i < len(p); i++ {
		checksum += p[i]
	}
	id.write([]byte{^checksum, POSTAMBLE})
	id.ss.DigitalWrite(core.HIGH)
}

// Reads data into p from the PN532 via SPI
func (id *deviceSPI) ReadData(p []byte) {
	id.ss.DigitalWrite(core.LOW)
	core.Delay(2)
	id.write([]byte{SPI_DATAREAD})
	id.read(p)
	id.ss.DigitalWrite(core.HIGH)
	return
}

// Read the SPI ACK signal
func (id *deviceSPI) ReadAck() bool {
	p := make([]byte, 6)
	id.ReadData(p)
	return bytes.Compare(p, ackBytes) == 0
}

// Return true if the PN532 is ready with a response
func (id *deviceSPI) Ready() bool {
	id.ss.DigitalWrite(core.LOW)
	core.Delay(2)
	id.write([]byte{SPI_STATREAD})
	p := make([]byte, 1)
	id.ReadData(p)
	return p[0] == SPI_READY
}

// Waits until the PN532 is ready or timeout
func (id *deviceSPI) WaitReady(timeout int64) bool {
	var timer int64 = 0
	for !id.Ready() {
		if timeout != 0 {
			timer += 10
			if timer > timeout {
				return false
			}
		}
		core.Delay(10)
	}
	return true
}

// Low-level SPI read wrapper
func (id *deviceSPI) read(p []byte) {
	for i := range p {
		id.clk.DigitalWrite(core.HIGH)
		for j := byte(0); j < 8; j++ {
			if id.miso.DigitalRead() == core.HIGH {
				p[i] |= (1 << j)
			}
			id.clk.DigitalWrite(core.LOW)
			id.clk.DigitalWrite(core.HIGH)
		}
	}
}

// Low-level SPI write wrapper
func (id *deviceSPI) write(p []byte) {
	for i := range p {
		id.clk.DigitalWrite(core.HIGH)
		for j := byte(0); j < 8; j++ {
			id.clk.DigitalWrite(core.LOW)
			if p[i]&(1<<j) != 0 {
				id.mosi.DigitalWrite(core.HIGH)
			} else {
				id.mosi.DigitalWrite(core.LOW)
			}
			id.clk.DigitalWrite(core.HIGH)
		}
	}
}

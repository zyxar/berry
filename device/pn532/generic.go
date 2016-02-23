package pn532

import (
	"bytes"
	"encoding/binary"
)

// Checks the firmware version of the PN5xx chip
func FirmwareVersion(device Device) uint32 {
	var p = []byte{COMMAND_GETFIRMWAREVERSION}
	if !SendCommandCheckAck(device, p, defaultTimeoutMs) {
		return 0
	}
	p = make([]byte, 12)
	device.ReadData(p)
	if bytes.Compare(p, responseFirmwarevers) != 0 {
		return 0
	}
	return binary.LittleEndian.Uint32(p[6:10])
}

// Configures the SAM (Secure Access Module)
func SamConfig(device Device) bool {
	p := []byte{
		COMMAND_SAMCONFIGURATION,
		0x01, // normal mode
		0x14, // timeout 50ms * 20 = 1 second
		0x01, // use IRQ pin
	}
	if !SendCommandCheckAck(device, p, defaultTimeoutMs) {
		return false
	}
	p = make([]byte, 8)
	device.ReadData(p)
	return p[5] == 0x15
}

// Sets the MxRtyPassiveActivation byte of the RFConfiguration register
func SetPassiveActivationRetries(device Device, max uint8) bool {
	p := []byte{
		COMMAND_RFCONFIGURATION,
		5,    // Config item 5 (MaxRetries)
		0xFF, // MxRtyATR (default = 0xFF)
		0x01, // MxRtyPSL (default = 0x01)
		max,
	}
	return SendCommandCheckAck(device, p, defaultTimeoutMs)
}

// Sends a command and waits a specified timeout for the ACK
func SendCommandCheckAck(device Device, p []byte, timeout int64) bool {
	device.WriteCommand(p)
	if !device.WaitReady(timeout) {
		return false
	}
	if !device.ReadAck() {
		return false
	}
	return device.WaitReady(timeout)
}

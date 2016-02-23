package pn532

import (
	"bytes"
	"errors"
)

var (
	ErrInvalidADPU     = errors.New("cannot send ADPU")
	ErrNoResponse      = errors.New("no response received")
	ErrNoPreamble      = errors.New("preamble missing")
	ErrInvalidResponse = errors.New("unexpected response")
)

// Waits for an ISO14443A target to enter the field
// ISO14443A card response should be in the following format:
// byte            Description
// -------------   ------------------------------------------
// b0..6           Frame header and preamble
// b7              Tags Found
// b8              Tag Number (only one used in this example)
// b9..10          SENS_RES
// b11             SEL_RES
// b12             NFCID Length
// b13..NFCIDLen   NFCID
func ReadPassiveTargetId(device Device, cardbaudrate uint8, timeout int64) ([]byte, error) {
	p := []byte{
		COMMAND_INLISTPASSIVETARGET,
		1, // max 1 cards at once (we can set this to 2 later)
		cardbaudrate,
	}
	if !SendCommandCheckAck(device, p, timeout) {
		return nil, ErrNoResponse
	}
	p = make([]byte, 20)
	device.ReadData(p)
	if p[7] != 1 {
		return nil, ErrInvalidResponse
	}
	// binary.LittleEndian.Uint16(p[9:11])
	uidLength := p[12]
	uid := make([]byte, uidLength)
	copy(uid, p[13:])
	return uid, nil
}

// Exchanges an APDU with the currently inlisted peer
func DataExchange(device Device, tag byte, data []byte) (response []byte, err error) {
	if len(data) > 62 {
		err = ErrBufferOverflow
		return
	}
	p := make([]byte, len(data)+2)
	p[0] = COMMAND_INDATAEXCHANGE
	p[1] = tag
	copy(p[2:], data)
	if !SendCommandCheckAck(device, p, defaultTimeoutMs) {
		err = ErrInvalidADPU
		return
	}
	if !device.WaitReady(defaultTimeoutMs) {
		err = ErrNoResponse
		return
	}
	p = make([]byte, 64)
	device.ReadData(p)
	if bytes.Compare(p[:3], []byte{0, 0, 0xFF}) != 0 {
		err = ErrNoPreamble
		return
	}
	length := p[3]
	if p[4] != byte(^length+1) {
		err = ErrInvalidDataLen
		return
	}
	if p[5] != PN532TOHOST || p[6] != RESPONSE_INDATAEXCHANGE || p[7]&0x3F != 0 {
		err = ErrInvalidResponse
		return
	}
	length -= 3
	response = p[8 : 8+length]
	return
}

// 'InLists' a passive target. PN532 acting as reader/initiator, peer acting as card/responder
func InListPassiveTarget(device Device) (tag byte, err error) {
	p := []byte{
		COMMAND_INLISTPASSIVETARGET, 1, 0,
	}
	if !SendCommandCheckAck(device, p, defaultTimeoutMs) {
		err = ErrDeviceNotReady
		return
	}
	if !device.WaitReady(30000) {
		err = ErrDeviceNotReady
		return
	}
	p = make([]byte, 64)
	device.ReadData(p)
	if bytes.Compare(p[:3], []byte{0, 0, 0xFF}) != 0 {
		err = ErrNoPreamble
		return
	}
	length := p[3]
	if p[4] != byte(^length+1) {
		err = ErrInvalidDataLen
		return
	}
	if p[5] != PN532TOHOST || p[6] != RESPONSE_INLISTPASSIVETARGET || p[7] != 1 {
		err = ErrInvalidResponse
		return
	}
	tag = p[8]
	return
}

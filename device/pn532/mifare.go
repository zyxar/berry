package pn532

import (
	"errors"
)

var (
	ErrAuthentificationFailed = errors.New("authentification failed")
)

// Mifare Classic methods
func MifareClassicIsFirstBlock(b uint32) bool {
	if b < 128 {
		return b%4 == 0
	}
	return b%16 == 0
}

func MifareClassicIsTrailerBlock(b uint32) bool {
	if b < 128 {
		return (b+1)%4 == 0
	}
	return (b+1)%16 == 0
}

// Tries to authenticate a block of memory on a MIFARE card using the INDATAEXCHANGE command.
func MifareClassicAuthenticateBlock(device Device,
	uid []byte, blockNumber uint8, keyType uint8, key []byte) error {
	// COPY uid & key into device?
	p := make([]byte, len(uid)+10)
	p[0] = COMMAND_INDATAEXCHANGE /* Data Exchange Header */
	p[1] = 1                      /* Max card numbers */
	p[2] = MIFARE_CMD_AUTH_A + (keyType & 0x01)
	p[3] = blockNumber /* Block Number (1K = 0..63, 4K = 0..255 */
	copy(p[4:10], key)
	copy(p[10:], uid)
	if !SendCommandCheckAck(device, p, defaultTimeoutMs) {
		return ErrNoResponse
	}
	p = make([]byte, 12)
	device.ReadData(p)
	if p[7] != 0x00 {
		return ErrAuthentificationFailed
	}
	return nil
}

// Tries to read an entire 16-byte data block at the specified block address
func MifareClassicReadDataBlock(device Device, blockNumber uint8) ([]byte, error) {
	p := []byte{
		COMMAND_INDATAEXCHANGE,
		1,               /* Card number */
		MIFARE_CMD_READ, /* Mifare Read command = 0x30 */
		blockNumber,     /* Block Number (0..63 for 1K, 0..255 for 4K) */
	}
	if !SendCommandCheckAck(device, p, defaultTimeoutMs) {
		return nil, ErrNoResponse
	}
	p = make([]byte, 26)
	device.ReadData(p)
	if p[7] != 0x00 {
		return nil, ErrInvalidResponse
	}
	return p[8:24], nil
}

// Tries to write an entire 16-byte data block at the specified block address
func MifareClassicWriteDataBlock(device Device, blockNumber uint8, data []byte) error {
	if len(data) != 16 {
		return ErrInvalidDataLen
	}
	p := make([]byte, 26)
	p[0] = COMMAND_INDATAEXCHANGE
	p[1] = 1                /* Card number */
	p[2] = MIFARE_CMD_WRITE /* Mifare Write command = 0xA0 */
	p[3] = blockNumber      /* Block Number (0..63 for 1K, 0..255 for 4K) */
	copy(p[4:20], data)
	if !SendCommandCheckAck(device, p[:20], defaultTimeoutMs) {
		return ErrNoResponse
	}
	device.ReadData(p)
	return nil
}

// Formats a Mifare Classic card to store NDEF Records
func MifareClassicFormatNDEF(device Device) (err error) {
	// Note 0xA0 0xA1 0xA2 0xA3 0xA4 0xA5 must be used for key A for the MAD sector in NDEF records (sector 0)
	if err = MifareClassicWriteDataBlock(device, 1,
		[]byte{0x14, 0x01, 0x03, 0xE1, 0x03, 0xE1, 0x03, 0xE1, 0x03, 0xE1, 0x03, 0xE1, 0x03, 0xE1, 0x03, 0xE1}); err != nil {
		return
	}
	if err = MifareClassicWriteDataBlock(device, 2,
		[]byte{0x03, 0xE1, 0x03, 0xE1, 0x03, 0xE1, 0x03, 0xE1, 0x03, 0xE1, 0x03, 0xE1, 0x03, 0xE1, 0x03, 0xE1}); err != nil {
		return
	}
	return MifareClassicWriteDataBlock(device, 3,
		[]byte{0xA0, 0xA1, 0xA2, 0xA3, 0xA4, 0xA5, 0x78, 0x77, 0x88, 0xC1, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})
}

// Writes an NDEF URI Record to the specified sector (1..15)
// Note that this function assumes that the Mifare Classic card is
// already formatted to work as an "NFC Forum Tag" and uses a MAD1
// file system.  You can use the NXP TagWriter app on Android to
// properly format cards for this
func MifareClassicWriteNDEFURI(device Device, sectorNumber uint8, id uint8, url string) (err error) {
	if sectorNumber < 1 || sectorNumber > 15 {
		err = errors.New("invalid sector number")
		return
	}
	length := byte(len(url))
	if length < 1 || length > 38 {
		err = ErrInvalidDataLen
		return
	}
	// Note 0xD3 0xF7 0xD3 0xF7 0xD3 0xF7 must be used for key A in NDEF records
	// Setup the sector buffer (w/pre-formatted TLV wrapper and NDEF message)
	sectorbuffer1 := []byte{0x00, 0x00, 0x03, length + 5, 0xD1, 0x01, length + 1, 0x55, id, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	sectorbuffer2 := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	sectorbuffer3 := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	sectorbuffer4 := []byte{0xD3, 0xF7, 0xD3, 0xF7, 0xD3, 0xF7, 0x7F, 0x07, 0x88, 0x40, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
	if length <= 6 {
		copy(sectorbuffer1[9:], url)
		sectorbuffer1[length+9] = 0xFE
	} else if length == 7 { // 0xFE needs to be wrapped around to next block
		copy(sectorbuffer1[9:], url)
		sectorbuffer2[0] = 0xFE
	} else if (length > 7) && (length <= 22) { // Url fits in two blocks
		copy(sectorbuffer1[9:], url[:7])
		copy(sectorbuffer2, url[7:])
		sectorbuffer2[length-7] = 0xFE
	} else if length == 23 { // 0xFE needs to be wrapped around to final block
		copy(sectorbuffer1[9:], url[:7])
		copy(sectorbuffer2, url[7:])
		sectorbuffer3[0] = 0xFE
	} else { // Url fits in three blocks
		copy(sectorbuffer1[9:], url[:7])
		copy(sectorbuffer2, url[7:23])
		copy(sectorbuffer3, url[23:])
		sectorbuffer3[length-22] = 0xFE
	}
	sectorNumber *= 4
	if err = MifareClassicWriteDataBlock(device, sectorNumber, sectorbuffer1); err != nil {
		return
	}
	if err = MifareClassicWriteDataBlock(device, sectorNumber+1, sectorbuffer2); err != nil {
		return
	}
	if err = MifareClassicWriteDataBlock(device, sectorNumber+2, sectorbuffer3); err != nil {
		return
	}
	return MifareClassicWriteDataBlock(device, sectorNumber+3, sectorbuffer4)
}

// Mifare Ultralight methods

// Tries to read an entire 4-byte page at the specified address
func MifareUltralightReadPage(device Device, page uint8) (data []byte, err error) {
	if page >= 64 {
		err = ErrPageOutOfRange
		return
	}
	p := []byte{
		COMMAND_INDATAEXCHANGE,
		1,               /* Card number */
		MIFARE_CMD_READ, /* Mifare Read command = 0x30 */
		page,
	}
	if !SendCommandCheckAck(device, p, defaultTimeoutMs) {
		err = ErrNoResponse
		return
	}
	p = make([]byte, 26)
	if p[7] != 0x00 {
		err = ErrInvalidResponse
		return
	}
	// Copy the 4 data bytes to the output buffer Block
	// content starts at byte 9 of a valid response
	// Note that the command actually reads 16 byte or 4
	// pages at a time ... simply discard the last 12 bytes
	data = p[8:14]
	return
}

// Tries to write an entire 4-byte page at the specified block address
func MifareUltralightWritePage(device Device, page uint8, data []byte) (err error) {
	if page >= 64 {
		err = ErrPageOutOfRange
		return
	}
	if len(data) != 4 {
		err = ErrInvalidDataLen
		return
	}
	p := []byte{
		COMMAND_INDATAEXCHANGE,
		1, /* Card number */
		MIFARE_ULTRALIGHT_CMD_WRITE, /* Mifare Ultralight Write command = 0xA2 */
		page, /* Page Number (0..63 for most cases) */
		data[0],
		data[1],
		data[2],
		data[3],
	}
	if !SendCommandCheckAck(device, p, defaultTimeoutMs) {
		err = ErrNoResponse
		return
	}
	p = make([]byte, 26)
	device.ReadData(p)
	return nil
}

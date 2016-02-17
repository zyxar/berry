package bus

import (
	"unsafe"
)

const (
	SMBUS_WRITE = iota
	SMBUS_READ

	// SMBus transaction types
	SMBUS_QUICK = iota
	SMBUS_BYTE
	SMBUS_BYTE_DATA
	SMBUS_WORD_DATA
	SMBUS_PROC_CALL
	SMBUS_BLOCK_DATA
	SMBUS_I2C_BLOCK_BROKEN
	SMBUS_BLOCK_PROC_CALL /* SMBus 2.0 */
	SMBUS_I2C_BLOCK_DATA

	// SMBus messages
	SMBUS_BLOCK_MAX     = 32
	SMBUS_I2C_BLOCK_MAX = 32
)

type smbusData struct {
	block [SMBUS_BLOCK_MAX + 2]uint8
}

type smbusIoctlData struct {
	rw   byte
	cmd  uint8
	size int
	data *smbusData
}

func smbusAccess(fd uintptr, rw byte, cmd uint8, size int, data *smbusData) error {
	d := smbusIoctlData{
		rw:   rw,
		cmd:  cmd,
		size: size,
		data: data,
	}
	return ioctl(fd, I2C_SMBUS, uintptr(unsafe.Pointer(&d)))
}

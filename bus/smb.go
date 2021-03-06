package bus

import (
	"encoding/binary"
	"unsafe"

	"github.com/zyxar/berry/sys"
)

const (
	SMBUS_WRITE = iota
	SMBUS_READ
)

const ( // SMBus transaction types
	SMBUS_QUICK = iota
	SMBUS_BYTE
	SMBUS_BYTE_DATA
	SMBUS_WORD_DATA
	SMBUS_PROC_CALL
	SMBUS_BLOCK_DATA
	SMBUS_I2C_BLOCK_BROKEN
	SMBUS_BLOCK_PROC_CALL /* SMBus 2.0 */
	SMBUS_I2C_BLOCK_DATA
)

const ( // SMBus messages
	SMBUS_BLOCK_MAX     = 32
	SMBUS_I2C_BLOCK_MAX = 32
)

const (
	I2C_FUNC_I2C                    = 0x00000001
	I2C_FUNC_10BIT_ADDR             = 0x00000002
	I2C_FUNC_PROTOCOL_MANGLING      = 0x00000004 /* I2C_M_IGNORE_NAK etc. */
	I2C_FUNC_SMBUS_PEC              = 0x00000008
	I2C_FUNC_NOSTART                = 0x00000010 /* I2C_M_NOSTART */
	I2C_FUNC_SMBUS_BLOCK_PROC_CALL  = 0x00008000 /* SMBus 2.0 */
	I2C_FUNC_SMBUS_QUICK            = 0x00010000
	I2C_FUNC_SMBUS_READ_BYTE        = 0x00020000
	I2C_FUNC_SMBUS_WRITE_BYTE       = 0x00040000
	I2C_FUNC_SMBUS_READ_BYTE_DATA   = 0x00080000
	I2C_FUNC_SMBUS_WRITE_BYTE_DATA  = 0x00100000
	I2C_FUNC_SMBUS_READ_WORD_DATA   = 0x00200000
	I2C_FUNC_SMBUS_WRITE_WORD_DATA  = 0x00400000
	I2C_FUNC_SMBUS_PROC_CALL        = 0x00800000
	I2C_FUNC_SMBUS_READ_BLOCK_DATA  = 0x01000000
	I2C_FUNC_SMBUS_WRITE_BLOCK_DATA = 0x02000000
	I2C_FUNC_SMBUS_READ_I2C_BLOCK   = 0x04000000 /* I2C-like block xfer  */
	I2C_FUNC_SMBUS_WRITE_I2C_BLOCK  = 0x08000000 /* w/ 1-byte reg. addr. */

	I2C_FUNC_SMBUS_BYTE       = (I2C_FUNC_SMBUS_READ_BYTE | I2C_FUNC_SMBUS_WRITE_BYTE)
	I2C_FUNC_SMBUS_BYTE_DATA  = (I2C_FUNC_SMBUS_READ_BYTE_DATA | I2C_FUNC_SMBUS_WRITE_BYTE_DATA)
	I2C_FUNC_SMBUS_WORD_DATA  = (I2C_FUNC_SMBUS_READ_WORD_DATA | I2C_FUNC_SMBUS_WRITE_WORD_DATA)
	I2C_FUNC_SMBUS_BLOCK_DATA = (I2C_FUNC_SMBUS_READ_BLOCK_DATA | I2C_FUNC_SMBUS_WRITE_BLOCK_DATA)
	I2C_FUNC_SMBUS_I2C_BLOCK  = (I2C_FUNC_SMBUS_READ_I2C_BLOCK | I2C_FUNC_SMBUS_WRITE_I2C_BLOCK)
)

type smbusData [SMBUS_BLOCK_MAX + 2]uint8

type smbusIoctlData struct {
	rw   uint8
	cmd  uint8
	size int
	data *smbusData
}

func smbusAccess(fd uintptr, rw uint8, cmd uint8, size int, data *smbusData) error {
	d := smbusIoctlData{
		rw:   rw,
		cmd:  cmd,
		size: size,
		data: data,
	}
	return sys.Ioctl(fd, I2C_SMBUS, uintptr(unsafe.Pointer(&d)))
}

func SMBusWriteQuick(fd uintptr, b uint8) error {
	return smbusAccess(fd, b, 0, SMBUS_QUICK, nil)
}

func SMBusRead(fd uintptr, cmd uint8, size int) (b []byte, err error) {
	var data smbusData
	if err = smbusAccess(fd, SMBUS_READ, cmd, size, &data); err != nil {
		return
	}
	switch size {
	case SMBUS_BYTE, SMBUS_BYTE_DATA:
		b = data[:1]
	case SMBUS_WORD_DATA:
		b = data[:2]
	case SMBUS_BLOCK_DATA:
		if l := data[0]; l > 0 {
			b = data[1 : l+1]
		}
	}
	return
}

func SMBusWrite(fd uintptr, cmd uint8, b ...uint8) error {
	length := len(b)
	if length == 0 {
		return smbusAccess(fd, SMBUS_WRITE, cmd, SMBUS_BYTE, nil)
	}
	var data smbusData
	var size int
	switch length {
	case 1:
		data[0] = b[0]
		size = SMBUS_BYTE_DATA
	case 2:
		copy(data[:], b)
		size = SMBUS_WORD_DATA
	default:
		data[0] = uint8(length)
		copy(data[1:], b)
		size = SMBUS_BLOCK_DATA
	}
	return smbusAccess(fd, SMBUS_WRITE, cmd, size, &data)
}

func SMBusProcessCall(fd uintptr, cmd uint8, b uint16) (v uint16, err error) {
	var data smbusData
	binary.LittleEndian.PutUint16(data[:2], b)
	if err = smbusAccess(fd, SMBUS_WRITE, cmd, SMBUS_PROC_CALL, &data); err != nil {
		return
	}
	v = binary.LittleEndian.Uint16(data[:2])
	return
}

func SMBusBlockProcessCall(fd uintptr, cmd uint8, b []byte) (v []byte, err error) {
	var data smbusData
	if len(b) > SMBUS_BLOCK_MAX {
		b = b[:SMBUS_BLOCK_MAX]
	}
	data[0] = uint8(len(b))
	copy(data[1:], b)
	if err = smbusAccess(fd, SMBUS_WRITE, cmd, SMBUS_BLOCK_PROC_CALL, &data); err != nil {
		return
	}
	if size := data[0]; size > 0 {
		v = data[1 : size+1]
	}
	return
}

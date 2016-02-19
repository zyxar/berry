package bus

import (
	"unsafe"

	"github.com/zyxar/berry/sys"
)

const (
	_SPI_IOC_MAGIC = 'k'
	_SPI_DEV0      = "/dev/spidev0.0"
	_SPI_DEV1      = "/dev/spidev0.1"
)

type spiIoctlTransfer struct {
	TxBuf, RxBuf          uint64
	Length, SpeedHz       uint32
	DelayUsecs            uint16
	BitsPerWord, CsChange uint8
	_                     uint32
}

// Read of SPI mode (SPI_MODE_0..SPI_MODE_3)
func SPI_IOC_RD_MODE() uintptr {
	return sys.IOR(_SPI_IOC_MAGIC, 1, 1)
}

// Write of SPI mode (SPI_MODE_0..SPI_MODE_3)
func SPI_IOC_WR_MODE() uintptr {
	return sys.IOW(_SPI_IOC_MAGIC, 1, 1)
}

// Read SPI bit justification
func SPI_IOC_RD_LSB_FIRST() uintptr {
	return sys.IOR(_SPI_IOC_MAGIC, 2, 1)
}

// Write SPI bit justification
func SPI_IOC_WR_LSB_FIRST() uintptr {
	return sys.IOW(_SPI_IOC_MAGIC, 2, 1)
}

// Read SPI device word length (1..N)
func SPI_IOC_RD_BITS_PER_WORD() uintptr {
	return sys.IOR(_SPI_IOC_MAGIC, 3, 1)
}

// Write SPI device word length (1..N)
func SPI_IOC_WR_BITS_PER_WORD() uintptr {
	return sys.IOW(_SPI_IOC_MAGIC, 3, 1)
}

// Read SPI device default max speed hz
func SPI_IOC_RD_MAX_SPEED_HZ() uintptr {
	return sys.IOR(_SPI_IOC_MAGIC, 4, 4)
}

// Write SPI device default max speed hz
func SPI_IOC_WR_MAX_SPEED_HZ() uintptr {
	return sys.IOW(_SPI_IOC_MAGIC, 4, 4)
}

// Write custom SPI message
func SPI_IOC_MESSAGE(n uintptr) uintptr {
	return sys.IOW(_SPI_IOC_MAGIC, 0, uintptr(SPI_MESSAGE_SIZE(n)))
}
func SPI_MESSAGE_SIZE(n uintptr) uintptr {
	if (n * unsafe.Sizeof(spiIoctlTransfer{})) < (1 << sys.IOC_SIZEBITS) {
		return (n * unsafe.Sizeof(spiIoctlTransfer{}))
	}
	return 0
}

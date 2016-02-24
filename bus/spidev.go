package bus

import (
	"io"
	"os"
	"runtime"
	"unsafe"

	"github.com/zyxar/berry/sys"
)

const (
	spiIoctlMAGIC = 'k'
	spiDev0       = "/dev/spidev0.0"
	spiDev1       = "/dev/spidev0.1"
)

type spiIoctlTransfer struct {
	TxBuf, RxBuf          uint64
	Length, SpeedHz       uint32
	DelayUsecs            uint16
	BitsPerWord, CsChange uint8
	_                     uint32
}

type spi struct {
	channel uint8
	speed   uint32
	file    *os.File
}

var spiBPW uint8 = 8

func OpenSPI(channel uint8, speed uint32, mode uint8) (device io.ReadWriteCloser, err error) {
	channel &= 1 // 0 or 1
	mode &= 3    // 0, 1, 2 or 3
	s := &spi{channel: channel, speed: speed}
	defer func() {
		if err != nil {
			s.Close()
		}
	}()
	if channel == 0 {
		if s.file, err = os.OpenFile(spiDev0, os.O_RDWR, 0); err != nil {
			return
		}
	} else {
		if s.file, err = os.OpenFile(spiDev1, os.O_RDWR, 0); err != nil {
			return
		}
	}
	if err = sys.Ioctl(s.file.Fd(), SPI_IOC_WR_MODE(), uintptr(unsafe.Pointer(&mode))); err != nil {
		return
	}
	if err = sys.Ioctl(s.file.Fd(), SPI_IOC_WR_BITS_PER_WORD(), uintptr(unsafe.Pointer(&spiBPW))); err != nil {
		return
	}
	if err = sys.Ioctl(s.file.Fd(), SPI_IOC_WR_MAX_SPEED_HZ(), uintptr(unsafe.Pointer(&speed))); err != nil {
		return
	}
	runtime.SetFinalizer(s, func(this *spi) {
		this.Close()
	})
	device = s
	return
}

func (this *spi) rw(p []byte) (n int, err error) {
	n = len(p)
	var transfer = spiIoctlTransfer{
		TxBuf:       uint64(uintptr(unsafe.Pointer(&p))),
		RxBuf:       uint64(uintptr(unsafe.Pointer(&p))),
		Length:      uint32(n),
		SpeedHz:     this.speed,
		DelayUsecs:  0,
		BitsPerWord: spiBPW,
	}
	err = sys.Ioctl(this.file.Fd(), SPI_IOC_MESSAGE(1), uintptr(unsafe.Pointer(&transfer)))
	return
}

func (this *spi) Read(p []byte) (n int, err error) {
	n, err = this.rw(p)
	return
}

func (this *spi) Write(p []byte) (n int, err error) {
	n, err = this.rw(p)
	return
}

func (this *spi) Close() (err error) {
	if this.file != nil {
		err = this.file.Close()
		this.file = nil
	}
	return
}

// Read of SPI mode (SPI_MODE_0..SPI_MODE_3)
func SPI_IOC_RD_MODE() uintptr {
	return sys.IOR(spiIoctlMAGIC, 1, 1)
}

// Write of SPI mode (SPI_MODE_0..SPI_MODE_3)
func SPI_IOC_WR_MODE() uintptr {
	return sys.IOW(spiIoctlMAGIC, 1, 1)
}

// Read SPI bit justification
func SPI_IOC_RD_LSB_FIRST() uintptr {
	return sys.IOR(spiIoctlMAGIC, 2, 1)
}

// Write SPI bit justification
func SPI_IOC_WR_LSB_FIRST() uintptr {
	return sys.IOW(spiIoctlMAGIC, 2, 1)
}

// Read SPI device word length (1..N)
func SPI_IOC_RD_BITS_PER_WORD() uintptr {
	return sys.IOR(spiIoctlMAGIC, 3, 1)
}

// Write SPI device word length (1..N)
func SPI_IOC_WR_BITS_PER_WORD() uintptr {
	return sys.IOW(spiIoctlMAGIC, 3, 1)
}

// Read SPI device default max speed hz
func SPI_IOC_RD_MAX_SPEED_HZ() uintptr {
	return sys.IOR(spiIoctlMAGIC, 4, 4)
}

// Write SPI device default max speed hz
func SPI_IOC_WR_MAX_SPEED_HZ() uintptr {
	return sys.IOW(spiIoctlMAGIC, 4, 4)
}

// Write custom SPI message
func SPI_IOC_MESSAGE(n uintptr) uintptr {
	return sys.IOW(spiIoctlMAGIC, 0, uintptr(SPI_MESSAGE_SIZE(n)))
}
func SPI_MESSAGE_SIZE(n uintptr) uintptr {
	if (n * unsafe.Sizeof(spiIoctlTransfer{})) < (1 << sys.IOC_SIZEBITS) {
		return (n * unsafe.Sizeof(spiIoctlTransfer{}))
	}
	return 0
}

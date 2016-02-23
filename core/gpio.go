package core

import (
	"errors"
	"os"
	"reflect"
	"syscall"
	"unsafe"

	"github.com/zyxar/berry/sys"
)

var (
	gpio                 []uint32
	pwm                  []uint32
	clk                  []uint32
	pads                 []uint32
	timer                []uint32
	ErrUnknownMode       = errors.New("unknown pin-mode")
	ErrUnimplementedMode = errors.New("unimplemented pin-mode")
	ErrInvalidValue      = errors.New("invalid value")
	ErrInvalidPlatform   = errors.New("invalid platform")
)

func init() {
	if err := setup(); err != nil {
		panic(err.Error())
	}
}

func setup() (err error) {
	var file *os.File
	if file, err = os.OpenFile(DEV_GPIO_MEM, os.O_RDWR|os.O_SYNC|os.O_EXCL, 0); os.IsNotExist(err) {
		file, err = os.OpenFile(DEV_MEM, os.O_RDWR|os.O_SYNC|os.O_EXCL, 0)
	}
	if err != nil {
		return
	}
	defer file.Close()

	var piMemBase int64 = 0x3F000000
	cpuinfo, err := sys.CPUInfo()
	if err != nil {
		return
	}
	switch cpuinfo.Hardware {
	case "BCM2708":
		piMemBase = 0x20000000
	case "BCM2709":
		piMemBase = 0x3F000000
	default:
		err = ErrInvalidPlatform
		return
	}

	var (
		padsMemBase  int64 = piMemBase + 0x00100000
		clockMemBase int64 = piMemBase + 0x00101000
		gpioMemBase  int64 = piMemBase + 0x00200000
		timerMemBase int64 = piMemBase + 0x0000B000
		pwmMemBase   int64 = piMemBase + 0x0020C000
	)

	var mmap = func(base int64) (p []uint32, err error) {
		var mem []byte
		if mem, err = syscall.Mmap(
			int(file.Fd()),
			base,
			MMAP_BLOCK_SIZE,
			syscall.PROT_READ|syscall.PROT_WRITE,
			syscall.MAP_SHARED); err != nil {
			return
		}
		s := *(*reflect.SliceHeader)(unsafe.Pointer(&mem))
		s.Len /= 4
		s.Cap /= 4
		p = *(*[]uint32)(unsafe.Pointer(&s))
		return
	}

	if gpio, err = mmap(gpioMemBase); err != nil {
		return
	}
	if pwm, err = mmap(pwmMemBase); err != nil {
		return
	}
	if clk, err = mmap(clockMemBase); err != nil {
		return
	}
	if pads, err = mmap(padsMemBase); err != nil {
		return
	}
	timer, err = mmap(timerMemBase)
	return
}

type Pin uint8

func (this Pin) Mode(m uint8) (err error) {
	p := uint8(this)
	var sel, shift uint8

	switch m {
	case INPUT, OUTPUT:
		sel := p / 10
		shift := (p % 10) * 3
		gpio[sel] = (gpio[sel] & ^(7 << shift)) | (uint32(m) << shift)
	case PULL_OFF, PULL_DOWN, PULL_UP:
		sel = p/32 + 38
		shift = p & 31
		gpio[37] = uint32(m-PULL_OFF) & 3
		DelayMicroseconds(1)
		gpio[sel] = 1 << shift
		DelayMicroseconds(1)
		gpio[37] = 0
		DelayMicroseconds(1)
		gpio[sel] = 0
		DelayMicroseconds(1)
	case PWM_OUTPUT:
		err = ErrUnimplementedMode
	case GPIO_CLOCK:
		err = ErrUnimplementedMode
	case SOFT_PWM_OUTPUT:
		err = ErrUnimplementedMode
	case SOFT_TONE_OUTPUT:
		err = ErrUnimplementedMode
	case PWM_TONE_OUTPUT:
		err = ErrUnimplementedMode
	default:
		err = ErrUnknownMode
	}
	return
}

func (this Pin) Input() error {
	return this.Mode(INPUT)
}

func (this Pin) Output() error {
	return this.Mode(OUTPUT)
}

func (this Pin) PullUp() error {
	return this.Mode(PULL_UP)
}

func (this Pin) PullDown() error {
	return this.Mode(PULL_DOWN)
}

func (this Pin) PullOff() error {
	return this.Mode(PULL_OFF)
}

func (this Pin) DigitalWrite(v uint8) error {
	p := uint8(this)
	switch v {
	case LOW:
		gpio[p/32+10] = 1 << (p & 31)
	case HIGH:
		gpio[p/32+7] = 1 << (p & 31)
	default:
		return ErrInvalidValue
	}
	return nil
}

func (this Pin) DigitalRead() uint8 {
	p := uint8(this)
	if (gpio[p/32+13] & (1 << p)) != 0 {
		return HIGH
	}
	return LOW
}

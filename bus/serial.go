//+build linux,arm

package bus

import (
	"errors"
	"os"
	"runtime"
	"syscall"
	"unsafe"

	"github.com/zyxar/berry/sys"
)

var (
	errInvalidBaud = errors.New("invalid baud")
)

type serial struct {
	fd uintptr
}

func OpenSerial(device string, baud uint) (s *serial, err error) {
	myBaud := getBaud(baud)
	if myBaud == 0 {
		err = errInvalidBaud
		return
	}
	fd, err := syscall.Open(
		device,
		os.O_RDWR|syscall.O_RDWR|syscall.O_NOCTTY|syscall.O_NDELAY|syscall.O_NONBLOCK,
		0666)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			syscall.Close(fd)
		}
	}()
	term := syscall.Termios{}
	if err = sys.Ioctl(uintptr(fd), syscall.TCGETS, uintptr(unsafe.Pointer(&term))); err != nil {
		return
	}
	term.Ispeed = myBaud
	term.Ospeed = myBaud
	term.Cflag |= (syscall.CLOCAL | syscall.CREAD)
	term.Cflag = uint32(int32(term.Cflag) & ^syscall.PARENB & ^syscall.CSTOPB & ^syscall.CSIZE)
	term.Cflag |= syscall.CS8
	term.Lflag = uint32(int32(term.Lflag) & ^(syscall.ICANON | syscall.ECHO | syscall.ECHOE | syscall.ISIG))
	term.Oflag = uint32(int32(term.Oflag) & ^syscall.OPOST)
	term.Cc[syscall.VMIN] = 0
	term.Cc[syscall.VTIME] = 100
	if err = sys.Ioctl(uintptr(fd), syscall.TCSETS, uintptr(unsafe.Pointer(&term))); err != nil {
		return
	}
	status := 0
	if err = sys.Ioctl(uintptr(fd), syscall.TIOCMGET, uintptr(unsafe.Pointer(&status))); err != nil {
		return
	}
	status |= syscall.TIOCM_DTR | syscall.TIOCM_RTS
	if err = sys.Ioctl(uintptr(fd), syscall.TIOCMSET, uintptr(unsafe.Pointer(&status))); err != nil {
		return
	}

	s = &serial{uintptr(fd)}
	runtime.SetFinalizer(s, func(this *serial) {
		this.Close()
	})
	return
}

func (this *serial) Close() error {
	return syscall.Close(int(this.fd))
}

func (this *serial) Flush() error {
	return sys.Ioctl(this.fd, syscall.TCIOFLUSH, 0)
}

func (this *serial) Write(p []byte) (n int, err error) {
	n, err = syscall.Write(int(this.fd), p)
	return
}

func (this *serial) Read(p []byte) (n int, err error) {
	n, err = syscall.Read(int(this.fd), p)
	return
}

func (this *serial) Available() (n int, err error) {
	err = sys.Ioctl(this.fd, syscall.TIOCINQ, uintptr(unsafe.Pointer(&n)))
	return
}

func getBaud(b uint) uint32 {
	switch b {
	case 50:
		return syscall.B50
	case 75:
		return syscall.B75
	case 110:
		return syscall.B110
	case 134:
		return syscall.B134
	case 150:
		return syscall.B150
	case 200:
		return syscall.B200
	case 300:
		return syscall.B300
	case 600:
		return syscall.B600
	case 1200:
		return syscall.B1200
	case 1800:
		return syscall.B1800
	case 2400:
		return syscall.B2400
	case 4800:
		return syscall.B4800
	case 9600:
		return syscall.B9600
	case 19200:
		return syscall.B19200
	case 38400:
		return syscall.B38400
	case 57600:
		return syscall.B57600
	case 115200:
		return syscall.B115200
	case 230400:
		return syscall.B230400
	default:
	}
	return 0
}

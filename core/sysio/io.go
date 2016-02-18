package sysio

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
)

type Pin struct {
	no uint8
	fd *os.File
}

var (
	ErrInvalidPin = errors.New("invalid pin number")
	ErrPinClosed  = errors.New("pin closed")
)

func OpenPin(n uint8) (p *Pin, err error) {
	if n&PI_GPIO_MASK != 0 {
		err = ErrInvalidPin
		return
	}
	fd, err := os.OpenFile(fmt.Sprintf(_SYSIO_PATH, n), os.O_RDWR, 0)
	if err != nil {
		return
	}
	p = &Pin{n, fd}
	runtime.SetFinalizer(p, func(this *Pin) {
		this.Close()
	})
	return
}

func (this *Pin) Close() {
	if this.fd != nil {
		this.fd.Close()
		this.fd = nil
	}
}

func (this *Pin) Mode(m uint8) (err error) {
	return
}

func (this *Pin) DigitalRead() (b uint8, err error) {
	if this.fd == nil {
		err = ErrPinClosed
		return
	}
	this.fd.Seek(0, os.SEEK_SET)
	v := make([]byte, 1)
	if _, err = this.fd.Read(v); err != nil {
		return
	}
	if v[0] == '0' {
		b = LOW
	} else {
		b = HIGH
	}
	return
}

func (this *Pin) DigitalWrite(b uint8) error {
	if this.fd == nil {
		return ErrPinClosed
	}
	if b == LOW {
		b = '0'
	} else {
		b = '1'
	}
	data := []byte{b, '\n'}
	return writeToFile(this.fd, data)
}

func writeToFile(fd *os.File, data []byte) error {
	fd.Seek(0, os.SEEK_SET)
	n, err := fd.Write(data)
	if err == nil && n < len(data) {
		err = io.ErrShortWrite
	}
	return err
}

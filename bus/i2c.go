// Package bus provides low level control over the linux buses.
package bus

// based on https://github.com/davecheney/i2c/blob/master/i2c.go

import (
	"fmt"
	"os"
	"runtime"
	"syscall"
	"unsafe"
)

const (
	I2C_RETRIES = 0x0701 /* number of times a device address should be polled when not acknowledging */
	I2C_TIMEOUT = 0x0702 /* set timeout in units of 10 ms */
	/* NOTE: Slave address is 7 or 10 bits, but 10-bit addresses
	 * are NOT supported! (due to code brokenness)
	 */
	I2C_SLAVE       = 0x0703 /* Use this slave address */
	I2C_SLAVE_FORCE = 0x0706 /* Use this slave address, even if it is already in use by a driver! */
	I2C_TENBIT      = 0x0704 /* 0 for 7 bit addrs, != 0 for 10 bit */
	I2C_FUNCS       = 0x0705 /* Get the adapter functionality mask */
	I2C_RDWR        = 0x0707 /* Combined R/W transfer (one STOP only) */
	I2C_PEC         = 0x0708 /* != 0 to use PEC with SMBus */
	I2C_SMBUS       = 0x0720 /* SMBus transfer */

	I2C_RDRW_IOCTL_MAX_MSGS = 42
) // from <linux/i2c-dev.h>

// I2C represents a connection to an i2c device.
type I2C struct {
	rc   *os.File
	addr uint
	dev  uint
	mask uint64
}

// New opens a connection to an i2c device.
func NewI2C(addr uint, dev uint) (i *I2C, err error) {
	f, err := os.OpenFile(fmt.Sprintf("/dev/i2c-%d", dev), os.O_RDWR, 0600)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			f.Close()
		}
	}()
	if addr <= 0x7F {
		if err = ioctl(f.Fd(), I2C_TENBIT, 0); err != nil {
			return
		}
	} else if addr <= 0x3FF {
		if err = ioctl(f.Fd(), I2C_TENBIT, 1); err != nil {
			return
		}
	} else {
		err = fmt.Errorf("address overflow: %d", addr)
		return
	}
	if err = ioctl(f.Fd(), I2C_SLAVE, uintptr(addr)); err != nil {
		return
	}
	var mask uint64
	if err = ioctl(f.Fd(), I2C_FUNCS, uintptr(unsafe.Pointer(&mask))); err != nil {
		return
	}
	i = &I2C{f, addr, dev, mask}
	runtime.SetFinalizer(i, func(this *I2C) {
		this.Close()
	})
	return
}

func (this *I2C) Close() {
	if this.rc != nil {
		this.rc.Close()
		this.rc = nil
	}
}

func (this *I2C) Fd() uintptr {
	return this.rc.Fd()
}

func (this *I2C) Mask() uint64 {
	return this.mask
}

// Write sends buf to the remote i2c device. The interpretation of
// the message is implementation dependant.
func (this *I2C) Write(buf ...byte) error {
	_, err := this.rc.Write(buf)
	return err
}

func (this *I2C) Read(b []byte) error {
	_, err := this.rc.Read(b)
	return err
}

func ioctl(fd, cmd, arg uintptr) (err error) {
	_, _, e1 := syscall.Syscall6(syscall.SYS_IOCTL, fd, cmd, arg, 0, 0, 0)
	if e1 != 0 {
		err = e1
	}
	return
}

const I2CCLOCK_CHANGE = 0x0740

func SetBusFreq(hz uint) error {
	if hz > 400000 || hz < 10000 {
		return fmt.Errorf("invalid bus freq: %d", hz)
	}
	f, err := os.OpenFile("/dev/hwi2c", os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	err = ioctl(f.Fd(), I2CCLOCK_CHANGE, uintptr(unsafe.Pointer(&hz)))
	return err
}

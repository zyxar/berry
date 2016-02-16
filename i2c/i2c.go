// Package i2c provides low level control over the linux i2c bus.
// based on https://github.com/davecheney/i2c/blob/master/i2c.go
package i2c

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
) // from <linux/i2c-dev.h>

const (
	I2C_SMBUS_WRITE = iota
	I2C_SMBUS_READ

	// SMBus transaction types
	I2C_SMBUS_QUICK = iota
	I2C_SMBUS_BYTE
	I2C_SMBUS_BYTE_DATA
	I2C_SMBUS_WORD_DATA
	I2C_SMBUS_PROC_CALL
	I2C_SMBUS_BLOCK_DATA
	I2C_SMBUS_I2C_BLOCK_BROKEN
	I2C_SMBUS_BLOCK_PROC_CALL /* SMBus 2.0 */
	I2C_SMBUS_I2C_BLOCK_DATA

	// SMBus messages
	I2C_SMBUS_BLOCK_MAX     = 32 /* As specified in SMBus standard */
	I2C_SMBUS_I2C_BLOCK_MAX = 32 /* Not specified but we use same structure */
)

// I2C represents a connection to an i2c device.
type I2C struct {
	rc   *os.File
	addr uint
	bus  int
}

// New opens a connection to an i2c device.
func New(addr uint, bus int) (i *I2C, err error) {
	f, err := os.OpenFile(fmt.Sprintf("/dev/i2c-%d", bus), os.O_RDWR, 0600)
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
	i = &I2C{f, addr, bus}
	runtime.SetFinalizer(i, func(this *I2C) {
		if this.rc != nil {
			this.rc.Close()
			this.rc = nil
		}
	})
	return
}

// Write sends buf to the remote i2c device. The interpretation of
// the message is implementation dependant.
func (i2c *I2C) Write(buf ...byte) error {
	_, err := i2c.rc.Write(buf)
	return err
}

func (i2c *I2C) Read(b []byte) error {
	_, err := i2c.rc.Read(b)
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

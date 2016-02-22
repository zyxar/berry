// Package bus provides low level control over the linux buses.
/*
I²C Device List
http://www.coris.org.uk/jdc/Notes/i2c-devices.html

                          Device                Address range

Bridge/Expander
                          SUNW,i2c-imax         0x09, 0x0b, 0x0c, 0x12, 0x30
Clock Generator
                          ICS951601             0x6e
                          ICS9FG108             0x6e
Controller
                          PCF8584               [Programmable]
                          Sun JBus I2C          [N/A]
                          Mentor Graphics       [N/A]
                          Sun Fire              [N/A]
                          Marvell MV78230       [N/A]
                          Allwinner Sun4i       [N/A]
                          DECchip 21272         [N/A]
DAC
                          TDA8444               0x40, 0x42, 0x44, 0x46, 0x48, 0x4a, 0x4c, 0x4e
EEPROM
                          AT24C{01,02,04,08,16} 0x50 to 0x57
                          AT24C32, AT24C64      0x50 to 0x57
                          AT34C02               0x50 to 0x57, 0x30 to 0x37
                          PCF8582C              0x50 to 0x57
                          SPD memory            0x50 to 0x57
GPIO
                          FM3560                0x37, 0x4e
                          PCF8574, PCF8574A     0x20 to 0x27, 0x38 to 0x3f
                          PCA9555               0x20 to 0x27
                          PCA9556               0x18 to 0x1f
PIC/Microcontroller
                          PIC16F818/819         0x29
Real-Time Clock
                          DS1307                0x68
Sensor/Hardware Monitor
                          ADM1021               0x18 to 0x1a, 0x29 to 0x2b, 0x4c to 0x4e
                          ADM1021               0x18 to 0x1a, 0x29 to 0x2b, 0x4c to 0x4e
                          ADM1021A              0x18 to 0x1a, 0x29 to 0x2b, 0x4c to 0x4e
                          ADM1023               0x18 to 0x1a, 0x29 to 0x2b, 0x4c to 0x4e
                          ADM1026               0x2c to 0x2f
                          ADM1031               0x2c to 0x2f
                          ADM1032               0x4c, 0x4d
                          ADM9240               0x2c to 0x2f
                          ADT7462               0x5b, 0x5c
                          DS1780                0x2c to 0x2f
                          DS75                  0x48 to 0x4f
                          G781                  0x4c
                          LM75                  0x48 to 0x4f
                          LM75A                 0x48 to 0x4f
                          LM76                  0x48 to 0x4d
                          LM77                  0x48 to 0x4b
                          LM81                  0x2c to 0x2f
                          LM84                  0x18 to 0x1a, 0x29 to 0x2b, 0x4c to 0x4e
                          LM87                  0x2c to 0x2f
                          MAX1617               0x18 to 0x1a, 0x29 to 0x2b, 0x4c to 0x4e
                          MAX1617A              0x18 to 0x1a, 0x29 to 0x2b, 0x4c to 0x4e
                          NE1617ADS             0x18 to 0x1a, 0x29 to 0x2b, 0x4c to 0x4e
Smart Card Interface
                          TDA8020HL             0x40, 0x42, 0x44, 0x46, 0x48, 0x4a, 0x4c, 0x4e
*/
package bus

// based on https://github.com/davecheney/i2c/blob/master/i2c.go

import (
	"fmt"
	"os"
	"runtime"
	"unsafe"

	"github.com/zyxar/berry/sys"
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
		if err = sys.Ioctl(f.Fd(), I2C_TENBIT, 0); err != nil {
			return
		}
	} else if addr <= 0x3FF {
		if err = sys.Ioctl(f.Fd(), I2C_TENBIT, 1); err != nil {
			return
		}
	} else {
		err = fmt.Errorf("address overflow: %d", addr)
		return
	}
	if err = sys.Ioctl(f.Fd(), I2C_SLAVE, uintptr(addr)); err != nil {
		return
	}
	var mask uint64
	if err = sys.Ioctl(f.Fd(), I2C_FUNCS, uintptr(unsafe.Pointer(&mask))); err != nil {
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

// Write sends buf to the i2c device.
func (this *I2C) Write(buf ...byte) error {
	_, err := this.rc.Write(buf)
	return err
}

// Read receives bytes from the i2c device.
func (this *I2C) Read(b []byte) error {
	_, err := this.rc.Read(b)
	return err
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
	err = sys.Ioctl(f.Fd(), I2CCLOCK_CHANGE, uintptr(unsafe.Pointer(&hz)))
	return err
}

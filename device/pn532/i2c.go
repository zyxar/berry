package pn532

import (
	"github.com/zyxar/berry/bus"
	"github.com/zyxar/berry/core"
)

type deviceI2c struct {
	irq, reset core.Pin
	uid        [8]byte // [len1:uid7] ISO14443A uid
	key        [6]byte // Mifare Classic key
	tag        byte    // Tg number of inlisted tag.
	dev        *bus.I2C
}

func openDeviceI2c(irq, rst uint8) (d *deviceI2c, err error) {
	dev, err := bus.NewI2C(0, 0x01)
	if err != nil {
		return
	}
	d = &deviceI2c{
		irq:   core.Pin(irq),
		reset: core.Pin(rst),
		dev:   dev,
	}
	d.irq.Input()
	d.reset.Output()
	d.reset.DigitalWrite(core.HIGH)
	d.reset.DigitalWrite(core.LOW)
	core.Delay(400)
	d.reset.DigitalWrite(core.HIGH)
	core.Delay(10)
	return
}

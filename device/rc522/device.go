package rc522

import (
	"errors"

	"github.com/zyxar/berry/bus"
	"github.com/zyxar/berry/core"
)

var (
	ErrNoTag        = errors.New("no tag found")
	ErrInvalidTag   = errors.New("invalid tag")
	ErrTagCRC       = errors.New("tag crc error")
	ErrTagCollision = errors.New("tag collision")
)

type Device struct {
	dev bus.SPIBus
}

func Open() (*Device, error) {
	dev, err := bus.OpenSPI(0, 5000, 0)
	if err != nil {
		return nil, err
	}
	d := &Device{dev}
	if err = d.Reset(); err != nil {
		return nil, err
	}
	if err = d.EnableAntenna(); err != nil {
		return nil, err
	}
	return d, nil
}

func (id *Device) FindTag() (b []byte, err error) {
	p, err := id.Request(PICC_REQIDL)
	if err != nil {
		return
	}
	b = p[:2]
	return
}

func (id *Device) ReadTag(reg uint8) (b []byte, err error) {
	p, err := id.ReadBytes(reg)
	if err != nil {
		return
	}
	b = p[:16]
	return
}

func (id *Device) ReadByte(reg byte) (byte, error) {
	p := []byte{((reg << 1) & 0x7E) | 0x80, 0}
	_, err := id.dev.Read(p)
	return p[1], err
}

func (id *Device) WriteByte(reg byte, data byte) error {
	p := []byte{(reg << 1) & 0x7E, data}
	_, err := id.dev.Write(p)
	return err
}

func (id *Device) SetMask(reg, mask byte) error {
	b, err := id.ReadByte(reg)
	if err != nil {
		return err
	}
	return id.WriteByte(reg, b|mask)
}

func (id *Device) ClearMask(reg, mask byte) error {
	b, err := id.ReadByte(reg)
	if err != nil {
		return err
	}
	return id.WriteByte(reg, b & ^mask)
}

func (id *Device) Reset() (err error) {
	if err = id.WriteByte(CommandReg, PCD_RESETPHASE); err != nil {
		return
	}
	core.Delay(10)
	if err = id.ClearMask(TxControlReg, 0x03); err != nil {
		return
	}
	core.Delay(10)
	if err = id.SetMask(TxControlReg, 0x03); err != nil {
		return
	}
	if err = id.WriteByte(TModeReg, 0x8D); err != nil {
		return
	}
	if err = id.WriteByte(TPrescalerReg, 0x3E); err != nil {
		return
	}
	if err = id.WriteByte(TReloadRegL, 30); err != nil {
		return
	}
	if err = id.WriteByte(TReloadRegH, 0); err != nil {
		return
	}
	if err = id.WriteByte(TxASKReg, 0x40); err != nil {
		return
	}
	if err = id.WriteByte(ModeReg, 0x3D); err != nil {
		return
	}
	if err = id.WriteByte(RxThresholdReg, 0x84); err != nil {
		return
	}
	if err = id.WriteByte(RFCfgReg, 0x68); err != nil {
		return
	}
	if err = id.WriteByte(GsNReg, 0xff); err != nil {
		return
	}
	err = id.WriteByte(CWGsCfgReg, 0x2f)
	return
}

func (id *Device) Command(cmd uint8, data []byte) (r []byte, err error) {
	var irq, wait uint8 = 0x00, 0x00
	switch cmd {
	case PCD_AUTHENT:
		irq = 0x12
		wait = 0x10
	case PCD_TRANSCEIVE:
		irq = 0x77
		wait = 0x30
	default:
	}
	if err = id.WriteByte(ComIEnReg, irq|0x80); err != nil {
		return
	}
	if err = id.ClearMask(ComIrqReg, 0x80); err != nil {
		return
	}
	if err = id.SetMask(FIFOLevelReg, 0x80); err != nil {
		return
	}
	if err = id.WriteByte(CommandReg, PCD_IDLE); err != nil {
		return
	}
	for i := 0; i < len(data); i++ {
		if err = id.WriteByte(FIFODataReg, data[i]); err != nil {
			return
		}
	}
	if err = id.WriteByte(CommandReg, cmd); err != nil {
		return
	}
	if cmd == PCD_TRANSCEIVE {
		if err = id.SetMask(BitFramingReg, 0x80); err != nil {
			return
		}
	}
	n, err := id.ReadByte(ComIrqReg)
	var i int
	for i = 150; i != 0 && n&0x01 == 0 && n&wait == 0; i-- {
		core.DelayMicroseconds(200)
		if n, err = id.ReadByte(ComIrqReg); err != nil {
			return
		}
	}
	if err = id.ClearMask(BitFramingReg, 0x80); err != nil {
		return
	}
	if i != 0 {
		var pcdErr byte
		if pcdErr, err = id.ReadByte(ErrorReg); err != nil {
			return
		}
		if pcdErr&0x11 == 0 {
			if n&irq&0x01 != 0 {
				err = ErrNoTag
				return
			}
			if cmd == PCD_TRANSCEIVE {
				if n, err = id.ReadByte(FIFOLevelReg); err != nil {
					return
				}
				var lastByte byte
				if lastByte, err = id.ReadByte(ControlReg); err != nil {
					return
				}
				lastByte &= 0x07
				var length byte
				if lastByte != 0 {
					length = (n-1)*8 + lastByte
				} else {
					length = n * 8
				}
				if n == 0 {
					n = 1
				}
				if n > maxLen {
					n = maxLen
				}
				r = make([]byte, n+1)
				for j := byte(0); j < n; j++ {
					r[j], _ = id.ReadByte(FIFODataReg)
				}
				r[n] = length
				return
			}
		} else if pcdErr&0x08 != 0 {
			err = ErrTagCollision
			return
		}
	}
	err = ErrInvalidTag
	return
}

func (id *Device) calculateCrc(p []byte) (b []byte, err error) {
	if err = id.ClearMask(DivIrqReg, 0x04); err != nil {
		return
	}
	if err = id.WriteByte(CommandReg, PCD_IDLE); err != nil {
		return
	}
	if err = id.SetMask(FIFOLevelReg, 0x80); err != nil {
		return
	}
	for i := 0; i < len(p); i++ {
		if err = id.WriteByte(FIFODataReg, p[i]); err != nil {
			return
		}
	}
	if err = id.WriteByte(CommandReg, PCD_CALCCRC); err != nil {
		return
	}
	var n byte
	if n, err = id.ReadByte(DivIrqReg); err != nil {
		return
	}
	for i := byte(0xFE); i != 0 && n&0x04 == 0; i-- {
		if n, err = id.ReadByte(DivIrqReg); err != nil {
			return
		}
	}
	b = make([]byte, 2)
	b[0], err = id.ReadByte(CRCResultRegL)
	b[1], err = id.ReadByte(CRCResultRegM)
	return
}

func (id *Device) EnableAntenna() (err error) {
	var b byte
	if b, err = id.ReadByte(TxControlReg); err != nil {
		return
	}
	if b&0x03 == 0 {
		err = id.SetMask(TxControlReg, 0x03)
	}
	return
}

func (id *Device) DisableAntenna() error {
	return id.ClearMask(TxControlReg, 0x03)
}

func (id *Device) Halt() (err error) {
	p := []byte{PICC_HALT, 0, 0, 0}
	b, err := id.calculateCrc(p[:2])
	if err != nil {
		return
	}
	copy(p[2:], b)
	_, err = id.Command(PCD_TRANSCEIVE, p)
	return
}

func (id *Device) Request(req uint8) (b []byte, err error) {
	if err = id.WriteByte(BitFramingReg, 0x07); err != nil {
		return
	}
	var p = []byte{req, 0, 0}
	if p, err = id.Command(PCD_TRANSCEIVE, p[:1]); err != nil {
		return
	}
	if p[len(p)-1] != 0x10 {
		err = ErrInvalidTag
		return
	}
	b = p[:2]
	return
}

// func (id *Device) Anticoll(cascade uint8, snr []byte) (err error) {

// }

func (id *Device) Select(cascade uint8, snr []byte) (err error) {
	p := make([]byte, 12)
	p[0] = cascade
	p[1] = 0x70
	copy(p[2:6], snr)
	p[6] = snr[0] ^ snr[1] ^ snr[2] ^ snr[3]
	var b []byte
	if b, err = id.calculateCrc(p[:7]); err != nil {
		return
	}
	copy(p[7:9], b)
	if err = id.ClearMask(Status2Reg, 0x08); err != nil {
		return
	}
	if b, err = id.Command(PCD_TRANSCEIVE, p[:9]); err != nil {
		return
	}
	if b[len(b)-1] != 0x18 {
		err = ErrInvalidTag
	}
	return
}

func (id *Device) AuthState(mode, reg uint8, key, snr []byte) (err error) {
	p := make([]byte, 12)
	p[0] = mode
	p[1] = reg
	copy(p[2:8], key)
	copy(p[8:12], snr)
	if _, err = id.Command(PCD_AUTHENT, p); err != nil {
		return
	}
	if p[0], err = id.ReadByte(Status2Reg); err != nil {
		return
	}
	if p[0]&0x08 == 0 {
		err = ErrInvalidTag
	}
	return
}

func (id *Device) ReadBytes(reg uint8) (data []byte, err error) {
	p := []byte{PICC_READ, reg, 0, 0}
	b, err := id.calculateCrc(p[:2])
	if err != nil {
		return
	}
	copy(p[2:], b)
	if b, err = id.Command(PCD_TRANSCEIVE, p); err != nil {
		return
	}
	if b[len(b)-1] != 0x90 {
		err = ErrInvalidTag
		return
	}
	var crc []byte
	if crc, err = id.calculateCrc(b[:16]); err != nil {
		return
	}
	if crc[0] != b[16] || crc[1] != b[17] {
		err = ErrTagCRC
		return
	}
	data = b[:16]
	return
}

func (id *Device) WriteBytes(reg uint8, data []byte) (err error) {
	p := []byte{PICC_WRITE, reg, 0, 0}
	b, err := id.calculateCrc(p[:2])
	if err != nil {
		return
	}
	copy(p[2:], b)
	if b, err = id.Command(PCD_TRANSCEIVE, p); err != nil {
		return
	}
	if b[0]&0x0F != 0x0A || b[len(b)-1] != 4 {
		err = ErrInvalidTag
		return
	}
	copy(b, data[:16])
	if p, err = id.calculateCrc(b[:16]); err != nil {
		return
	}
	copy(b[16:18], p)
	if p, err = id.Command(PCD_TRANSCEIVE, b); err != nil {
		return
	}
	if p[0]&0x0F != 0x0A || p[len(p)-1] != 4 {
		err = ErrInvalidTag
	}
	return
}

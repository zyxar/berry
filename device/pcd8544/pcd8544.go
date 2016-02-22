package pcd8544

import (
	"fmt"

	. "github.com/zyxar/berry/core"
)

// reference:
// https://github.com/adafruit/Adafruit-PCD8544-Nokia-5110-LCD-library

const (
	BLACK               = 1
	WHITE               = 0
	LCDWIDTH            = 84
	LCDHEIGHT           = 48
	POWERDOWN           = 0x04
	ENTRYMODE           = 0x02
	EXTENDEDINSTRUCTION = 0x01
	DISPLAYBLANK        = 0x00
	DISPLAYALLON        = 0x01
	DISPLAYNORMAL       = 0x04
	DISPLAYINVERTED     = 0x05
	DISPLAYCONTROL      = 0x08
	SETTEMP             = 0x04
	SETBIAS             = 0x10
	FUNCTIONSET         = 0x20
	SETYADDR            = 0x40
	SETXADDR            = 0x80
	SETVOP              = 0x80
)

type LCD struct {
	din, clk, dc, rst, cs Pin
	contrast              byte
	x, y                  byte
	size, color           byte
	buffer                []byte
}

func abs(a, b byte) byte {
	if a > b {
		return a - b
	}
	return b - a
}

func OpenLCD(din, clk, dc, rst, cs, contrast byte) (lcd *LCD) {
	b := make([]byte, LCDWIDTH*LCDHEIGHT/8)
	lcd = &LCD{
		din:      Pin(din),
		clk:      Pin(clk),
		dc:       Pin(dc),
		rst:      Pin(rst),
		cs:       Pin(cs),
		contrast: contrast,
		x:        0,
		y:        0,
		size:     1,
		color:    BLACK,
		buffer:   b,
	}
	lcd.Reset()
	return
}

func (this *LCD) Reset() {
	this.din.Output()
	this.clk.Output()
	this.dc.Output()
	this.rst.Output()
	this.cs.Output()
	this.cs.DigitalWrite(LOW)
	this.rst.DigitalWrite(LOW)
	DelayMicroseconds(10)
	this.rst.DigitalWrite(HIGH)
	this.Command(FUNCTIONSET | EXTENDEDINSTRUCTION)
	this.Command(SETBIAS | 0x4)
	this.contrast &= 0x7F
	this.Command(SETVOP | this.contrast)
	this.Command(FUNCTIONSET)
	this.Command(DISPLAYCONTROL | DISPLAYNORMAL)
	DelayMicroseconds(10)
}

func (this *LCD) DrawBitmap(x, y byte, bitmap []byte, w, h, color byte) {
	for j := 0; j < int(h); j++ {
		for i := 0; i < int(w); i++ {
			if bitmap[i+(j/8)*int(w)]&(1<<uint(j%8)) != 0 {
				this.SetPixel(x+byte(i), y+byte(j), color)
			}
		}
	}
}

func (this *LCD) DrawString(x, y byte, c interface{}) {
	switch c.(type) {
	case string:
		v := c.(string)
		this.x = x
		this.y = y
		for i := 0; i < len(v); i++ {
			this.Write(v[i])
		}
	case []byte:
		v := c.([]byte)
		this.x = x
		this.y = y
		for i := 0; i < len(v); i++ {
			this.Write(v[i])
		}
	default:
		this.DrawString(x, y, fmt.Sprintf("%v", c))
	}
}

func (this *LCD) DrawChar(x, y byte, c byte) {
	if y >= LCDHEIGHT {
		return
	}
	if x+5 >= LCDWIDTH {
		return
	}
	var i, j byte
	for i = 0; i < 5; i++ {
		d := font[int(c)*5+int(i)]
		for j = 0; j < 8; j++ {
			if d&(1<<uint(j)) != 0 {
				this.SetPixel(x+i, y+j, this.color)
			} else {
				this.SetPixel(x+i, y+j, 1-this.color)
			}
		}
	}
	for j = 0; j < 8; j++ {
		this.SetPixel(x+5, y+j, 1-this.color)
	}
}

func (this *LCD) Write(c byte) {
	if c == '\n' {
		this.y += this.size * 8
		this.x = 0
	} else if c == '\r' {
	} else {
		this.DrawChar(this.x, this.y, c)
		this.x += this.size * 6
		if this.x >= (LCDWIDTH - 5) {
			this.x = 0
			this.y += 8
		}
		if this.y >= LCDHEIGHT {
			this.y = 0
		}
	}
}

func (this *LCD) SetCursor(x, y byte) {
	this.x = x
	this.y = y
}

func (this *LCD) DrawLine(x0, y0, x1, y1, color byte) {
	p := abs(y1, y0) > abs(x1, x0)
	if p {
		x0, y0 = y0, x0
		x1, y1 = y1, x1
	}
	if x0 > x1 {
		x0, x1 = x1, x0
		y0, y1 = y1, y0
	}
	dx := x1 - x0
	var dy, ystep byte
	if y1 > y0 {
		dy = y1 - y0
		ystep = 1
	} else {
		dy = y0 - y1
		ystep = 0xFF
	}
	er := dx / 2
	for ; x0 <= x1; x0++ {
		if p {
			this.SetPixel(y0, x0, color)
		} else {
			this.SetPixel(x0, y0, color)
		}
		er -= dy
		if er >= 0x80 {
			y0 += ystep
			er += dx
		}
	}
}

func (this *LCD) FillRect(x, y, w, h, color byte) {
	for i := x; i < x+w; i++ {
		for j := y; j < y+h; j++ {
			this.SetPixel(i, j, color)
		}
	}
}

func (this *LCD) DrawRect(x, y, w, h, color byte) {
	for i := x; i < x+w; i++ {
		this.SetPixel(i, y, color)
		this.SetPixel(i, y+h-1, color)
	}
	for i := y; i < y+h; i++ {
		this.SetPixel(x, i, color)
		this.SetPixel(x+w-1, i, color)
	}
}

func (this *LCD) DrawCircle(x0, y0, r, color byte) {
	f := 1 - r
	ddF_x := byte(1)
	ddF_y := -int8(2 * r)
	x := byte(0)
	y := r
	this.SetPixel(x0, y0+r, color)
	this.SetPixel(x0, y0-r, color)
	this.SetPixel(x0+r, y0, color)
	this.SetPixel(x0-r, y0, color)
	for x < y {
		if f < 0x80 {
			y--
			ddF_y += 2
			f += byte(ddF_y)
		}
		x++
		ddF_x += 2
		f += ddF_x
	}
}

func (this *LCD) FillCircle(x0, y0, r, color byte) {
	f := 1 - r
	ddF_x := byte(1)
	ddF_y := -int8(2 * r)
	x := byte(0)
	y := r
	for i := y0 - r; i <= y0+r; i++ {
		this.SetPixel(x0, i, color)
	}
	for x < y {
		if f < 0x80 {
			y--
			ddF_y += 2
			f += byte(ddF_y)
		}
		x++
		ddF_x += 2
		f += ddF_x
		for j := y0 - y; j < y0+y; j++ {
			this.SetPixel(x0+x, j, color)
			this.SetPixel(x0-x, j, color)
		}
		for j := y0 - x; j <= y0+x; j++ {
			this.SetPixel(x0+y, j, color)
			this.SetPixel(x0-y, j, color)
		}
	}
}

func (this *LCD) SetPixel(x, y, color byte) {
	if x >= LCDWIDTH || y >= LCDHEIGHT {
		return
	}
	if color != 0 {
		this.buffer[int(x)+int(y/8)*LCDWIDTH] |= (1 << (y % 8))
	} else {
		this.buffer[int(x)+int(y/8)*LCDWIDTH] &= ^(1 << (y % 8))
	}
}

func (this *LCD) GetPixel(x, y byte) byte {
	if x >= LCDWIDTH || y >= LCDHEIGHT {
		return 0
	}
	return (this.buffer[int(x)+int(y/8)*LCDWIDTH] >> (7 - (y % 8))) & 0x1
}

func (this *LCD) Command(c byte) {
	this.dc.DigitalWrite(LOW)
	ShiftOut(this.din, this.clk, MSBFIRST, c)
}

func (this *LCD) Data(c byte) {
	this.dc.DigitalWrite(HIGH)
	ShiftOut(this.din, this.clk, MSBFIRST, c)
}

func (this *LCD) SetContrast(val byte) {
	val &= 0x7F
	this.contrast = val
	this.Command(FUNCTIONSET | EXTENDEDINSTRUCTION)
	this.Command(SETVOP | val)
	this.Command(FUNCTIONSET)
}

func (this *LCD) Display() {
	for p := 0; p < 6; p++ {
		this.Command(byte(SETYADDR | p))
		this.Command(SETXADDR)
		for col := 0; col < LCDWIDTH; col++ {
			this.Data(this.buffer[LCDWIDTH*p+col])
		}
	}
}

func (this *LCD) Off() {
	for p := 0; p < 6; p++ {
		this.Command(byte(SETYADDR | p))
		this.Command(SETXADDR)
		for col := 0; col < LCDWIDTH; col++ {
			this.Data(0)
		}
	}
	this.din.DigitalWrite(LOW)
	this.clk.DigitalWrite(LOW)
	this.dc.DigitalWrite(LOW)
	this.rst.DigitalWrite(LOW)
	this.cs.DigitalWrite(LOW)
}

func (this *LCD) Clear() {
	for i := 0; i < LCDWIDTH*LCDHEIGHT/8; i++ {
		this.buffer[i] = 0
	}
	this.x = 0
	this.y = 0
}

func (this *LCD) ShowLogo() {
	for p := 0; p < 6; p++ {
		this.Command(byte(SETYADDR | p))
		this.Command(SETXADDR)
		for col := 0; col < LCDWIDTH; col++ {
			this.Data(piLogo[LCDWIDTH*p+col])
		}
	}
}

var font = []byte{
	0x00, 0x00, 0x00, 0x00, 0x00,
	0x3E, 0x5B, 0x4F, 0x5B, 0x3E,
	0x3E, 0x6B, 0x4F, 0x6B, 0x3E,
	0x1C, 0x3E, 0x7C, 0x3E, 0x1C,
	0x18, 0x3C, 0x7E, 0x3C, 0x18,
	0x1C, 0x57, 0x7D, 0x57, 0x1C,
	0x1C, 0x5E, 0x7F, 0x5E, 0x1C,
	0x00, 0x18, 0x3C, 0x18, 0x00,
	0xFF, 0xE7, 0xC3, 0xE7, 0xFF,
	0x00, 0x18, 0x24, 0x18, 0x00,
	0xFF, 0xE7, 0xDB, 0xE7, 0xFF,
	0x30, 0x48, 0x3A, 0x06, 0x0E,
	0x26, 0x29, 0x79, 0x29, 0x26,
	0x40, 0x7F, 0x05, 0x05, 0x07,
	0x40, 0x7F, 0x05, 0x25, 0x3F,
	0x5A, 0x3C, 0xE7, 0x3C, 0x5A,
	0x7F, 0x3E, 0x1C, 0x1C, 0x08,
	0x08, 0x1C, 0x1C, 0x3E, 0x7F,
	0x14, 0x22, 0x7F, 0x22, 0x14,
	0x5F, 0x5F, 0x00, 0x5F, 0x5F,
	0x06, 0x09, 0x7F, 0x01, 0x7F,
	0x00, 0x66, 0x89, 0x95, 0x6A,
	0x60, 0x60, 0x60, 0x60, 0x60,
	0x94, 0xA2, 0xFF, 0xA2, 0x94,
	0x08, 0x04, 0x7E, 0x04, 0x08,
	0x10, 0x20, 0x7E, 0x20, 0x10,
	0x08, 0x08, 0x2A, 0x1C, 0x08,
	0x08, 0x1C, 0x2A, 0x08, 0x08,
	0x1E, 0x10, 0x10, 0x10, 0x10,
	0x0C, 0x1E, 0x0C, 0x1E, 0x0C,
	0x30, 0x38, 0x3E, 0x38, 0x30,
	0x06, 0x0E, 0x3E, 0x0E, 0x06,
	0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x5F, 0x00, 0x00,
	0x00, 0x07, 0x00, 0x07, 0x00,
	0x14, 0x7F, 0x14, 0x7F, 0x14,
	0x24, 0x2A, 0x7F, 0x2A, 0x12,
	0x23, 0x13, 0x08, 0x64, 0x62,
	0x36, 0x49, 0x56, 0x20, 0x50,
	0x00, 0x08, 0x07, 0x03, 0x00,
	0x00, 0x1C, 0x22, 0x41, 0x00,
	0x00, 0x41, 0x22, 0x1C, 0x00,
	0x2A, 0x1C, 0x7F, 0x1C, 0x2A,
	0x08, 0x08, 0x3E, 0x08, 0x08,
	0x00, 0x80, 0x70, 0x30, 0x00,
	0x08, 0x08, 0x08, 0x08, 0x08,
	0x00, 0x00, 0x60, 0x60, 0x00,
	0x20, 0x10, 0x08, 0x04, 0x02,
	0x3E, 0x51, 0x49, 0x45, 0x3E,
	0x00, 0x42, 0x7F, 0x40, 0x00,
	0x72, 0x49, 0x49, 0x49, 0x46,
	0x21, 0x41, 0x49, 0x4D, 0x33,
	0x18, 0x14, 0x12, 0x7F, 0x10,
	0x27, 0x45, 0x45, 0x45, 0x39,
	0x3C, 0x4A, 0x49, 0x49, 0x31,
	0x41, 0x21, 0x11, 0x09, 0x07,
	0x36, 0x49, 0x49, 0x49, 0x36,
	0x46, 0x49, 0x49, 0x29, 0x1E,
	0x00, 0x00, 0x14, 0x00, 0x00,
	0x00, 0x40, 0x34, 0x00, 0x00,
	0x00, 0x08, 0x14, 0x22, 0x41,
	0x14, 0x14, 0x14, 0x14, 0x14,
	0x00, 0x41, 0x22, 0x14, 0x08,
	0x02, 0x01, 0x59, 0x09, 0x06,
	0x3E, 0x41, 0x5D, 0x59, 0x4E,
	0x7C, 0x12, 0x11, 0x12, 0x7C,
	0x7F, 0x49, 0x49, 0x49, 0x36,
	0x3E, 0x41, 0x41, 0x41, 0x22,
	0x7F, 0x41, 0x41, 0x41, 0x3E,
	0x7F, 0x49, 0x49, 0x49, 0x41,
	0x7F, 0x09, 0x09, 0x09, 0x01,
	0x3E, 0x41, 0x41, 0x51, 0x73,
	0x7F, 0x08, 0x08, 0x08, 0x7F,
	0x00, 0x41, 0x7F, 0x41, 0x00,
	0x20, 0x40, 0x41, 0x3F, 0x01,
	0x7F, 0x08, 0x14, 0x22, 0x41,
	0x7F, 0x40, 0x40, 0x40, 0x40,
	0x7F, 0x02, 0x1C, 0x02, 0x7F,
	0x7F, 0x04, 0x08, 0x10, 0x7F,
	0x3E, 0x41, 0x41, 0x41, 0x3E,
	0x7F, 0x09, 0x09, 0x09, 0x06,
	0x3E, 0x41, 0x51, 0x21, 0x5E,
	0x7F, 0x09, 0x19, 0x29, 0x46,
	0x26, 0x49, 0x49, 0x49, 0x32,
	0x03, 0x01, 0x7F, 0x01, 0x03,
	0x3F, 0x40, 0x40, 0x40, 0x3F,
	0x1F, 0x20, 0x40, 0x20, 0x1F,
	0x3F, 0x40, 0x38, 0x40, 0x3F,
	0x63, 0x14, 0x08, 0x14, 0x63,
	0x03, 0x04, 0x78, 0x04, 0x03,
	0x61, 0x59, 0x49, 0x4D, 0x43,
	0x00, 0x7F, 0x41, 0x41, 0x41,
	0x02, 0x04, 0x08, 0x10, 0x20,
	0x00, 0x41, 0x41, 0x41, 0x7F,
	0x04, 0x02, 0x01, 0x02, 0x04,
	0x40, 0x40, 0x40, 0x40, 0x40,
	0x00, 0x03, 0x07, 0x08, 0x00,
	0x20, 0x54, 0x54, 0x78, 0x40,
	0x7F, 0x28, 0x44, 0x44, 0x38,
	0x38, 0x44, 0x44, 0x44, 0x28,
	0x38, 0x44, 0x44, 0x28, 0x7F,
	0x38, 0x54, 0x54, 0x54, 0x18,
	0x00, 0x08, 0x7E, 0x09, 0x02,
	0x18, 0xA4, 0xA4, 0x9C, 0x78,
	0x7F, 0x08, 0x04, 0x04, 0x78,
	0x00, 0x44, 0x7D, 0x40, 0x00,
	0x20, 0x40, 0x40, 0x3D, 0x00,
	0x7F, 0x10, 0x28, 0x44, 0x00,
	0x00, 0x41, 0x7F, 0x40, 0x00,
	0x7C, 0x04, 0x78, 0x04, 0x78,
	0x7C, 0x08, 0x04, 0x04, 0x78,
	0x38, 0x44, 0x44, 0x44, 0x38,
	0xFC, 0x18, 0x24, 0x24, 0x18,
	0x18, 0x24, 0x24, 0x18, 0xFC,
	0x7C, 0x08, 0x04, 0x04, 0x08,
	0x48, 0x54, 0x54, 0x54, 0x24,
	0x04, 0x04, 0x3F, 0x44, 0x24,
	0x3C, 0x40, 0x40, 0x20, 0x7C,
	0x1C, 0x20, 0x40, 0x20, 0x1C,
	0x3C, 0x40, 0x30, 0x40, 0x3C,
	0x44, 0x28, 0x10, 0x28, 0x44,
	0x4C, 0x90, 0x90, 0x90, 0x7C,
	0x44, 0x64, 0x54, 0x4C, 0x44,
	0x00, 0x08, 0x36, 0x41, 0x00,
	0x00, 0x00, 0x77, 0x00, 0x00,
	0x00, 0x41, 0x36, 0x08, 0x00,
	0x02, 0x01, 0x02, 0x04, 0x02,
	0x3C, 0x26, 0x23, 0x26, 0x3C,
	0x1E, 0xA1, 0xA1, 0x61, 0x12,
	0x3A, 0x40, 0x40, 0x20, 0x7A,
	0x38, 0x54, 0x54, 0x55, 0x59,
	0x21, 0x55, 0x55, 0x79, 0x41,
	0x21, 0x54, 0x54, 0x78, 0x41,
	0x21, 0x55, 0x54, 0x78, 0x40,
	0x20, 0x54, 0x55, 0x79, 0x40,
	0x0C, 0x1E, 0x52, 0x72, 0x12,
	0x39, 0x55, 0x55, 0x55, 0x59,
	0x39, 0x54, 0x54, 0x54, 0x59,
	0x39, 0x55, 0x54, 0x54, 0x58,
	0x00, 0x00, 0x45, 0x7C, 0x41,
	0x00, 0x02, 0x45, 0x7D, 0x42,
	0x00, 0x01, 0x45, 0x7C, 0x40,
	0xF0, 0x29, 0x24, 0x29, 0xF0,
	0xF0, 0x28, 0x25, 0x28, 0xF0,
	0x7C, 0x54, 0x55, 0x45, 0x00,
	0x20, 0x54, 0x54, 0x7C, 0x54,
	0x7C, 0x0A, 0x09, 0x7F, 0x49,
	0x32, 0x49, 0x49, 0x49, 0x32,
	0x32, 0x48, 0x48, 0x48, 0x32,
	0x32, 0x4A, 0x48, 0x48, 0x30,
	0x3A, 0x41, 0x41, 0x21, 0x7A,
	0x3A, 0x42, 0x40, 0x20, 0x78,
	0x00, 0x9D, 0xA0, 0xA0, 0x7D,
	0x39, 0x44, 0x44, 0x44, 0x39,
	0x3D, 0x40, 0x40, 0x40, 0x3D,
	0x3C, 0x24, 0xFF, 0x24, 0x24,
	0x48, 0x7E, 0x49, 0x43, 0x66,
	0x2B, 0x2F, 0xFC, 0x2F, 0x2B,
	0xFF, 0x09, 0x29, 0xF6, 0x20,
	0xC0, 0x88, 0x7E, 0x09, 0x03,
	0x20, 0x54, 0x54, 0x79, 0x41,
	0x00, 0x00, 0x44, 0x7D, 0x41,
	0x30, 0x48, 0x48, 0x4A, 0x32,
	0x38, 0x40, 0x40, 0x22, 0x7A,
	0x00, 0x7A, 0x0A, 0x0A, 0x72,
	0x7D, 0x0D, 0x19, 0x31, 0x7D,
	0x26, 0x29, 0x29, 0x2F, 0x28,
	0x26, 0x29, 0x29, 0x29, 0x26,
	0x30, 0x48, 0x4D, 0x40, 0x20,
	0x38, 0x08, 0x08, 0x08, 0x08,
	0x08, 0x08, 0x08, 0x08, 0x38,
	0x2F, 0x10, 0xC8, 0xAC, 0xBA,
	0x2F, 0x10, 0x28, 0x34, 0xFA,
	0x00, 0x00, 0x7B, 0x00, 0x00,
	0x08, 0x14, 0x2A, 0x14, 0x22,
	0x22, 0x14, 0x2A, 0x14, 0x08,
	0xAA, 0x00, 0x55, 0x00, 0xAA,
	0xAA, 0x55, 0xAA, 0x55, 0xAA,
	0x00, 0x00, 0x00, 0xFF, 0x00,
	0x10, 0x10, 0x10, 0xFF, 0x00,
	0x14, 0x14, 0x14, 0xFF, 0x00,
	0x10, 0x10, 0xFF, 0x00, 0xFF,
	0x10, 0x10, 0xF0, 0x10, 0xF0,
	0x14, 0x14, 0x14, 0xFC, 0x00,
	0x14, 0x14, 0xF7, 0x00, 0xFF,
	0x00, 0x00, 0xFF, 0x00, 0xFF,
	0x14, 0x14, 0xF4, 0x04, 0xFC,
	0x14, 0x14, 0x17, 0x10, 0x1F,
	0x10, 0x10, 0x1F, 0x10, 0x1F,
	0x14, 0x14, 0x14, 0x1F, 0x00,
	0x10, 0x10, 0x10, 0xF0, 0x00,
	0x00, 0x00, 0x00, 0x1F, 0x10,
	0x10, 0x10, 0x10, 0x1F, 0x10,
	0x10, 0x10, 0x10, 0xF0, 0x10,
	0x00, 0x00, 0x00, 0xFF, 0x10,
	0x10, 0x10, 0x10, 0x10, 0x10,
	0x10, 0x10, 0x10, 0xFF, 0x10,
	0x00, 0x00, 0x00, 0xFF, 0x14,
	0x00, 0x00, 0xFF, 0x00, 0xFF,
	0x00, 0x00, 0x1F, 0x10, 0x17,
	0x00, 0x00, 0xFC, 0x04, 0xF4,
	0x14, 0x14, 0x17, 0x10, 0x17,
	0x14, 0x14, 0xF4, 0x04, 0xF4,
	0x00, 0x00, 0xFF, 0x00, 0xF7,
	0x14, 0x14, 0x14, 0x14, 0x14,
	0x14, 0x14, 0xF7, 0x00, 0xF7,
	0x14, 0x14, 0x14, 0x17, 0x14,
	0x10, 0x10, 0x1F, 0x10, 0x1F,
	0x14, 0x14, 0x14, 0xF4, 0x14,
	0x10, 0x10, 0xF0, 0x10, 0xF0,
	0x00, 0x00, 0x1F, 0x10, 0x1F,
	0x00, 0x00, 0x00, 0x1F, 0x14,
	0x00, 0x00, 0x00, 0xFC, 0x14,
	0x00, 0x00, 0xF0, 0x10, 0xF0,
	0x10, 0x10, 0xFF, 0x10, 0xFF,
	0x14, 0x14, 0x14, 0xFF, 0x14,
	0x10, 0x10, 0x10, 0x1F, 0x00,
	0x00, 0x00, 0x00, 0xF0, 0x10,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xF0, 0xF0, 0xF0, 0xF0, 0xF0,
	0xFF, 0xFF, 0xFF, 0x00, 0x00,
	0x00, 0x00, 0x00, 0xFF, 0xFF,
	0x0F, 0x0F, 0x0F, 0x0F, 0x0F,
	0x38, 0x44, 0x44, 0x38, 0x44,
	0x7C, 0x2A, 0x2A, 0x3E, 0x14,
	0x7E, 0x02, 0x02, 0x06, 0x06,
	0x02, 0x7E, 0x02, 0x7E, 0x02,
	0x63, 0x55, 0x49, 0x41, 0x63,
	0x38, 0x44, 0x44, 0x3C, 0x04,
	0x40, 0x7E, 0x20, 0x1E, 0x20,
	0x06, 0x02, 0x7E, 0x02, 0x02,
	0x99, 0xA5, 0xE7, 0xA5, 0x99,
	0x1C, 0x2A, 0x49, 0x2A, 0x1C,
	0x4C, 0x72, 0x01, 0x72, 0x4C,
	0x30, 0x4A, 0x4D, 0x4D, 0x30,
	0x30, 0x48, 0x78, 0x48, 0x30,
	0xBC, 0x62, 0x5A, 0x46, 0x3D,
	0x3E, 0x49, 0x49, 0x49, 0x00,
	0x7E, 0x01, 0x01, 0x01, 0x7E,
	0x2A, 0x2A, 0x2A, 0x2A, 0x2A,
	0x44, 0x44, 0x5F, 0x44, 0x44,
	0x40, 0x51, 0x4A, 0x44, 0x40,
	0x40, 0x44, 0x4A, 0x51, 0x40,
	0x00, 0x00, 0xFF, 0x01, 0x03,
	0xE0, 0x80, 0xFF, 0x00, 0x00,
	0x08, 0x08, 0x6B, 0x6B, 0x08,
	0x36, 0x12, 0x36, 0x24, 0x36,
	0x06, 0x0F, 0x09, 0x0F, 0x06,
	0x00, 0x00, 0x18, 0x18, 0x00,
	0x00, 0x00, 0x10, 0x10, 0x00,
	0x30, 0x40, 0xFF, 0x01, 0x01,
	0x00, 0x1F, 0x01, 0x01, 0x1E,
	0x00, 0x19, 0x1D, 0x17, 0x12,
	0x00, 0x3C, 0x3C, 0x3C, 0x3C,
	0x00, 0x00, 0x00, 0x00, 0x00,
}

var piLogo = []byte{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // 0x0010 (16) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xF8, 0xF8, 0xFC, 0xAE, 0x0E, 0x0E, 0x06, 0x0E, 0x06, // 0x0020 (32) pixels
	0xCE, 0x86, 0x8E, 0x0E, 0x0E, 0x1C, 0xB8, 0xF0, 0xF8, 0x78, 0x38, 0x1E, 0x0E, 0x8E, 0x8E, 0xC6, // 0x0030 (48) pixels
	0x0E, 0x06, 0x0E, 0x06, 0x0E, 0x9E, 0xFE, 0xFC, 0xF8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // 0x0040 (64) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // 0x0050 (80) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // 0x0060 (96) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03, 0x0F, 0x0F, 0xFE, // 0x0070 (112) pixels
	0xF8, 0xF0, 0x60, 0x60, 0xE0, 0xE1, 0xE3, 0xF7, 0x7E, 0x3E, 0x1E, 0x1F, 0x1F, 0x1F, 0x3E, 0x7E, // 0x0080 (128) pixels
	0xFB, 0xF3, 0xE1, 0xE0, 0x60, 0x70, 0xF0, 0xF8, 0xBE, 0x1F, 0x0F, 0x07, 0x00, 0x00, 0x00, 0x00, // 0x0090 (144) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // 0x00A0 (160) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // 0x00B0 (176) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80, 0xC0, // 0x00C0 (192) pixels
	0xE0, 0xFC, 0xFE, 0xFF, 0xF3, 0x38, 0x38, 0x0C, 0x0E, 0x0F, 0x0F, 0x0F, 0x0E, 0x3C, 0x38, 0xF8, // 0x00D0 (208) pixels
	0xF8, 0x38, 0x3C, 0x0E, 0x0F, 0x0F, 0x0F, 0x0E, 0x0C, 0x38, 0x38, 0xF3, 0xFF, 0xFF, 0xF8, 0xE0, // 0x00E0 (224) pixels
	0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // 0x00F0 (240) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // 0x0100 (256) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // 0x0110 (272) pixels
	0x00, 0x7F, 0xFF, 0xE7, 0xC3, 0xC1, 0xE0, 0xFF, 0xFF, 0x78, 0xE0, 0xC0, 0xC0, 0xC0, 0xC0, 0xE0, // 0x0120 (288) pixels
	0x60, 0x78, 0x38, 0x3F, 0x3F, 0x38, 0x38, 0x60, 0x60, 0xC0, 0xC0, 0xC0, 0xC0, 0xE0, 0xF8, 0x7F, // 0x0130 (304) pixels
	0xFF, 0xE0, 0xC1, 0xC3, 0xE7, 0x7F, 0x3E, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // 0x0140 (320) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // 0x0150 (336) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // 0x0160 (352) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03, 0x0F, 0x7F, 0xFF, 0xF1, 0xE0, 0xC0, 0x80, 0x01, // 0x0170 (368) pixels
	0x03, 0x9F, 0xFF, 0xF0, 0xE0, 0xE0, 0xC0, 0xC0, 0xC0, 0xC0, 0xC0, 0xE0, 0xE0, 0xF0, 0xFF, 0x9F, // 0x0180 (384) pixels
	0x03, 0x01, 0x80, 0xC0, 0xE0, 0xF1, 0x7F, 0x1F, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // 0x0190 (400) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // 0x01A0 (416) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // 0x01B0 (432) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, // 0x01C0 (448) pixels
	0x03, 0x03, 0x07, 0x07, 0x0F, 0x1F, 0x1F, 0x3F, 0x3B, 0x71, 0x60, 0x60, 0x60, 0x60, 0x60, 0x71, // 0x01D0 (464) pixels
	0x3B, 0x1F, 0x0F, 0x0F, 0x0F, 0x07, 0x03, 0x03, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // 0x01E0 (480) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // 0x01F0 (496) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
}

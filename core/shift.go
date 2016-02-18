package core

func ShiftIn(dataPin, clockPin Pin, bitOrder byte) byte {
	var value byte = 0
	clockPin.Output()
	dataPin.Input()
	for i := uint(0); i < 8; i++ {
		clockPin.DigitalWrite(HIGH)
		if bitOrder == LSBFIRST {
			value |= (dataPin.DigitalRead() << i)
		} else {
			value |= (dataPin.DigitalRead() << (7 - i))
		}
		clockPin.DigitalWrite(LOW)
	}
	return value
}

func ShiftOut(dataPin, clockPin Pin, bitOrder, value byte) {
	clockPin.Output()
	dataPin.Output()
	for i := uint(0); i < 8; i++ {
		if bitOrder == LSBFIRST {
			dataPin.DigitalWrite(((value >> i) & 0x01))
		} else {
			dataPin.DigitalWrite(((value >> (7 - i)) & 0x01))
		}
		clockPin.DigitalWrite(HIGH)
		clockPin.DigitalWrite(LOW)
	}
}

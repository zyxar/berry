package pn532

type Device interface {
	// Low level communication methods
	WriteCommand(p []byte)
	ReadData(p []byte)
	ReadAck() bool
	Ready() bool
	WaitReady(int64) bool
}

func OpenDevice(ss, clk, miso, mosi uint8) (device Device, err error) {
	return openDeviceSPI(clk, miso, mosi, ss)
}

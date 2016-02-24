package rc522

const (
	//MF522 command
	PCD_IDLE       = 0x00
	PCD_AUTHENT    = 0x0E
	PCD_RECEIVE    = 0x08
	PCD_TRANSMIT   = 0x04
	PCD_TRANSCEIVE = 0x0C
	PCD_RESETPHASE = 0x0F
	PCD_CALCCRC    = 0x03

	//Mifare_One
	PICC_REQIDL    = 0x26
	PICC_REQALL    = 0x52
	PICC_ANTICOLL1 = 0x93
	PICC_ANTICOLL2 = 0x95
	PICC_ANTICOLL3 = 0x97
	PICC_AUTHENT1A = 0x60
	PICC_AUTHENT1B = 0x61
	PICC_READ      = 0x30
	PICC_WRITE     = 0xA0
	PICC_DECREMENT = 0xC0
	PICC_INCREMENT = 0xC1
	PICC_RESTORE   = 0xC2
	PICC_TRANSFER  = 0xB0
	PICC_HALT      = 0x50

	//MF522 FIFO
	fifoLEN = 64
	maxLen  = 18
)

const ( //MF522 registers
	_ = iota
	CommandReg
	ComIEnReg
	DivlEnReg
	ComIrqReg
	DivIrqReg
	ErrorReg
	Status1Reg
	Status2Reg
	FIFODataReg
	FIFOLevelReg
	WaterLevelReg
	ControlReg
	BitFramingReg
	CollReg
	_
	_
	ModeReg
	TxModeReg
	RxModeReg
	TxControlReg
	TxASKReg
	TxSelReg
	RxSelReg
	RxThresholdReg
	DemodReg
	_
	_
	MifareReg
	_
	_
	SerialSpeedReg
	_
	CRCResultRegM
	CRCResultRegL
	_
	ModWidthReg
	_
	RFCfgReg
	GsNReg
	CWGsCfgReg
	ModGsCfgReg
	TModeReg
	TPrescalerReg
	TReloadRegH
	TReloadRegL
	TCounterValueRegH
	TCounterValueRegL
	_
	TestSel1Reg
	TestSel2Reg
	TestPinEnReg
	TestPinValueReg
	TestBusReg
	AutoTestReg
	VersionReg
	AnalogTestReg
	TestDAC1Reg
	TestDAC2Reg
	TestADCReg
)

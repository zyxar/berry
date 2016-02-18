package sysio

const (
	LOW = iota
	HIGH

	PI_GPIO_MASK = 0xC0
	_SYSIO_PATH  = "/sys/class/gpio/gpio%d/value"
)

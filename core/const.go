package core

const (
	LOW = iota
	HIGH

	INPUT = iota
	OUTPUT
	PULL_OFF
	PULL_DOWN
	PULL_UP
	PWM_OUTPUT
	GPIO_CLOCK
	SOFT_PWM_OUTPUT
	SOFT_TONE_OUTPUT
	PWM_TONE_OUTPUT

	LSBFIRST = iota
	MSBFIRST

	CHANGE = iota + 1
	FALLING
	RISING

	PI_GPIO_MASK    = 0xFFFFFFC0
	DEV_GPIO_MEM    = "/dev/gpiomem"
	DEV_MEM         = "/dev/mem"
	MMAP_BLOCK_SIZE = 4096
)

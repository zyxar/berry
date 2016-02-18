package core

const (
	LOW = iota
	HIGH
)

const (
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
)

const (
	LSBFIRST = iota
	MSBFIRST
)

const (
	_ = iota
	CHANGE
	FALLING
	RISING
)

const (
	MMAP_BLOCK_SIZE = 4096
	DEV_GPIO_MEM    = "/dev/gpiomem"
	DEV_MEM         = "/dev/mem"
	SYS_SOC_RANGES  = "/sys/firmware/devicetree/base/soc/ranges"
)

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/zyxar/berry/bus"
)

var (
	dev            = flag.Uint("bus", 1, "specify i2c bus")
	commands       map[string]func(...string) error
	fns            []fn
	errUnsupported = errors.New("unsupported command")
	errNoRegister  = errors.New("no register specified")
	errNoAddr      = errors.New("no i2c-bus address speified")
)

type fn struct {
	code uint32
	name string
}

func init() {
	commands = map[string]func(...string) error{
		"scan":      scan,
		"probe":     factory("probe"),
		"read":      factory("read"),
		"readdata":  factory("readdata"),
		"readword":  factory("readword"),
		"readblock": factory("readblock"),
	}

	fns = []fn{
		fn{bus.I2C_FUNC_I2C, "I2C"},
		fn{bus.I2C_FUNC_10BIT_ADDR, "10BIT_ADDR"},
		fn{bus.I2C_FUNC_PROTOCOL_MANGLING, "PROTOCOL_MANGLING"},
		fn{bus.I2C_FUNC_SMBUS_PEC, "SMBUS_PEC"},
		fn{bus.I2C_FUNC_NOSTART, "NOSTART"},
		fn{bus.I2C_FUNC_SMBUS_BLOCK_PROC_CALL, "SMBUS_BLOCK_PROC_CALL"},
		fn{bus.I2C_FUNC_SMBUS_QUICK, "SMBUS_QUICK"},
		fn{bus.I2C_FUNC_SMBUS_READ_BYTE, "SMBUS_READ_BYTE"},
		fn{bus.I2C_FUNC_SMBUS_WRITE_BYTE, "SMBUS_WRITE_BYTE"},
		fn{bus.I2C_FUNC_SMBUS_READ_BYTE_DATA, "SMBUS_READ_BYTE_DATA"},
		fn{bus.I2C_FUNC_SMBUS_WRITE_BYTE_DATA, "SMBUS_WRITE_BYTE_DATA"},
		fn{bus.I2C_FUNC_SMBUS_READ_WORD_DATA, "SMBUS_READ_WORD_DATA"},
		fn{bus.I2C_FUNC_SMBUS_WRITE_WORD_DATA, "SMBUS_WRITE_WORD_DATA"},
		fn{bus.I2C_FUNC_SMBUS_PROC_CALL, "SMBUS_PROC_CALL"},
		fn{bus.I2C_FUNC_SMBUS_READ_BLOCK_DATA, "SMBUS_READ_BLOCK_DATA"},
		fn{bus.I2C_FUNC_SMBUS_WRITE_BLOCK_DATA, "SMBUS_WRITE_BLOCK_DATA"},
		fn{bus.I2C_FUNC_SMBUS_READ_I2C_BLOCK, "SMBUS_READ_I2C_BLOCK"},
		fn{bus.I2C_FUNC_SMBUS_WRITE_I2C_BLOCK, "SMBUS_WRITE_I2C_BLOCK"},
	}
}

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Println("usage: i2cbus [options] {COMMAND}")
		flag.PrintDefaults()
		fmt.Println("\navailable commands:")
		for cmd, _ := range commands {
			fmt.Printf("\t%q\n", cmd)
		}
		fmt.Println()
		os.Exit(1)
	}
	var err error
	args := flag.Args()
	if fn, ok := commands[args[0]]; ok && fn != nil {
		if err = fn(args[1:]...); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
	} else {
		fmt.Fprintln(os.Stderr, errUnsupported)
	}
}

func factory(args ...string) func(args ...string) error {
	var cmd string = args[0]
	return func(args ...string) error {
		if len(args) == 0 {
			return errNoAddr
		}
		addr, err := strconv.ParseUint(args[0], 10, 16)
		if err != nil {
			return err
		}
		s, err := bus.NewI2C(uint(addr), *dev)
		if err != nil {
			return err
		}
		defer s.Close()
		mask := s.Mask()
		fmt.Printf("bus: %d, addr: 0x%02x, mask: 0x%08X\n", *dev, addr, mask)
		switch cmd {
		case "read":
			b, err := bus.SMBusReadByte(s.Fd())
			if err != nil {
				return err
			}
			fmt.Printf("%X\n", b)
		case "readdata":
			if len(args) < 2 {
				return errNoRegister
			}
			b := make([]uint8, len(args)-1)
			for i, arg := range args[1:] {
				reg, err := strconv.ParseUint(arg, 10, 8)
				if err != nil {
					return err
				}
				if b[i], err = bus.SMBusReadByteData(s.Fd(), uint8(reg)); err != nil {
					return err
				}
			}
			fmt.Printf("%X\n", b)
		case "readword":
			if len(args) < 2 {
				return errNoRegister
			}
			b := make([]uint16, len(args)-1)
			for i, arg := range args[1:] {
				reg, err := strconv.ParseUint(arg, 10, 8)
				if err != nil {
					return err
				}
				if b[i], err = bus.SMBusReadWordData(s.Fd(), uint8(reg)); err != nil {
					return err
				}
			}
			fmt.Printf("%X\n", b)
		case "readblock":
			if len(args) < 2 {
				return errNoRegister
			}
			reg, err := strconv.ParseUint(args[1], 10, 8)
			if err != nil {
				return err
			}
			b, err := bus.SMBusReadBlockData(s.Fd(), uint8(reg))
			if err != nil {
				return err
			}
			fmt.Printf("%X\n", b)
		case "probe":
			for i, _ := range fns {
				if uint64(fns[i].code)&mask != 0 {
					fmt.Printf("\t[x] %s\n", fns[i].name)
				} else {
					fmt.Printf("\t[ ] %s\n", fns[i].name)
				}
			}
		default:
			return errUnsupported
		}
		return nil
	}
}

func scan(...string) error {
	var addr uint
	fmt.Println("     0  1  2  3  4  5  6  7  8  9  a  b  c  d  e  f")
	for addr = 0x00; addr < 0x77; addr++ {
		if addr%16 == 0 {
			fmt.Printf("%02x: ", addr/16)
		}
		s, err := bus.NewI2C(addr, *dev)
		if err != nil {
			return err
		}
		// if ((addr < 0x30) || (addr >= 0x40 && addr <= 0x47) || (addr >= 0x60)) && s.Mask()&bus.I2C_FUNC_SMBUS_QUICK != 0 {
		// 	if err = bus.SMBusWriteQuick(s.Fd(), bus.SMBUS_WRITE); err != nil {
		// 		fmt.Print("~~ ")
		// 	} else {
		// 		fmt.Print("[x]")
		// 	}
		// } else {
		b, err := bus.SMBusReadByte(s.Fd())
		if err != nil {
			fmt.Print("-- ")
		} else {
			fmt.Printf("%02x ", b)
		}
		// }
		if addr%16 == 15 {
			fmt.Println()
		}
	}
	fmt.Println()
	return nil
}

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/zyxar/berry/bus"
)

var (
	dev   = flag.Uint("bus", 1, "specify i2c bus")
	quick = flag.Bool("q", false, "enable quick mode")
)

func main() {
	flag.Parse()
	var addr uint
	fmt.Println("     0  1  2  3  4  5  6  7  8  9  a  b  c  d  e  f")
	for addr = 0x00; addr < 0x77; addr++ {
		if addr%16 == 0 {
			fmt.Printf("%02x: ", addr/16)
		}
		s, err := bus.NewI2C(addr, *dev)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		/*if !*quick && ((addr >= 0x30 && addr <= 0x37) || (addr >= 0x50 && addr <= 0x5F)) */
		_, err = bus.SMBusReadByte(s.Fd())
		if err != nil {
			fmt.Print("-- ")
		} else {
			fmt.Printf("%02x ", addr)
		}
		if addr%16 == 15 {
			fmt.Println()
		}
	}
	fmt.Println()
}

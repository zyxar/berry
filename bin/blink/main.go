package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	. "github.com/zyxar/berry/core"
)

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Println("usage: blink PIN")
		os.Exit(1)
	}
	v, err := strconv.Atoi(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	p := Pin(v)
	p.Output()
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for s := range ch {
			fmt.Printf("%v caught, exit\n", s)
			p.DigitalWrite(LOW)
			os.Exit(0)
		}
	}()
	for {
		p.DigitalWrite(HIGH)
		Delay(500)
		p.DigitalWrite(LOW)
		Delay(500)
	}
}

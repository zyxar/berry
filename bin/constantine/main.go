//+build linux,arm

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zyxar/berry/core"
)

var (
	quit chan byte
	now  time.Time
)

func init() {
	quit = make(chan byte)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for s := range ch {
			fmt.Printf("%v caught, exit\n", s)
			if lcd != nil {
				lcd.Reset()
			}
			close(quit)
		}
	}()
}

func main() {
	flag.Parse()
	fmt.Println("[Constantine]")
	initLcd()
	initService()
	if err := initClock(); err != nil {
		printToLcd(fmt.Sprintf("RTC Error: %v", err))
	} else {
		printToLcd("RTC ready.")
	}
	core.Delay(500)
	go clockRoutine()
	go lcdRoutine()
	<-quit
}

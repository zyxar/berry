//+build linux,arm

package main

import (
	"fmt"
	"time"

	"github.com/zyxar/berry/core"
	"github.com/zyxar/berry/device/ds1307"
)

var (
	clock  *ds1307.Clock
	addrid uint = 0x68
	busid  uint = 0x01
)

func initClock() (err error) {
	clock, err = ds1307.New(addrid, busid)
	return
}

func clockRoutine() {
	if clock != nil {
		var err error
		for {
			if now, err = clock.Get(); err != nil {
				printToLcd(fmt.Sprintf("err in read clock: %v", err))
			}
			core.Delay(step)
		}
	} else {
		for {
			now = time.Now()
			core.Delay(step)
		}
	}
}

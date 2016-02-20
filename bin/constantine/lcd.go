//+build linux,arm

package main

import (
	"fmt"

	"github.com/zyxar/berry/core"
	"github.com/zyxar/berry/device/pcd8544"
	"github.com/zyxar/berry/sys"
)

var (
	lcd  *pcd8544.LCD
	step int64 = 1000
)

const (
	lcd_width  = 84 / 6
	lcd_height = 48 / 8
	lcd_size   = lcd_height * lcd_width
)

func initLcd() {
	lcd = pcd8544.OpenLCD(19, 26, 13, 5, 6, 60)
}

func printToLcd2(s string) {
	lcd.Clear()
	line := byte(0)
	for len(s) > 14 {
		lcd.DrawString(0, line, s[:14])
		line += 8
		s = s[14:]
	}
	if len(s) > 0 {
		lcd.DrawString(0, line, s)
	}
	lcd.Display()
}

func printToLcd(s string) {
	for len(s) > lcd_size {
		printToLcd2(s[:lcd_size])
		s = s[lcd_size:]
		core.Delay(1000)
	}
	if len(s) > 0 {
		printToLcd2(s)
	}
}

func lcdRoutine() {
	for {
		if sysinfo, err := sys.Info(); err == nil {
			lcd.Clear()
			lcd.DrawString(0, 0, hostname)
			lcd.DrawLine(0, 9, 83, 9, pcd8544.BLACK)
			lcd.DrawString(0, 12, addr)
			lcd.DrawString(0, 20, "UP "+sysinfo.Uptime.String())
			lcd.DrawString(0, 28, fmt.Sprintf("LD %2.1f %2.1f %2.1f", sysinfo.Loads[0], sysinfo.Loads[1], sysinfo.Loads[2]))
			lcd.DrawString(0, 36, fmt.Sprintf("%v %.2d:%.2d:%.2d", now.Weekday().String()[:3], now.Hour(), now.Minute(), now.Second()))
			lcd.DrawLine(0, 45, 83, 45, pcd8544.BLACK)
			lcd.Display()
		} else {
			printToLcd(fmt.Sprint(err))
		}
		core.Delay(step)
	}
}

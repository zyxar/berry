package core

import (
	"time"
)

func Millis() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func Micros() int64 {
	return time.Now().UnixNano() / int64(time.Microsecond)
}

func Delay(ms int64) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

func DelayMicroseconds(us int64) {
	time.Sleep(time.Duration(us) * time.Microsecond)
}

func DelayShed(ms int64) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

func DelayMicrosecondsSched(us int64) {
	time.Sleep(time.Duration(us) * time.Microsecond)
}

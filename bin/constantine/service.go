//+build linux,arm

package main

import (
	"net"
	"os"
)

var (
	hostname string
	addr     string
)

func initService() {
	hostname, _ = os.Hostname()
	if hostname == "" {
		hostname = "RaspberryPi"
	}
	hostname += ":"
	if conn, err := net.Dial("udp", "google.com:80"); err != nil {
		addr = "127.0.0.1"
	} else {
		addr = conn.LocalAddr().String()
		for i := 0; i < len(addr); i++ {
			if addr[i] == ':' {
				addr = addr[:i]
				break
			}
		}
	}
}

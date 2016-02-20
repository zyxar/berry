//+build linux

package main

// ref: https://github.com/capnm/sysinfo/blob/master/sysinfo.go

import (
	"fmt"
	"sync"
	"syscall"
	"time"
)

type Sysinfo struct {
	Uptime       time.Duration // time since boot
	Loads        [3]float64    // 1, 5, and 15 minute load averages, see e.g. UPTIME(1)
	Procs        uint64        // number of current processes
	TotalRam     uint64        // total usable main memory size [kB]
	FreeRam      uint64        // available memory size [kB]
	SharedRam    uint64        // amount of shared memory [kB]
	BufferRam    uint64        // memory used by buffers [kB]
	TotalSwap    uint64        // total swap space size [kB]
	FreeSwap     uint64        // swap space still available [kB]
	TotalHighRam uint64        // total high memory size [kB]
	FreeHighRam  uint64        // available high memory size [kB]
	sync.Mutex                 // ensures atomic writes; protects the following fields
}

const scale = 65536.0 // magic

var singelton *Sysinfo = &Sysinfo{}

func Read() (*Sysinfo, error) {
	rawsysinfo := &syscall.Sysinfo_t{}
	if err := syscall.Sysinfo(rawsysinfo); err != nil {
		return nil, err
	}
	defer singelton.Unlock()
	singelton.Lock()
	unit := uint64(rawsysinfo.Unit) * 1024 // kB
	singelton.Uptime = time.Duration(rawsysinfo.Uptime) * time.Second
	singelton.Loads[0] = float64(rawsysinfo.Loads[0]) / scale
	singelton.Loads[1] = float64(rawsysinfo.Loads[1]) / scale
	singelton.Loads[2] = float64(rawsysinfo.Loads[2]) / scale
	singelton.Procs = uint64(rawsysinfo.Procs)

	singelton.TotalRam = uint64(rawsysinfo.Totalram) / unit
	singelton.FreeRam = uint64(rawsysinfo.Freeram) / unit
	singelton.BufferRam = uint64(rawsysinfo.Bufferram) / unit
	singelton.TotalSwap = uint64(rawsysinfo.Totalswap) / unit
	singelton.FreeSwap = uint64(rawsysinfo.Freeswap) / unit
	singelton.TotalHighRam = uint64(rawsysinfo.Totalhigh) / unit
	singelton.FreeHighRam = uint64(rawsysinfo.Freehigh) / unit
	return singelton, nil
}

func (si Sysinfo) String() string {
	si.Lock()
	r := fmt.Sprintf("uptime\t\t%v\nload\t\t%2.2f %2.2f %2.2f\nprocs\t\t%d\n"+
		"ram  total\t%d kB\nram  free\t%d kB\nram  buffer\t%d kB\n"+
		"swap total\t%d kB\nswap free\t%d kB",
		//"high ram total\t%d kB\nhigh ram free\t%d kB\n"
		si.Uptime, si.Loads[0], si.Loads[1], si.Loads[2], si.Procs,
		si.TotalRam, si.FreeRam, si.BufferRam,
		si.TotalSwap, si.FreeSwap,
		// archaic si.TotalHighRam, si.FreeHighRam
	)
	si.Unlock()
	return r
}

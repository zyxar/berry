package ds1307

import (
	"fmt"
	"sync"
	"time"

	"github.com/zyxar/berry/bus"
)

type Clock struct {
	h *bus.I2C
	m *sync.Mutex
}

var clockTable map[uint64]*Clock
var mutex = sync.Mutex{} // protect table

func init() {
	clockTable = make(map[uint64]*Clock)
}

func New(addr, name uint) (*Clock, error) {
	defer mutex.Unlock()
	mutex.Lock()
	if k, ok := clockTable[addrKey(addr, name)]; ok {
		return k, nil
	}
	i, err := bus.NewI2C(addr, int(name))
	if err != nil {
		return nil, err
	}
	r := &Clock{i, &sync.Mutex{}}
	clockTable[addrKey(addr, name)] = r
	return r, nil
}

func (this *Clock) Get() (*time.Time, error) {
	defer this.m.Unlock()
	this.m.Lock()
	this.h.Write(0)
	b := make([]byte, 7)
	if err := this.h.Read(b); err != nil {
		return nil, err
	}
	// A few of these need masks because certain bits are control bits
	second := bcdToDec(b[0] & 0x7f)
	minute := bcdToDec(b[1])
	hour := bcdToDec(b[2] & 0x3f) // Need to change this if 12 hour am/pm
	dayOfMonth := bcdToDec(b[4])
	month := bcdToDec(b[5])
	year := bcdToDec(b[6])
	_, zone := time.Now().Zone()
	t, err := time.Parse(time.RFC3339, fmt.Sprintf("20%.2d-%.2d-%.2dT%.2d:%.2d:%.2d%+.2d:00", year, month, dayOfMonth, hour, minute, second, zone/3600))
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (this *Clock) Set(now *time.Time) error {
	second, minute, hour, dayOfWeek, dayOfMonth, month, year := byte(now.Second()), byte(now.Minute()), byte(now.Hour()), byte(now.Weekday()), byte(now.Day()), byte(now.Month()), byte(now.Year()%2000)
	defer this.m.Unlock()
	this.m.Lock()
	return this.h.Write(0,
		decToBcd(second),
		decToBcd(minute),
		decToBcd(hour),
		decToBcd(dayOfWeek),
		decToBcd(dayOfMonth),
		decToBcd(month),
		decToBcd(year))
}

func decToBcd(val byte) byte {
	return ((val / 10 * 16) + (val % 10))
}

func bcdToDec(val byte) byte {
	return ((val / 16 * 10) + (val % 16))
}

func addrKey(addr, name uint) uint64 {
	return (uint64(addr) << 32) | uint64(name)
}

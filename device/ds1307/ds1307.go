/*
http://datasheets.maximintegrated.com/en/ds/DS1307.pdf

DS1307 real-time clock is a lowpower, full binary-coded decimal clock/calendar plus 56 bytes of NV SRAM. Address and data are transferred serially through an I2
C, bidirectional bus.
The clock/calendar provides seconds, minutes, hours, day, date, month, and year information. The end of the month date is automatically adjusted for months with fewer than 31 days, including corrections for leap
year. The clock operates in either the 24-hour or 12-hour format with AM/PM indicator.
The DS1307 has a built-in power-sense circuit that detects power failures and automatically switches to the backup supply.
Timekeeping operation continues while the part operates from the backup supply.

        Timekeeper Registers
ADDRESS|BIT 7|BIT 6|BIT 5|BIT 4|BIT 3|BIT 2|BIT 1|BIT 0|FUNCTION |RANGE
00h    | CH  |   10 Seconds    | Seconds               | Seconds |00–59
01h    | 0   |   10 Minutes    | Minutes               | Minutes |00–59
02h    | 0   | 12  |10H  | 10H | Hours                 | Hours   |1–12
02h    | 0   | 24  |PM/AM| 10H | Hours                 | Hours   |00-23
03h    | 0   | 0   | 0   | 0   | 0   | DAY             | Day     |01–07
04h    | 0   | 0   | 10 Date   | Date                  | Date    |01–31
05h    | 0   | 0   | 0   | 10M | Month                 | Month   |01–12
06h    | 10 Year               | Year                  | Year    |00–99
07h    | OUT | 0   | 0   | SQWE| 0   |   0 | RS1 | RS0 | Control |—
08h–3Fh|                                               |RAM56 x 8|00h–FFh

0 = Always reads back as 0.

*/
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

func New(addr, dev uint) (*Clock, error) {
	defer mutex.Unlock()
	mutex.Lock()
	if k, ok := clockTable[addrKey(addr, dev)]; ok {
		return k, nil
	}
	i, err := bus.NewI2C(addr, dev)
	if err != nil {
		return nil, err
	}
	r := &Clock{i, &sync.Mutex{}}
	clockTable[addrKey(addr, dev)] = r
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

func addrKey(addr, dev uint) uint64 {
	return (uint64(addr) << 32) | uint64(dev)
}

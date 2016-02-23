package sys

import (
	"bufio"
	"io"
	"os"
	"strconv"
	"strings"
)

type coreinfo struct {
	Processor uint
	ModelName string
	BogoMIPS  float64
	Features  []string
	CPU       struct {
		Implementer  byte
		Architecture byte
		Variant      byte
		Part         uint16
		Revision     byte
	}
}

type cpuinfo struct {
	Cores    []coreinfo
	Hardware string
	Revision uint64
	Serial   []byte
}

func CPUInfo() (c *cpuinfo, err error) {
	var file *os.File
	if file, err = os.Open("/proc/cpuinfo"); err != nil {
		return
	}
	c, err = decodeCpuInfo(file)
	return
}

func decodeCpuInfo(r io.Reader) (*cpuinfo, error) {
	rd := bufio.NewReader(r)
	var _processors []uint
	var _models []string
	var _bogomips []float64
	var _features [][]string
	var _cpu_implementer []byte
	var _cpu_architecture []byte
	var _cpu_variant []byte
	var _cpu_part []uint16
	var _cpu_revision []byte
	var info cpuinfo
	for {
		line, err := rd.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.Trim(line, "\n")
		fields := strings.Split(line, ":")
		if len(fields) < 2 {
			continue
		}
		key := strings.TrimSpace(fields[0])
		val := strings.TrimSpace(fields[1])
		switch strings.ToLower(key) {
		default:

		case "processor":
			u, err := strconv.ParseUint(val, 10, 64)
			if err != nil {
				return &info, err
			}
			_processors = append(_processors, uint(u))
		case "model name":
			_models = append(_models, val)
		case "bogomips":
			f, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return &info, err
			}
			_bogomips = append(_bogomips, f)
		case "features":
			_features = append(_features, strings.Split(val, " "))
		case "cpu implementer":
			var v uint64
			if strings.HasPrefix(val, "0x") {
				v, err = strconv.ParseUint(val[2:], 16, 8)
			} else {
				v, err = strconv.ParseUint(val, 10, 8)
			}
			if err != nil {
				return &info, err
			}
			_cpu_implementer = append(_cpu_implementer, byte(v))
		case "cpu architecture":
			var v uint64
			if strings.HasPrefix(val, "0x") {
				v, err = strconv.ParseUint(val[2:], 16, 8)
			} else {
				v, err = strconv.ParseUint(val, 10, 8)
			}
			if err != nil {
				return &info, err
			}
			_cpu_architecture = append(_cpu_architecture, byte(v))
		case "cpu variant":
			var v uint64
			if strings.HasPrefix(val, "0x") {
				v, err = strconv.ParseUint(val[2:], 16, 8)
			} else {
				v, err = strconv.ParseUint(val, 10, 8)
			}
			if err != nil {
				return &info, err
			}
			_cpu_variant = append(_cpu_variant, byte(v))
		case "cpu part":
			var v uint64
			if strings.HasPrefix(val, "0x") {
				v, err = strconv.ParseUint(val[2:], 16, 16)
			} else {
				v, err = strconv.ParseUint(val, 10, 8)
			}
			if err != nil {
				return &info, err
			}
			_cpu_part = append(_cpu_part, uint16(v))
		case "cpu revision":
			var v uint64
			if strings.HasPrefix(val, "0x") {
				v, err = strconv.ParseUint(val[2:], 16, 8)
			} else {
				v, err = strconv.ParseUint(val, 10, 8)
			}
			if err != nil {
				return &info, err
			}
			_cpu_revision = append(_cpu_revision, byte(v))
		case "hardware":
			info.Hardware = val
		case "revision":
			var v uint64
			if v, err = strconv.ParseUint(val, 16, 64); err != nil {
				return &info, err
			}
			info.Revision = v
		case "serial":
			l := len(val) / 2
			info.Serial = make([]byte, l)
			var v uint64
			for i := 0; i < l; i++ {
				if v, err = strconv.ParseUint(val[i*2:i*2+2], 16, 8); err != nil {
					return &info, err
				}
				info.Serial[i] = byte(v)
			}
		}
	}
	count := len(_processors)
	info.Cores = make([]coreinfo, count)
	for i := 0; i < count; i++ {
		info.Cores[i].Processor = _processors[i]
		info.Cores[i].ModelName = _models[i]
		info.Cores[i].BogoMIPS = _bogomips[i]
		info.Cores[i].Features = _features[i]
		info.Cores[i].CPU.Implementer = _cpu_implementer[i]
		info.Cores[i].CPU.Architecture = _cpu_architecture[i]
		info.Cores[i].CPU.Variant = _cpu_variant[i]
		info.Cores[i].CPU.Part = _cpu_part[i]
		info.Cores[i].CPU.Revision = _cpu_revision[i]
	}
	return &info, nil
}

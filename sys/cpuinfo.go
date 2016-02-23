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

var (
	cpuInfo *cpuinfo
)

func CPUInfo() (*cpuinfo, error) {
	if cpuInfo != nil {
		return cpuInfo, nil
	}
	var (
		file *os.File
		err  error
	)
	if file, err = os.Open("/proc/cpuinfo"); err != nil {
		return nil, err
	}
	cpuInfo = &cpuinfo{}
	if err = decodeCpuInfo(file, cpuInfo); err != nil {
		cpuInfo = nil
	}
	file.Close()
	return cpuInfo, err
}

func decodeCpuInfo(r io.Reader, info *cpuinfo) error {
	rd := bufio.NewReader(r)
	var (
		processors       []uint
		modelNames       []string
		bogomips         []float64
		features         [][]string
		cpuImplementers  []byte
		cpuArchitectures []byte
		cpuVariants      []byte
		cpuParts         []uint16
		cpuRevisions     []byte
	)
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
				return err
			}
			processors = append(processors, uint(u))
		case "model name":
			modelNames = append(modelNames, val)
		case "bogomips":
			f, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return err
			}
			bogomips = append(bogomips, f)
		case "features":
			features = append(features, strings.Split(val, " "))
		case "cpu implementer":
			var v uint64
			if strings.HasPrefix(val, "0x") {
				v, err = strconv.ParseUint(val[2:], 16, 8)
			} else {
				v, err = strconv.ParseUint(val, 10, 8)
			}
			if err != nil {
				return err
			}
			cpuImplementers = append(cpuImplementers, byte(v))
		case "cpu architecture":
			var v uint64
			if strings.HasPrefix(val, "0x") {
				v, err = strconv.ParseUint(val[2:], 16, 8)
			} else {
				v, err = strconv.ParseUint(val, 10, 8)
			}
			if err != nil {
				return err
			}
			cpuArchitectures = append(cpuArchitectures, byte(v))
		case "cpu variant":
			var v uint64
			if strings.HasPrefix(val, "0x") {
				v, err = strconv.ParseUint(val[2:], 16, 8)
			} else {
				v, err = strconv.ParseUint(val, 10, 8)
			}
			if err != nil {
				return err
			}
			cpuVariants = append(cpuVariants, byte(v))
		case "cpu part":
			var v uint64
			if strings.HasPrefix(val, "0x") {
				v, err = strconv.ParseUint(val[2:], 16, 16)
			} else {
				v, err = strconv.ParseUint(val, 10, 8)
			}
			if err != nil {
				return err
			}
			cpuParts = append(cpuParts, uint16(v))
		case "cpu revision":
			var v uint64
			if strings.HasPrefix(val, "0x") {
				v, err = strconv.ParseUint(val[2:], 16, 8)
			} else {
				v, err = strconv.ParseUint(val, 10, 8)
			}
			if err != nil {
				return err
			}
			cpuRevisions = append(cpuRevisions, byte(v))
		case "hardware":
			info.Hardware = val
		case "revision":
			var v uint64
			if v, err = strconv.ParseUint(val, 16, 64); err != nil {
				return err
			}
			info.Revision = v
		case "serial":
			l := len(val) / 2
			info.Serial = make([]byte, l)
			var v uint64
			for i := 0; i < l; i++ {
				if v, err = strconv.ParseUint(val[i*2:i*2+2], 16, 8); err != nil {
					return err
				}
				info.Serial[i] = byte(v)
			}
		}
	}
	count := len(processors)
	info.Cores = make([]coreinfo, count)
	for i := 0; i < count; i++ {
		info.Cores[i].Processor = processors[i]
		info.Cores[i].ModelName = modelNames[i]
		info.Cores[i].BogoMIPS = bogomips[i]
		info.Cores[i].Features = features[i]
		info.Cores[i].CPU.Implementer = cpuImplementers[i]
		info.Cores[i].CPU.Architecture = cpuArchitectures[i]
		info.Cores[i].CPU.Variant = cpuVariants[i]
		info.Cores[i].CPU.Part = cpuParts[i]
		info.Cores[i].CPU.Revision = cpuRevisions[i]
	}
	return nil
}

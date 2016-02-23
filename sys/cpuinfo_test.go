package sys

import (
	"bytes"
	"strings"
	"testing"
)

func TestDecode(t *testing.T) {
	r := strings.NewReader(cpuInfoText)
	if info, err := decodeCpuInfo(r); err != nil {
		t.Error(err)
	} else {
		if len(info.Cores) != 4 {
			t.Errorf("Core number mismatch: %d", len(info.Cores))
		} else {
			for n := range info.Cores {
				if info.Cores[n].Processor != uint(n) {
					t.Errorf("processor number mismatch: %s", info.Cores[n].Processor)
				}
				if info.Cores[n].ModelName != "ARMv7 Processor rev 5 (v7l)" {
					t.Errorf("Model name mismatch: %s", info.Cores[n].ModelName)
				}
				if info.Cores[n].BogoMIPS != 38.40 {
					t.Errorf("BogoMIPS mismatch: %d", info.Cores[n].BogoMIPS)
				}
				if info.Cores[n].CPU.Implementer != 0x41 {
					t.Errorf("CPU implementer mismatch: %d", info.Cores[n].CPU.Implementer)
				}
				if info.Cores[n].CPU.Architecture != 7 {
					t.Errorf("CPU architecture mismatch: %d", info.Cores[n].CPU.Architecture)
				}
				if info.Cores[n].CPU.Variant != 0 {
					t.Errorf("CPU variant mismatch: %d", info.Cores[n].CPU.Variant)
				}
				if info.Cores[n].CPU.Part != 0xc07 {
					t.Errorf("CPU part mismatch: %d", info.Cores[n].CPU.Part)
				}
				if info.Cores[n].CPU.Revision != 5 {
					t.Errorf("CPU revision mismatch: %d", info.Cores[n].CPU.Revision)
				}
			}
		}
		if strings.Compare(info.Hardware, "BCM2709") != 0 {
			t.Errorf("Hardware mismatch: %s", info.Hardware)
		}
		if info.Revision != 0xa21041 {
			t.Errorf("Revision mismatch: %d", info.Revision)
		}
		if bytes.Compare(info.Serial, []byte{0x00, 0x00, 0x00, 0x00, 0x1f, 0xf1, 0x50, 0xd5}) != 0 {
			t.Errorf("Serial mismatch: %x", info.Serial)
		}
	}
}

const cpuInfoText = `processor : 0
model name  : ARMv7 Processor rev 5 (v7l)
BogoMIPS  : 38.40
Features  : half thumb fastmult vfp edsp neon vfpv3 tls vfpv4 idiva idivt vfpd32 lpae evtstrm
CPU implementer : 0x41
CPU architecture: 7
CPU variant : 0x0
CPU part  : 0xc07
CPU revision  : 5

processor : 1
model name  : ARMv7 Processor rev 5 (v7l)
BogoMIPS  : 38.40
Features  : half thumb fastmult vfp edsp neon vfpv3 tls vfpv4 idiva idivt vfpd32 lpae evtstrm
CPU implementer : 0x41
CPU architecture: 7
CPU variant : 0x0
CPU part  : 0xc07
CPU revision  : 5

processor : 2
model name  : ARMv7 Processor rev 5 (v7l)
BogoMIPS  : 38.40
Features  : half thumb fastmult vfp edsp neon vfpv3 tls vfpv4 idiva idivt vfpd32 lpae evtstrm
CPU implementer : 0x41
CPU architecture: 7
CPU variant : 0x0
CPU part  : 0xc07
CPU revision  : 5

processor : 3
model name  : ARMv7 Processor rev 5 (v7l)
BogoMIPS  : 38.40
Features  : half thumb fastmult vfp edsp neon vfpv3 tls vfpv4 idiva idivt vfpd32 lpae evtstrm
CPU implementer : 0x41
CPU architecture: 7
CPU variant : 0x0
CPU part  : 0xc07
CPU revision  : 5

Hardware  : BCM2709
Revision  : a21041
Serial    : 000000001ff150d5
`

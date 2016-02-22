package sys

import (
	"testing"
)

func TestParse(t *testing.T) {
	return
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
Serial    : 000000001ff150d5`

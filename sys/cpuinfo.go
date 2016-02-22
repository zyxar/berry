package sys

type CPUInfo struct {
	cores []struct {
		Processor uint
		Model     string
		BogoMIPS  float64
		Features  []string
		CPU       struct {
			Implementer  byte
			Architecture byte
			Variant      byte
			Part         byte
			Revision     byte
		}
	}
	Hardware string
	Revision string
	Serial   []byte
}

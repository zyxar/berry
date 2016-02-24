package rc522

type RfidReader interface {
	FindTag() ([]byte, error)
	ReadTag(uint8) ([]byte, error)
	SelectTag()
}

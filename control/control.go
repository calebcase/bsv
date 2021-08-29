package control

// Type is the control block type.
type Type = byte

// Control Block Types
var (
	Invalid Type = 0b0000_0000
	Data    Type = 0b0000_0001
	Size    Type = 0b0000_0010
	Size2   Type = 0b0000_0100
	Skip    Type = 0b0000_1000
)

// Control Block Masks
var (
	dataMask  byte = 0b0000_0001
	sizeMask  byte = 0b0000_0011
	size2Mask byte = 0b0000_0111
	skipMask  byte = 0b0000_1111
)

// Parse returns the control block type and value.
func Parse(b byte) (t Type, value uint8, err error) {
	if b&dataMask == Data {
		v := b & ^dataMask >> 1

		return Data, v, nil
	} else if b&sizeMask == Size {
		v := b & ^sizeMask >> 2

		return Size, v, nil
	} else if b&size2Mask == Size2 {
		v := b & ^size2Mask >> 3

		return Size2, v, nil
	} else if b&skipMask == Skip {
		v := b & ^skipMask >> 4

		return Skip, v, nil
	}

	return Invalid, 0, Error.New("invalid control byte: %08b", b)
}

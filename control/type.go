package control

type Type struct {
	Prefix byte
	Mask   byte
	Abbr   string
}

// Match returns true if this control type matches the given byte.
func (t Type) Match(b byte) bool {
	return b&^t.Mask == t.Prefix
}

type types []Type

func (ts types) Match(b byte) (t Type, ok bool) {
	for _, t := range ts {
		if t.Match(b) {
			return t, true
		}
	}

	return t, false
}

var (
	Unknown            = Type{}
	Data               = Type{0b_1000_0000, 0b_0111_1111, "d"}
	DataSize           = Type{0b_0100_0000, 0b_0011_1111, "dz"}
	Data1              = Type{0b_0010_0000, 0b_0001_1111, "d1"}
	Data2              = Type{0b_0001_0000, 0b_0000_1111, "d2"}
	DataSizeSize       = Type{0b_0000_1000, 0b_0000_0111, "dzz"}
	ContainerSymmetric = Type{0b_0000_0111, 0b_0000_0000, "cs"}
	ContainerBounded   = Type{0b_0000_0101, 0b_0000_0000, "cb"}
	ContainerUnbounded = Type{0b_0000_0110, 0b_0000_0000, "cu"}
	ContainerEnd       = Type{0b_0000_0100, 0b_0000_0000, "ce"}
	SkipSize           = Type{0b_0000_0010, 0b_0000_0001, "sz"}
	Empty              = Type{0b_0000_0001, 0b_0000_0000, "e"}
	Null               = Type{0b_0000_0000, 0b_0000_0000, "n"}

	Types = types{
		Data,
		DataSize,
		Data1,
		Data2,
		DataSizeSize,
		ContainerSymmetric,
		ContainerBounded,
		ContainerUnbounded,
		ContainerEnd,
		SkipSize,
		Empty,
		Null,
	}
)

package control

import (
	"encoding/binary"
	"io"
)

// Type is the control block type.
type Type = byte

// Control Block Types
var (
	Invalid      Type = 0b0000_0000
	Data         Type = 0b1000_0000
	DataSize     Type = 0b0100_0000
	Data1        Type = 0b0010_0000
	Data2        Type = 0b0001_0000
	Skip         Type = 0b0000_1000
	DataSizeSize Type = 0b0000_0100
	SkipSizeSize Type = 0b0000_0010
	Null         Type = 0b0000_0001
)

// Control Block Masks
var (
	dataMask         byte = 0b1000_0000
	dataSizeMask     byte = 0b1100_0000
	data1Mask        byte = 0b1110_0000
	data2Mask        byte = 0b1111_0000
	skipMask         byte = 0b1111_1000
	dataSizeSizeMask byte = 0b1111_1100
	skipSizeSizeMask byte = 0b1111_1110
)

// TypeString returns the string name for the type.
func TypeString(t Type) string {
	switch t {
	case Invalid:
		return "Invalid"
	case Data:
		return "Data"
	case DataSize:
		return "Data Size"
	case Data1:
		return "Data + 1"
	case Data2:
		return "Data + 2"
	case Skip:
		return "Skip"
	case DataSizeSize:
		return "Data Size Size"
	case SkipSizeSize:
		return "Skip Size Size"
	case Null:
		return "Null"
	}

	return "Impossible"
}

// New returns a configured control byte.
func New(t Type, value uint8) (b byte, err error) {
	switch t {
	case Invalid:
		err = Error.New("refusing to create invalid block")

		return
	case Data:
		if value > 127 {
			err = Error.New("value too large: %d", value)

			return
		}

		b = (^dataMask & value) | Data
	case DataSize:
		if value > 63 {
			err = Error.New("value too large: %d", value)

			return
		}

		value = value - 1
		b = (^dataSizeMask & value) | DataSize
	case Data1:
		if value > 31 {
			err = Error.New("value too large: %d", value)

			return
		}

		b = (^data1Mask & value) | Data1
	case Data2:
		if value > 15 {
			err = Error.New("value too large: %d", value)

			return
		}

		b = (^data2Mask & value) | Data2
	case Skip:
		if value > 8 {
			err = Error.New("value too large: %d", value)

			return
		}

		value = value - 1
		b = (^skipMask & value) | Skip
	case DataSizeSize:
		if value > 4 {
			err = Error.New("value too large: %d", value)

			return
		}

		value = value - 1
		b = (^dataSizeSizeMask & value) | DataSizeSize
	case SkipSizeSize:
		if value > 2 {
			err = Error.New("value too large: %d", value)

			return
		}

		value = value - 1
		b = (^skipSizeSizeMask & value) | SkipSizeSize
	case Null:
		b = Null
	}

	return
}

// Parse returns the control block type and value.
func Parse(b byte) (t Type, value uint8, err error) {
	if b&dataMask == Data {
		v := b & ^dataMask

		return Data, v, nil
	} else if b&dataSizeMask == DataSize {
		v := b & ^dataSizeMask

		return DataSize, v + 1, nil
	} else if b&data1Mask == Data1 {
		v := b & ^data1Mask

		return Data1, v, nil
	} else if b&data2Mask == Data2 {
		v := b & ^data2Mask

		return Data2, v, nil
	} else if b&skipMask == Skip {
		v := b & ^skipMask

		return Skip, v + 1, nil
	} else if b&dataSizeSizeMask == DataSizeSize {
		v := b & ^dataSizeSizeMask

		return DataSizeSize, v + 1, nil
	} else if b&skipSizeSizeMask == SkipSizeSize {
		v := b & ^skipSizeSizeMask

		return SkipSizeSize, v + 1, nil
	} else if b == Null {
		return Null, 0, nil
	}

	return Invalid, 0, Error.New("invalid control byte: %08b", b)
}

// Block is a block.
type Block struct {
	Type Type
	Data []byte
	Size uint64
}

func (b *Block) resize(size uint64) {
	if uint64(cap(b.Data)) < size {
		b.Data = make([]byte, size)
	} else {
		b.Data = b.Data[:size]
	}
}

// Decoder is a decoder.
type Decoder struct {
	r io.Reader
}

// Decode parses a block from the reader.
func (d *Decoder) Decode(b *Block) (err error) {
	cb := make([]byte, 1)

	_, err = io.ReadFull(d.r, cb)
	if err != nil {
		return
	}

	t, v, err := Parse(cb[0])
	if err != nil {
		return
	}

	b.Type = t
	b.Data = b.Data[:0]
	b.Size = 0

	switch t {
	case Invalid:
		panic("invalid block detected when an error should have prevented it")
	case Data:
		b.Data = append(b.Data, byte(v))
	case DataSize:
		b.resize(uint64(v))

		_, err = io.ReadFull(d.r, b.Data)
		if err != nil {
			return
		}
	case Data1:
		b.resize(2)

		b.Data[0] = v
		_, err = io.ReadFull(d.r, b.Data[1:])
		if err != nil {
			return
		}
	case Data2:
		b.resize(3)

		b.Data[0] = v
		_, err = io.ReadFull(d.r, b.Data[1:])
		if err != nil {
			return
		}
	case Skip:
		b.Size = uint64(v)
	case DataSizeSize:
		b.resize(uint64(v))

		_, err = io.ReadFull(d.r, b.Data)
		if err != nil {
			return
		}

		b.Size = binary.LittleEndian.Uint64(b.Data)

		b.resize(b.Size)

		_, err = io.ReadFull(d.r, b.Data)
		if err != nil {
			return
		}
	case SkipSizeSize:
		b.resize(uint64(v))

		_, err = io.ReadFull(d.r, b.Data)
		if err != nil {
			return
		}

		b.Size = binary.LittleEndian.Uint64(b.Data)
	case Null:
		// Nothing to do in this case. The block is already in the
		// right condition.
	}

	return
}

// Encoder is an encoder.
type Encoder struct {
	w io.Writer
}

// NewEncoder returns a new encoder.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w: w,
	}
}

// Encode write a block to the writer.
func (e *Encoder) Encode(b *Block) (err error) {
	switch b.Type {
	case Invalid:
		e.w.Write([]byte{Invalid})
	case Data:
		cb, err := New(b.Type, b.Data[0])
		if err != nil {
			return err
		}

		_, err = e.w.Write([]byte{cb})
		if err != nil {
			return err
		}
	case DataSize:
		cb, err := New(b.Type, uint8(len(b.Data)))
		if err != nil {
			return err
		}

		_, err = e.w.Write([]byte{cb})
		if err != nil {
			return err
		}

		_, err = e.w.Write(b.Data)
		if err != nil {
			return err
		}
	case Data1:
		cb, err := New(b.Type, b.Data[0])
		if err != nil {
			return err
		}

		_, err = e.w.Write([]byte{cb})
		if err != nil {
			return err
		}

		_, err = e.w.Write(b.Data[1:])
		if err != nil {
			return err
		}
	case Data2:
		cb, err := New(b.Type, b.Data[0])
		if err != nil {
			return err
		}

		_, err = e.w.Write([]byte{cb})
		if err != nil {
			return err
		}

		_, err = e.w.Write(b.Data[1:])
		if err != nil {
			return err
		}
	case Skip:
		cb, err := New(b.Type, uint8(b.Size))
		if err != nil {
			return err
		}

		_, err = e.w.Write([]byte{cb})
		if err != nil {
			return err
		}
	case DataSizeSize:
		var bytes uint8
		buf := make([]byte, 4)
		size := len(b.Data)

		if size < 1<<8 {
			bytes = 1
			buf[0] = uint8(size)
			buf = buf[:1]
		} else if size < 1<<16 {
			bytes = 2
			binary.LittleEndian.PutUint16(buf[:2], uint16(size))
			buf = buf[:2]
		} else if size < 1<<24 {
			bytes = 3
			binary.LittleEndian.PutUint32(buf[:4], uint32(size))
			buf = buf[:3]
		} else {
			bytes = 4
			binary.LittleEndian.PutUint32(buf[:4], uint32(size))
		}

		cb, err := New(b.Type, bytes)
		if err != nil {
			return err
		}

		_, err = e.w.Write([]byte{cb})
		if err != nil {
			return err
		}

		_, err = e.w.Write(buf)
		if err != nil {
			return err
		}

		_, err = e.w.Write(b.Data)
		if err != nil {
			return err
		}
	case SkipSizeSize:
		var bytes uint8
		buf := make([]byte, 2)

		if b.Size <= 1<<8 {
			bytes = 1
			buf[0] = uint8(b.Size - 1)
			buf = buf[:1]
		} else {
			bytes = 2
			binary.LittleEndian.PutUint16(buf[:2], uint16(b.Size-1))
			buf = buf[:2]
		}

		cb, err := New(b.Type, bytes)
		if err != nil {
			return err
		}

		_, err = e.w.Write([]byte{cb})
		if err != nil {
			return err
		}

		_, err = e.w.Write(buf)
		if err != nil {
			return err
		}
	case Null:
		_, err = e.w.Write([]byte{Null})
		if err != nil {
			return err
		}
	}

	return nil
}

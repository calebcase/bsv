package integer

import (
	"math/big"

	"github.com/calebcase/bsv/control"
)

// Block is a signed integer number.
type Block struct {
	Value    []byte
	Negative bool
}

// MarshalBinary implements encoding.BinaryMarshaler.
func (b Block) MarshalBinary() (data []byte, err error) {
	i := new(big.Int).SetBytes(b.Value)

	i.Lsh(i, 1)
	if b.Negative {
		i.SetBit(i, 0, 1)
	}

	data = i.Bytes()

	// Note: big.Int encodes zero as an empty byte array, but we
	// desire zero to be an actual zero byte.
	if data == nil || len(data) == 0 {
		data = []byte{0}
	}

	return data, nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler.
func (b *Block) UnmarshalBinary(data []byte) (err error) {
	i := new(big.Int).SetBytes(data)

	b.Negative = i.Bit(0) == 1
	i.Rsh(i, 1)

	data = i.Bytes()

	// Note: big.Int encodes zero as an empty byte array, but we
	// desire zero to be an actual zero byte.
	if data == nil || len(data) == 0 {
		data = []byte{0}
	}

	b.Value = data

	return nil
}

// Schema for an integer.
type Schema struct {
	Bits uint64

	Signed bool

	Nullable    bool
	Key         bool
	ContentType string
}

// Decoder is a decoder.
type Decoder struct {
	schema Schema
	cd     *control.Decoder
}

// NewDecoder returns a new decoder.
func NewDecoder(schema Schema, cd *control.Decoder) *Decoder {
	return &Decoder{
		schema: schema,
		cd:     cd,
	}
}

// Decode parses a block from the reader.
func (d *Decoder) Decode(b *Block) (err error) {
	defer Error.WrapP(&err)

	cb := &control.Block{}

	err = d.cd.Decode(cb)
	if err != nil {
		return err
	}

	if d.schema.Signed {
		// TODO: Use schema information to optimize this choice (e.g.
		// don't use big.Int if the value is small enough to be
		// directly encoded to a fixed int format like uint64).
		i := new(big.Int).SetBytes(cb.Data)
		b.Negative = i.Bit(0) == 1
		i.Rsh(i, 1)
		b.Value = i.Bytes()
	} else {
		b.Value = cb.Data
	}

	return nil
}

// Encoder is an encoder.
type Encoder struct {
	schema Schema
	ce     *control.Encoder
}

// NewEncoder returns a new encoder.
func NewEncoder(schema Schema, ce *control.Encoder) *Encoder {
	return &Encoder{
		schema: schema,
		ce:     ce,
	}
}

// Encode write a block to the writer.
func (e *Encoder) Encode(b *Block) (err error) {
	defer Error.WrapP(&err)

	var bits int
	var bytes []byte

	if b.Value != nil {
		i := new(big.Int).SetBytes(b.Value)
		if e.schema.Signed {
			i.Lsh(i, 1)
			if b.Negative {
				i.SetBit(i, 0, 1)
			}
		}

		bits = i.BitLen()
		bytes = i.Bytes()

		// Note: big.Int encodes zero as an empty byte array, but we
		// desire zero to be an actual zero byte.
		if bytes == nil || len(bytes) == 0 {
			bytes = []byte{0}
		}
	}

	switch {
	case bytes == nil:
		return e.ce.Encode(&control.Block{
			Type: control.Null,
		})
	case bits <= 7:
		return e.ce.Encode(&control.Block{
			Type: control.Data,
			Data: bytes,
		})
	case bits <= 13: // 5+8
		return e.ce.Encode(&control.Block{
			Type: control.Data1,
			Data: bytes,
		})
	case bits <= 20: // 4+8+8
		return e.ce.Encode(&control.Block{
			Type: control.Data2,
			Data: bytes,
		})
	case len(bytes) < 64: // 2^6 - 1
		return e.ce.Encode(&control.Block{
			Type: control.DataSize,
			Data: bytes,
		})
	case len(bytes) <= 4294967296: // 2^(4 * 8)
		return e.ce.Encode(&control.Block{
			Type: control.DataSizeSize,
			Data: bytes,
		})
	}

	return Error.New("too large")
}

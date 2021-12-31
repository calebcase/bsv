package decimal

import (
	"github.com/calebcase/bsv/control"
	"github.com/calebcase/bsv/integer"
)

// Block is a fixed point base 10 decimal number.
type Block struct {
	Value     *integer.Block
	Scale     *integer.Block
	ScaleSize uint8
}

// Schema represents a configured number format.
type Schema struct {
	Scale uint32

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

	b.ScaleSize = cb.Data[len(cb.Data)-1] & 0b0000_0011

	data := cb.Data[:len(cb.Data)-1]

	switch {
	case b.ScaleSize == 0b00:
		b.Value = &integer.Block{}
		err = b.Value.UnmarshalBinary(data)
		if err != nil {
			return err
		}
	case b.ScaleSize == 0b01:
		b.Scale = &integer.Block{}
		err = b.Scale.UnmarshalBinary([]byte{
			data[len(data)-1] >> 2,
		})
		if err != nil {
			return err
		}

		b.Value = &integer.Block{}
		err = b.Value.UnmarshalBinary(data[:len(data)-2])
		if err != nil {
			return err
		}
	case b.ScaleSize == 0b10:
		b.Scale = &integer.Block{}
		err = b.Scale.UnmarshalBinary([]byte{
			data[len(data)-2] >> 2,
			((data[len(data)-2] & 0b11) << 6) | (data[len(data)-1] >> 2),
		})
		if err != nil {
			return err
		}

		b.Value = &integer.Block{}
		err = b.Value.UnmarshalBinary(data[:len(data)-3])
		if err != nil {
			return err
		}
	case b.ScaleSize == 0b11:
		b.Scale = &integer.Block{}
		err = b.Scale.UnmarshalBinary([]byte{
			data[len(data)-3] >> 2,
			((data[len(data)-3] & 0b11) << 6) | (data[len(data)-2] >> 2),
			((data[len(data)-2] & 0b11) << 6) | (data[len(data)-1] >> 2),
		})
		if err != nil {
			return err
		}

		b.Value = &integer.Block{}
		err = b.Value.UnmarshalBinary(data[:len(data)-4])
		if err != nil {
			return err
		}
	}

	return nil
}

/*
// DecodeBigInt read an integer.
func (d *Decoder) DecodeBigInt(i *big.Int) (err error) {
	defer Error.WrapP(&err)

	b := &control.Block{}

	err = d.blk.Decode(b)
	if err != nil {
		return err
	}

	return nil
}
*/

// Encoder is an encoder.
type Encoder struct {
	schema Schema
	blk    *control.Encoder
}

// NewEncoder returns a new encoder.
func NewEncoder(schema Schema, blk *control.Encoder) *Encoder {
	return &Encoder{
		schema: schema,
		blk:    blk,
	}
}

/*
// EncodeBigInt writes the integer.
//
// NOTE: The provided integer will be modified in place for encoding.
func (e *Encoder) EncodeBigInt(i *big.Int) (err error) {
	defer Error.WrapP(&err)

	if e.schema.Signed {
		i.Lsh(i, 1)

		if i.Sign() < 0 {
			i.SetBit(i, 0, 1)
			i.Mul(i, negOne)
		}
	}

	bits := i.BitLen()
	bytes := i.Bytes()

	switch {
	case bits <= 7:
		return e.blk.Encode(&control.Block{
			Type: control.Data,
			Data: bytes,
		})
	case bits <= 13:
		return e.blk.Encode(&control.Block{
			Type: control.Data1,
			Data: bytes,
		})
	case bits <= 20:
		return e.blk.Encode(&control.Block{
			Type: control.Data2,
			Data: bytes,
		})
	case len(bytes) <= 64:
		return e.blk.Encode(&control.Block{
			Type: control.DataSize,
			Data: bytes,
			Size: uint64(len(bytes)),
		})
	case len(bytes) <= 4294967296:
		return e.blk.Encode(&control.Block{
			Type: control.DataSizeSize,
			Data: bytes,
			Size: uint64(len(bytes)),
		})
	}

	return Error.New("too large")
}
*/

package integer

import (
	"bytes"
	"fmt"
	"math/big"
	"testing"

	"github.com/calebcase/bsv/control"
	"github.com/stretchr/testify/require"
)

func TestMarshalUnmarshal(t *testing.T) {
	type TC struct {
		name string
		blk  *Block
		data []byte
	}

	tcs := []TC{
		{
			name: "+0",
			blk: &Block{
				Value: []byte{
					0b0000_0000,
				},
				Negative: false,
			},
			data: []byte{
				0b0000_0000,
			},
		},
		{
			name: "+1",
			blk: &Block{
				Value: []byte{
					0b0000_0001,
				},
				Negative: false,
			},
			data: []byte{
				0b0000_0010,
			},
		},
		{
			name: "-1",
			blk: &Block{
				Value: []byte{
					0b0000_0001,
				},
				Negative: true,
			},
			data: []byte{
				0b0000_0011,
			},
		},
		{
			name: "-127",
			blk: &Block{
				Value: []byte{
					0b0111_1111,
				},
				Negative: true,
			},
			data: []byte{
				0b1111_1111,
			},
		},
		{
			name: "+127",
			blk: &Block{
				Value: []byte{
					0b0111_1111,
				},
				Negative: false,
			},
			data: []byte{
				0b1111_1110,
			},
		},
		{
			name: "+32767",
			blk: &Block{
				Value: []byte{
					0b0111_1111,
					0b1111_1111,
				},
				Negative: false,
			},
			data: []byte{
				0b1111_1111,
				0b1111_1110,
			},
		},
	}

	for i, tc := range tcs {
		t.Run(fmt.Sprintf("[%d]%s", i, tc.name), func(t *testing.T) {
			t.Run("marshal", func(t *testing.T) {
				data, err := tc.blk.MarshalBinary()
				require.NoError(t, err)
				require.Equal(t, tc.data, data)
			})

			t.Run("unmarshal", func(t *testing.T) {
				blk := &Block{}
				err := blk.UnmarshalBinary(tc.data)
				require.NoError(t, err)
				require.Equal(t, tc.blk, blk)

				// These checks ensure that our test case name matches the value.
				i := new(big.Int)
				err = i.UnmarshalText([]byte(tc.name))
				require.NoError(t, err)

				bs := i.Bytes()
				if len(bs) == 0 {
					bs = []byte{0}
				}
				require.Equal(t, bs, blk.Value)

				if i.Sign() < 0 {
					require.True(t, blk.Negative)
				} else {
					require.False(t, blk.Negative)
				}
			})
		})
	}
}

func TestEncodeDecode(t *testing.T) {
	type TC struct {
		name   string
		schema Schema
		blk    *Block
		data   []byte
	}

	tcs := []TC{
		{
			name: "0",
			schema: Schema{
				Bits: 64,
			},
			blk: &Block{
				Value: []byte{
					0b0000_0000,
				},
				Negative: false,
			},
			data: []byte{
				0b1000_0000,
			},
		},
		{
			name: "1",
			schema: Schema{
				Bits: 64,
			},
			blk: &Block{
				Value: []byte{
					0b0000_0001,
				},
				Negative: false,
			},
			data: []byte{
				0b1000_0001,
			},
		},
		{
			name: "+1",
			schema: Schema{
				Bits:   64,
				Signed: true,
			},
			blk: &Block{
				Value: []byte{
					0b0000_0001,
				},
				Negative: false,
			},
			data: []byte{
				0b1000_0010,
			},
		},
		{
			name: "-1",
			schema: Schema{
				Bits:   64,
				Signed: true,
			},
			blk: &Block{
				Value: []byte{
					0b0000_0001,
				},
				Negative: true,
			},
			data: []byte{
				0b1000_0011,
			},
		},
		{
			name: "-63",
			schema: Schema{
				Bits:   64,
				Signed: true,
			},
			blk: &Block{
				Value: []byte{
					0b0011_1111,
				},
				Negative: true,
			},
			data: []byte{
				0b1111_1111,
			},
		},
		{
			name: "+63",
			schema: Schema{
				Bits:   64,
				Signed: true,
			},
			blk: &Block{
				Value: []byte{
					0b0011_1111,
				},
				Negative: false,
			},
			data: []byte{
				0b1111_1110,
			},
		},
		{
			name: "+4095",
			schema: Schema{
				Bits:   64,
				Signed: true,
			},
			blk: &Block{
				Value: []byte{
					0b0000_1111,
					0b1111_1111,
				},
				Negative: false,
			},
			data: []byte{
				0b0011_1111,
				0b1111_1110,
			},
		},
		{
			name: "-524287",
			schema: Schema{
				Bits:   64,
				Signed: true,
			},
			blk: &Block{
				Value: []byte{
					0b0000_0111,
					0b1111_1111,
					0b1111_1111,
				},
				Negative: true,
			},
			data: []byte{
				0b0001_1111,
				0b1111_1111,
				0b1111_1111,
			},
		},
		{
			name: "-26187124863169134960105517574620793217733136368344518315866330944769070371237396439066160738607233257207093473020480568073738052367083144426628220715007",
			schema: Schema{
				Bits:   504,
				Signed: true,
			},
			blk: &Block{
				Value: []byte{
					0b0111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
					0b1111_1111,
				},
				Negative: true,
			},
			data: []byte{
				0b0111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
				0b1111_1111,
			},
		},
	}

	for i, tc := range tcs {
		t.Run(fmt.Sprintf("[%d]%s", i, tc.name), func(t *testing.T) {
			buf := bytes.NewBuffer(nil)

			t.Run("encode", func(t *testing.T) {
				enc := NewEncoder(tc.schema, control.NewEncoder(buf))
				err := enc.Encode(tc.blk)
				require.NoError(t, err)
				require.Equal(t, tc.data, buf.Bytes())
			})

			t.Run("decode", func(t *testing.T) {
				dec := NewDecoder(tc.schema, control.NewDecoder(buf))
				blk := &Block{}
				err := dec.Decode(blk)
				require.NoError(t, err)
				require.Equal(t, tc.blk, blk)

				// These checks ensure that our test case name matches the value.
				i := new(big.Int)
				err = i.UnmarshalText([]byte(tc.name))
				require.NoError(t, err)

				bs := i.Bytes()
				if len(bs) == 0 {
					bs = []byte{0}
				}
				require.Equal(t, bs, blk.Value)

				if i.Sign() < 0 {
					require.True(t, blk.Negative)
				} else {
					require.False(t, blk.Negative)
				}
			})
		})
	}
}

func BenchmarkEncode(b *testing.B) {
	buf := bytes.NewBuffer(nil)
	ce := control.NewEncoder(buf)

	schema := Schema{
		Bits:   64,
		Signed: true,
	}
	enc := NewEncoder(schema, ce)

	blk := &Block{
		Value: []byte{
			0b0000_0111,
			0b1111_1111,
			0b1111_1111,
		},
		Negative: true,
	}

	for n := 0; n < b.N; n++ {
		err := enc.Encode(blk)
		if err != nil {
			b.Fatalf("%+v", err)
		}
	}
}

func BenchmarkDecode(b *testing.B) {
	data := []byte{
		0b0001_1111,
		0b1111_1111,
		0b1111_1111,
	}

	blk := Block{}

	for n := 0; n < b.N; n++ {
		buf := bytes.NewBuffer(data)
		cd := control.NewDecoder(buf)

		schema := Schema{
			Bits:   64,
			Signed: true,
		}
		dec := NewDecoder(schema, cd)

		err := dec.Decode(&blk)
		if err != nil {
			b.Fatalf("%+v", err)
		}
	}
}

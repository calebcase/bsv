package control

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestControl(t *testing.T) {
	type TC struct {
		b   byte
		t   Type
		v   uint8
		err bool
	}

	tcs := []TC{
		// Zero values:
		{
			b:   0b0000_0001,
			t:   Data,
			v:   0,
			err: false,
		},
		{
			b:   0b0000_0010,
			t:   DataSize,
			v:   1,
			err: false,
		},
		{
			b:   0b0000_0100,
			t:   Data1,
			v:   0,
			err: false,
		},
		{
			b:   0b0000_1000,
			t:   Data2,
			v:   0,
			err: false,
		},
		{
			b:   0b0001_0000,
			t:   Skip,
			v:   1,
			err: false,
		},
		{
			b:   0b0010_0000,
			t:   DataSizeSize,
			v:   1,
			err: false,
		},
		{
			b:   0b0100_0000,
			t:   SkipSizeSize,
			v:   1,
			err: false,
		},
		{
			b:   0b1000_0000,
			t:   Null,
			v:   0,
			err: false,
		},
		{
			b:   0b0000_0000,
			t:   Invalid,
			v:   0,
			err: true,
		},
		// Non-zero Values:
		{
			b:   0b1010_1011,
			t:   Data,
			v:   85,
			err: false,
		},
		{
			b:   0b1010_1010,
			t:   DataSize,
			v:   43,
			err: false,
		},
		{
			b:   0b1010_1100,
			t:   Data1,
			v:   21,
			err: false,
		},
		{
			b:   0b1010_1000,
			t:   Data2,
			v:   10,
			err: false,
		},
		{
			b:   0b1011_0000,
			t:   Skip,
			v:   6,
			err: false,
		},
		{
			b:   0b1010_0000,
			t:   DataSizeSize,
			v:   3,
			err: false,
		},
		{
			b:   0b1100_0000,
			t:   SkipSizeSize,
			v:   2,
			err: false,
		},
	}

	for _, tc := range tcs {
		t.Run(fmt.Sprintf("%08b", tc.b), func(t *testing.T) {
			t.Run("parse", func(t *testing.T) {
				typ, v, err := Parse(tc.b)
				require.Equal(t, tc.t, typ)
				require.Equal(t, tc.v, v)

				if tc.err {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
			})
			t.Run("new", func(t *testing.T) {
				b, err := New(tc.t, tc.v)
				require.Equal(t, tc.b, b)

				if tc.err {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
			})
		})
	}
}

func TestEncode(t *testing.T) {
	type TC struct {
		blk  *Block
		data []byte
	}

	tcs := []TC{
		{
			blk: &Block{
				Type: Data,
				Data: []byte{0b0000_0000}, // 0
			},
			data: []byte{0b0000_0001},
		},
		{
			blk: &Block{
				Type: Data,
				Data: []byte{0b0101_0101}, // 85
			},
			data: []byte{0b1010_1011},
		},
		{
			blk: &Block{
				Type: Data,
				Data: []byte{0b0111_1111}, // 127
			},
			data: []byte{0b1111_1111},
		},
		{
			blk: &Block{
				Type: DataSize,
				Data: []byte{}, // E.g. zero length string
			},
			data: []byte{0b0000_0010},
		},
		{
			blk: &Block{
				Type: DataSize,
				Data: []byte{'a', 'b', 'c', 'd'},
			},
			data: []byte{0b0001_0010, 'a', 'b', 'c', 'd'},
		},
		{
			blk: &Block{
				Type: Data1,
				Data: []byte{0x1f, 0xff}, // 8_191
			},
			data: []byte{0b1111_1100, 0xff},
		},
		{
			blk: &Block{
				Type: Data2,
				Data: []byte{0x0f, 0xff, 0xff}, // 1_048_575
			},
			data: []byte{0b1111_1000, 0xff, 0xff},
		},
		{
			blk: &Block{
				Type: Skip,
				Size: 8,
			},
			data: []byte{0b1111_0000},
		},
		{
			blk: &Block{
				Type: DataSizeSize,
				Data: []byte{'a', 'b', 'c', 'd'},
			},
			data: []byte{0b0010_0000, 0b0000_0100, 'a', 'b', 'c', 'd'},
		},
		{
			blk: &Block{
				Type: SkipSizeSize,
				Size: 1 << 16, // 65536
			},
			data: []byte{0b1100_0000, 0xff, 0xff},
		},
	}

	for _, tc := range tcs {
		t.Run(fmt.Sprintf("%s %08b", TypeString(tc.blk.Type), tc.data), func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			enc := NewEncoder(buf)
			err := enc.Encode(tc.blk)
			require.NoError(t, err)
			require.Equal(t, tc.data, buf.Bytes())
		})
	}
}

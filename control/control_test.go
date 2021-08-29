package control

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	type TC struct {
		b   byte
		t   Type
		v   uint8
		err bool
	}

	tcs := []TC{
		{
			b:   0b0000_0001,
			t:   Data,
			v:   0,
			err: false,
		},
		{
			b:   0b0000_0010,
			t:   Size,
			v:   0,
			err: false,
		},
		{
			b:   0b0000_0100,
			t:   Size2,
			v:   0,
			err: false,
		},
		{
			b:   0b0000_1000,
			t:   Skip,
			v:   0,
			err: false,
		},
		{
			b:   0b1010_1011,
			t:   Data,
			v:   85,
			err: false,
		},
		{
			b:   0b1010_1010,
			t:   Size,
			v:   42,
			err: false,
		},
		{
			b:   0b1010_1100,
			t:   Size2,
			v:   21,
			err: false,
		},
		{
			b:   0b1010_1000,
			t:   Skip,
			v:   10,
			err: false,
		},
		{
			b:   0b1010_0000,
			t:   Invalid,
			v:   0,
			err: true,
		},
	}

	for _, tc := range tcs {
		t.Run(fmt.Sprintf("%08b", tc.b), func(t *testing.T) {
			typ, v, err := Parse(tc.b)
			require.Equal(t, tc.t, typ)
			require.Equal(t, tc.v, v)

			if tc.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

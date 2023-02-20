package control_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"

	"github.com/calebcase/bsv/control"
	"github.com/calebcase/oops"
)

func TestDecoder(t *testing.T) {
	type TC struct {
		Input  []byte
		Types  []control.Type
		Data   []byte
		Amount uint64
		Mark   error
	}

	t.Run("read", func(t *testing.T) {
		tcs := []TC{
			{
				Input: []byte{0b_1000_0000},
				Types: []control.Type{control.Data},
				Data:  []byte{0b_0000_0000},
				Mark:  oops.New("unexpected"),
			},
			{
				Input: []byte{0b_0100_0000, 0b_0000_0000},
				Types: []control.Type{control.DataSize},
				Data:  []byte{0b_0000_0000},
				Mark:  oops.New("unexpected"),
			},
			{
				Input: []byte{0b_0010_0000, 0b_0000_0000},
				Types: []control.Type{control.Data1},
				Data:  []byte{0b_0000_0000, 0b_0000_0000},
				Mark:  oops.New("unexpected"),
			},
			{
				Input: []byte{0b_0001_0000, 0b_0000_0000, 0b_0000_0000},
				Types: []control.Type{control.Data2},
				Data:  []byte{0b_0000_0000, 0b_0000_0000, 0b_0000_0000},
				Mark:  oops.New("unexpected"),
			},
			{
				Input: []byte{0b_0000_1000, 0b_0000_0000, 0b_0000_0000},
				Types: []control.Type{control.DataSizeSize},
				Data:  []byte{0b_0000_0000},
				Mark:  oops.New("unexpected"),
			},
			{
				Input: []byte{0b_0000_0111, 0b_0010_0000, 0b_0000_0000, 0b_0010_0000, 0b_0000_0111},
				Types: []control.Type{
					control.ContainerSymmetric,
					control.Data1,
				},
				Data: []byte{0b_0000_0000, 0b_0000_0000},
				Mark: oops.New("unexpected"),
			},
			{
				Input: []byte{0b_0000_0101, 0b_1000_0000, 0b_1000_0000},
				Types: []control.Type{
					control.ContainerBounded,
					control.Data,
				},
				Data: []byte{0b_0000_0000},
				Mark: oops.New("unexpected"),
			},
			{
				Input: []byte{0b_0000_0110, 0b_1000_0000, 0b_0000_0100},
				Types: []control.Type{
					control.ContainerUnbounded,
					control.Data,
					control.ContainerEnd,
				},
				Data: []byte{0b_0000_0000},
				Mark: oops.New("unexpected"),
			},
			{
				Input: []byte{0b_0000_0010, 0b_0000_0000},
				Types: []control.Type{
					control.SkipSize,
				},
				Amount: 1,
				Mark:   oops.New("unexpected"),
			},
			{
				Input: []byte{0b_0000_0001},
				Types: []control.Type{control.Empty},
				Mark:  oops.New("unexpected"),
			},
			{
				Input: []byte{0b_0000_0000},
				Types: []control.Type{control.Null},
				Mark:  oops.New("unexpected"),
			},
		}

		for _, tc := range tcs {
			name := []string{}
			for _, field := range tc.Types {
				name = append(name, field.Abbr)
			}

			t.Run(strings.Join(name, ","), func(t *testing.T) {
				var err error

				d := control.NewDecoder(bytes.NewBuffer(tc.Input))

				types := []control.Type{}
				var data []byte
				var amount uint64

				for d.Next() {
					field := d.Type()
					types = append(types, field)

					t.Logf("Type: %s\n", field.Abbr)

					switch d.Type() {
					case control.Data,
						control.Data1,
						control.Data2,
						control.DataSize,
						control.DataSizeSize:

						tmp, err := d.Data()
						require.NoError(t, err, tc.Mark)

						t.Logf("Data: %0b\n", tmp)

						data = append(data, tmp...)
					case control.ContainerBounded,
						control.ContainerUnbounded,
						control.ContainerSymmetric:

						err = d.Enter()
						require.NoError(t, err, tc.Mark)
					case control.ContainerEnd:
					case control.SkipSize:
						amt, err := d.Amount()
						require.NoError(t, err, tc.Mark)

						t.Logf("Amount: %d\n", amt)
						amount += amt
					case control.Empty, control.Null:
					}

					t.Logf("Stack: %s\n", spew.Sdump(d.Stack()))
				}
				require.NoError(t, d.Err(), tc.Mark)

				require.Equal(t, tc.Types, types, tc.Mark)
				require.Equal(t, tc.Data, data, tc.Mark)
				require.Equal(t, tc.Amount, amount, tc.Mark)
				require.Equal(t, 0, d.Depth(), tc.Mark)
			})
		}
	})

	t.Run("next", func(t *testing.T) {
		tcs := []TC{
			{
				Input: []byte{0b_1000_0000},
				Types: []control.Type{control.Data},
				Mark:  oops.New("unexpected"),
			},
			{
				Input: []byte{0b_0100_0000, 0b_0000_0000},
				Types: []control.Type{control.DataSize},
				Mark:  oops.New("unexpected"),
			},
			{
				Input: []byte{0b_0010_0000, 0b_0000_0000},
				Types: []control.Type{control.Data1},
				Mark:  oops.New("unexpected"),
			},
			{
				Input: []byte{0b_0001_0000, 0b_0000_0000, 0b_0000_0000},
				Types: []control.Type{control.Data2},
				Mark:  oops.New("unexpected"),
			},
			{
				Input: []byte{0b_0000_1000, 0b_0000_0000, 0b_0000_0000},
				Types: []control.Type{control.DataSizeSize},
				Mark:  oops.New("unexpected"),
			},
			{
				Input: []byte{0b_0000_0111, 0b_0010_0000, 0b_0000_0000, 0b_0010_0000, 0b_0000_0111},
				Types: []control.Type{
					control.ContainerSymmetric,
				},
				Mark: oops.New("unexpected"),
			},
			{
				Input: []byte{0b_0000_0101, 0b_1000_0000, 0b_1000_0000},
				Types: []control.Type{
					control.ContainerBounded,
				},
				Mark: oops.New("unexpected"),
			},
			{
				Input: []byte{0b_0000_0110, 0b_1000_0000, 0b_0000_0100},
				Types: []control.Type{
					control.ContainerUnbounded,
				},
				Mark: oops.New("unexpected"),
			},
			{
				Input: []byte{0b_0000_0010, 0b_0000_0000},
				Types: []control.Type{
					control.SkipSize,
				},
				Mark: oops.New("unexpected"),
			},
			{
				Input: []byte{0b_0000_0001},
				Types: []control.Type{control.Empty},
				Mark:  oops.New("unexpected"),
			},
			{
				Input: []byte{0b_0000_0000},
				Types: []control.Type{control.Null},
				Mark:  oops.New("unexpected"),
			},

			// More nesting situations:
			{
				Input: []byte{
					0b_0000_0110, // cu
					0b_0000_0110, // cu
					0b_1000_0000, // d
					0b_0000_0100, // ce
					0b_0000_0100, // ce
				},
				Types: []control.Type{
					control.ContainerUnbounded,
				},
				Mark: oops.New("unexpected"),
			},
			{
				Input: []byte{
					0b_1000_0000, // d
					0b_0000_0110, // cu
					0b_1000_0000, // d
					0b_0000_0100, // ce
					0b_1000_0000, // d
				},
				Types: []control.Type{
					control.Data,
					control.ContainerUnbounded,
					control.Data,
				},
				Mark: oops.New("unexpected"),
			},
		}

		for _, tc := range tcs {
			name := []string{}
			for _, field := range tc.Types {
				name = append(name, field.Abbr)
			}

			t.Run(strings.Join(name, ","), func(t *testing.T) {
				d := control.NewDecoder(bytes.NewBuffer(tc.Input))

				types := []control.Type{}

				for d.Next() {
					field := d.Type()
					types = append(types, field)

					t.Logf("Type: %s\n", field.Abbr)
					t.Logf("Stack: %s\n", spew.Sdump(d.Stack()))
				}
				err := d.Err()
				require.NoError(t, err, tc.Mark)

				require.Equal(t, tc.Types, types, tc.Mark)
				require.Equal(t, 0, d.Depth(), tc.Mark)
			})
		}
	})
}

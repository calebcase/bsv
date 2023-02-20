package control_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/calebcase/bsv/control"
	"github.com/calebcase/oops"
)

func TestRoundtrip(t *testing.T) {
	t.Run("data", func(t *testing.T) {
		type TC struct {
			Input  []byte
			Output []byte
			Mark   error
		}

		tcs := []TC{
			{
				Input:  []byte{0b_0000_0000},
				Output: []byte{0b_1000_0000},
				Mark:   oops.New("unexpected"),
			},
			{
				Input:  []byte{0b_0100_0000},
				Output: []byte{0b_1100_0000},
				Mark:   oops.New("unexpected"),
			},
			{
				Input:  []byte{0b_1000_0000},
				Output: []byte{0b_0100_0000, 0b_1000_0000},
				Mark:   oops.New("unexpected"),
			},
			{
				Input:  []byte{0b_0000_0000, 0b_0000_0000},
				Output: []byte{0b_0010_0000, 0b_0000_0000},
				Mark:   oops.New("unexpected"),
			},
			{
				Input:  []byte{0b_0001_0000, 0b_0000_0000},
				Output: []byte{0b_0011_0000, 0b_0000_0000},
				Mark:   oops.New("unexpected"),
			},
			{
				Input:  []byte{0b_0010_0000, 0b_0000_0000},
				Output: []byte{0b_0100_0000, 0b_0010_0000, 0b_0000_0000},
				Mark:   oops.New("unexpected"),
			},
			{
				Input:  []byte{0b_0000_0000, 0b_0000_0000, 0b_0000_0000},
				Output: []byte{0b_0001_0000, 0b_0000_0000, 0b_0000_0000},
				Mark:   oops.New("unexpected"),
			},
			{
				Input:  []byte{0b_0000_1000, 0b_0000_0000, 0b_0000_0000},
				Output: []byte{0b_0001_1000, 0b_0000_0000, 0b_0000_0000},
				Mark:   oops.New("unexpected"),
			},
			{
				Input:  []byte{0b_0001_0000, 0b_0000_0000, 0b_0000_0000},
				Output: []byte{0b_0100_0010, 0b_0001_0000, 0b_0000_0000, 0b_0000_0000},
				Mark:   oops.New("unexpected"),
			},
			{
				Input:  make([]byte, 4),
				Output: append([]byte{0b_0100_0011}, make([]byte, 4)...),
				Mark:   oops.New("unexpected"),
			},
			{
				Input:  make([]byte, 64),
				Output: append([]byte{0b_0111_1111}, make([]byte, 64)...),
				Mark:   oops.New("unexpected"),
			},
			{
				Input:  make([]byte, 65),
				Output: append([]byte{0b_0000_1000, 0b_0100_0000}, make([]byte, 65)...),
				Mark:   oops.New("unexpected"),
			},
			{
				Input: make([]byte, 1024),
				Output: append(
					[]byte{0b_0000_1001, 0b_0000_0011, 0b_1111_1111},
					make([]byte, 1024)...,
				),
				Mark: oops.New("unexpected"),
			},
		}

		for i, tc := range tcs {
			t.Run(shortName(i, tc.Output), func(t *testing.T) {
				var err error

				output := &bytes.Buffer{}
				e := control.NewEncoder(output)

				err = e.Data(tc.Input)
				require.NoError(t, err, tc.Mark)

				d := control.NewDecoder(output)

				ok := d.Next()
				require.True(t, ok, tc.Mark)

				input, err := d.Data()
				require.NoError(t, err, tc.Mark)
				require.Equal(t, tc.Input, input, tc.Mark)

				ok = d.Next()
				require.False(t, ok, tc.Mark)

				err = d.Err()
				require.NoError(t, err, tc.Mark)
			})
		}
	})

	t.Run("bound", func(t *testing.T) {
		type TC struct {
			Input  []byte
			Output []byte
			Mark   error
		}

		tcs := []TC{
			{
				Input:  []byte{0b_1000_0000},
				Output: []byte{0b_0000_0101, 0b_1000_0000, 0b_1000_0000},
				Mark:   oops.New("unexpected"),
			},
			{
				Input:  []byte{0b_1000_0000, 0b_1111_1111},
				Output: []byte{0b_0000_0101, 0b_1000_0001, 0b_1000_0000, 0b_1111_1111},
				Mark:   oops.New("unexpected"),
			},
		}

		for i, tc := range tcs {
			t.Run(shortName(i, tc.Output), func(t *testing.T) {
				var (
					err error
					ok  bool
				)

				output := &bytes.Buffer{}
				e := control.NewEncoder(output)

				err = e.Bound(tc.Input)
				require.NoError(t, err, tc.Mark)

				t.Logf("output=%08b", output.Bytes())

				d := control.NewDecoder(output)

				ok = d.Next()
				require.NoError(t, d.Err(), tc.Mark)
				require.True(t, ok, tc.Mark)
				require.Equal(t, control.ContainerBounded, d.Type(), tc.Mark)

				input, err := d.BSV()
				require.NoError(t, err, tc.Mark)
				require.Equal(t, tc.Input, input, tc.Mark)

				ok = d.Next()
				require.NoError(t, d.Err(), tc.Mark)
				require.False(t, ok, tc.Mark)
			})
		}
	})

	t.Run("unbound", func(t *testing.T) {
		type TC struct {
			Input  []byte
			Output []byte
			Mark   error
		}

		tcs := []TC{
			{
				Input:  []byte{0b_0000_0000},
				Output: []byte{0b_0000_0110, 0b_1000_0000, 0b_0000_0100},
				Mark:   oops.New("unexpected"),
			},
		}

		for i, tc := range tcs {
			t.Run(shortName(i, tc.Output), func(t *testing.T) {
				var (
					err error
					ok  bool
				)

				output := &bytes.Buffer{}
				e := control.NewEncoder(output)

				err = e.Unbound(func(e control.Encoder) (err error) {
					return e.Data(tc.Input)
				})
				require.NoError(t, err, tc.Mark)

				d := control.NewDecoder(output)

				ok = d.Next()
				require.NoError(t, d.Err(), tc.Mark)
				require.True(t, ok, tc.Mark)
				require.Equal(t, control.ContainerUnbounded, d.Type(), tc.Mark)

				err = d.Enter()
				require.NoError(t, err, tc.Mark)

				ok = d.Next()
				require.NoError(t, d.Err(), tc.Mark)
				require.True(t, ok, tc.Mark)
				require.Equal(t, control.Data, d.Type(), tc.Mark)

				input, err := d.Data()
				require.NoError(t, err, tc.Mark)
				require.Equal(t, tc.Input, input, tc.Mark)

				ok = d.Next()
				require.NoError(t, d.Err(), tc.Mark)
				require.Equal(t, control.ContainerEnd, d.Type(), tc.Mark)

				ok = d.Next()
				require.NoError(t, d.Err(), tc.Mark)
				require.False(t, ok, tc.Mark)
			})
		}
	})

	t.Run("skip", func(t *testing.T) {
		type TC struct {
			Amount uint64
			Output []byte
			Mark   error
		}

		tcs := []TC{
			{
				Amount: 1,
				Output: []byte{0b_0000_0010, 0b_0000_0000},
				Mark:   oops.New("unexpected"),
			},
			{
				Amount: 256,
				Output: []byte{0b_0000_0010, 0b_1111_1111},
				Mark:   oops.New("unexpected"),
			},
			{
				Amount: 512,
				Output: []byte{0b_0000_0011, 0b_0000_0001, 0b_1111_1111},
				Mark:   oops.New("unexpected"),
			},
			{
				Amount: 65536,
				Output: []byte{0b_0000_0011, 0b_1111_1111, 0b_1111_1111},
				Mark:   oops.New("unexpected"),
			},
		}

		for i, tc := range tcs {
			t.Run(shortName(i, tc.Output), func(t *testing.T) {
				var (
					err error
					ok  bool
				)

				output := &bytes.Buffer{}
				e := control.NewEncoder(output)

				err = e.Skip(tc.Amount)
				require.NoError(t, err, tc.Mark)

				d := control.NewDecoder(output)

				ok = d.Next()
				require.NoError(t, d.Err(), tc.Mark)
				require.True(t, ok, tc.Mark)
				require.Equal(t, control.SkipSize, d.Type(), tc.Mark)

				amount, err := d.Amount()
				require.NoError(t, err, tc.Mark)
				require.Equal(t, tc.Amount, amount)
			})
		}
	})

	t.Run("symmetric", func(t *testing.T) {
		type TC struct {
			Fn     func(control.Encoder) error
			Input  []byte
			Output []byte
			Mark   error
		}

		tcs := []TC{
			// Data
			{
				Fn: func(e control.Encoder) (err error) {
					return e.Data([]byte{
						0b_0000_0000,
					})
				},
				Input: []byte{
					0b_0000_0000,
				},
				Output: []byte{
					0b_1000_0000,
				},
				Mark: oops.New("unexpected"),
			},
			{
				Fn: func(e control.Encoder) (err error) {
					return e.Data([]byte{
						0b_0000_0001,
						0b_0000_0000,
					})
				},
				Input: []byte{
					0b_0000_0001,
					0b_0000_0000,
				},
				Output: []byte{
					0b_0000_0111,
					0b_0010_0001,
					0b_0000_0000,
					0b_0010_0001,
					0b_0000_0111,
				},
				Mark: oops.New("unexpected"),
			},
			{
				Fn: func(e control.Encoder) (err error) {
					return e.Data([]byte{
						0b_0000_0001,
						0b_0000_0000,
						0b_0000_0000,
					})
				},
				Input: []byte{
					0b_0000_0001,
					0b_0000_0000,
					0b_0000_0000,
				},
				Output: []byte{
					0b_0000_0111,
					0b_0001_0001,
					0b_0000_0000,
					0b_0000_0000,
					0b_0001_0001,
					0b_0000_0111,
				},
				Mark: oops.New("unexpected"),
			},
			{
				Fn: func(e control.Encoder) (err error) {
					return e.Data([]byte{
						0b_0000_0001,
						0b_0000_0000,
						0b_0000_0000,
						0b_0000_0000,
					})
				},
				Input: []byte{
					0b_0000_0001,
					0b_0000_0000,
					0b_0000_0000,
					0b_0000_0000,
				},
				Output: []byte{
					0b_0000_0111,
					0b_0100_0011,
					0b_0000_0001,
					0b_0000_0000,
					0b_0000_0000,
					0b_0000_0000,
					0b_0100_0011,
					0b_0000_0111,
				},
				Mark: oops.New("unexpected"),
			},
			{
				Fn: func(e control.Encoder) (err error) {
					return e.Data(make([]byte, 65))
				},
				Input: make([]byte, 65),
				Output: append(
					append(
						[]byte{
							0b_0000_0111,
							0b_0000_1000,
							0b_0100_0000,
						},
						make([]byte, 65)...,
					),
					0b_0100_0000,
					0b_0000_1000,
					0b_0000_0111,
				),
				Mark: oops.New("unexpected"),
			},
			{
				Fn: func(e control.Encoder) (err error) {
					return e.Data(make([]byte, 1024))
				},
				Input: make([]byte, 1024),
				Output: append(
					append(
						[]byte{
							0b_0000_0111,
							0b_0000_1001,
							0b_0000_0011,
							0b_1111_1111,
						},
						make([]byte, 1024)...,
					),
					0b_0000_0011,
					0b_1111_1111,
					0b_0000_1001,
					0b_0000_0111,
				),
				Mark: oops.New("unexpected"),
			},
			// Bound
			{
				Fn: func(e control.Encoder) (err error) {
					return e.Bound([]byte{
						0b_1000_0000,
					})
				},
				Input: []byte{
					0b_1000_0000,
				},
				Output: []byte{
					0b_0000_0111,
					0b_0000_0101,
					0b_1000_0000,
					0b_1000_0000,
					0b_1000_0000,
					0b_0000_0101,
					0b_0000_0111,
				},
				Mark: oops.New("unexpected"),
			},
			{
				Fn: func(e control.Encoder) (err error) {
					return e.Bound([]byte{
						0b_1000_0000,
						0b_1111_1111,
					})
				},
				Input: []byte{
					0b_1000_0000,
					0b_1111_1111,
				},
				Output: []byte{
					0b_0000_0111,
					0b_0000_0101,
					0b_1000_0001,
					0b_1000_0000,
					0b_1111_1111,
					0b_1000_0001,
					0b_0000_0101,
					0b_0000_0111,
				},
				Mark: oops.New("unexpected"),
			},
			// Unbound
			{
				Fn: func(e control.Encoder) (err error) {
					return e.Unbound(func(e control.Encoder) (err error) {
						return e.Data([]byte{0b_0000_0000})
					})
				},
				Input: []byte{
					0b_0000_0000,
				},
				Output: []byte{
					0b_0000_0110,
					0b_1000_0000,
					0b_0000_0100,
				},
				Mark: oops.New("unexpected"),
			},
			// Skip
			{
				Fn: func(e control.Encoder) (err error) {
					return e.Skip(1)
				},
				Output: []byte{
					0b_0000_0111,
					0b_0000_0010,
					0b_0000_0000,
					0b_0000_0010,
					0b_0000_0111,
				},
				Mark: oops.New("unexpected"),
			},
			{
				Fn: func(e control.Encoder) (err error) {
					return e.Skip(256)
				},
				Output: []byte{
					0b_0000_0111,
					0b_0000_0010,
					0b_1111_1111,
					0b_0000_0010,
					0b_0000_0111,
				},
				Mark: oops.New("unexpected"),
			},
			{
				Fn: func(e control.Encoder) (err error) {
					return e.Skip(512)
				},
				Output: []byte{
					0b_0000_0111,
					0b_0000_0011,
					0b_0000_0001,
					0b_1111_1111,
					0b_0000_0011,
					0b_0000_0111,
				},
				Mark: oops.New("unexpected"),
			},
			{
				Fn: func(e control.Encoder) (err error) {
					return e.Skip(65536)
				},
				Output: []byte{
					0b_0000_0111,
					0b_0000_0011,
					0b_1111_1111,
					0b_1111_1111,
					0b_0000_0011,
					0b_0000_0111,
				},
				Mark: oops.New("unexpected"),
			},
			// Empty
			{
				Fn: func(e control.Encoder) (err error) {
					return e.Empty()
				},
				Input: []byte{
					0b_0000_0001,
				},
				Output: []byte{
					0b_0000_0001,
				},
				Mark: oops.New("unexpected"),
			},
			// Null
			{
				Fn: func(e control.Encoder) (err error) {
					return e.Null()
				},
				Input: []byte{
					0b_0000_0000,
				},
				Output: []byte{
					0b_0000_0000,
				},
				Mark: oops.New("unexpected"),
			},
		}

		for i, tc := range tcs {
			t.Run(shortName(i, tc.Output), func(t *testing.T) {
				var (
					err error
					ok  bool
				)

				output := &bytes.Buffer{}
				e := control.NewEncoder(output)

				err = e.Symmetric(tc.Fn)
				require.NoError(t, err, tc.Mark)

				t.Logf("output=%02b", output.Bytes())

				require.Equal(t, tc.Output, output.Bytes(), tc.Mark)

				d := control.NewDecoder(output)

				ok = d.Next()
				require.NoError(t, d.Err(), tc.Mark)
				require.True(t, ok, tc.Mark)

				switch d.Type() {
				case control.Data:
					input, err := d.Data()
					require.NoError(t, err, tc.Mark)
					require.Equal(t, tc.Input, input)
				case control.ContainerSymmetric:
					// Enter the symmetric block.
					err = d.Enter()
					require.NoError(t, err, tc.Mark)

					// Get first (and only) embedded field.
					ok = d.Next()
					require.NoError(t, d.Err(), tc.Mark)
					require.True(t, ok, tc.Mark)

					t.Logf("type=%s", d.Type().Abbr)

					// This will be false because:
					//
					// * We only have one symmetric field in the test
					// * Only one field may be embedded in a symmetric block
					ok = d.Next()
					require.NoError(t, d.Err(), tc.Mark)
					require.False(t, ok, tc.Mark)
				case control.ContainerUnbounded:
					ok = d.Next()
					require.NoError(t, d.Err(), tc.Mark)
					require.False(t, ok, tc.Mark)
				case control.Empty, control.Null:
					require.Equal(t, tc.Input, []byte{d.Type().Prefix}, tc.Mark)
				default:
					require.Fail(t, "unexpected control block", "%s %+v", d.Type(), tc.Mark)
				}

				require.Equal(t, 0, d.Depth(), tc.Mark)
				require.Equal(t, len(tc.Output), int(d.Consumed()), tc.Mark)
			})
		}
	})

	t.Run("empty", func(t *testing.T) {
		var (
			err error
			ok  bool
		)

		output := &bytes.Buffer{}
		e := control.NewEncoder(output)

		err = e.Empty()
		require.NoError(t, err)

		d := control.NewDecoder(output)

		ok = d.Next()
		require.NoError(t, d.Err())
		require.True(t, ok)
		require.Equal(t, control.Empty, d.Type())
	})

	t.Run("null", func(t *testing.T) {
		var (
			err error
			ok  bool
		)

		output := &bytes.Buffer{}
		e := control.NewEncoder(output)

		err = e.Null()
		require.NoError(t, err)

		d := control.NewDecoder(output)

		ok = d.Next()
		require.NoError(t, d.Err())
		require.True(t, ok)
		require.Equal(t, control.Null, d.Type())
	})
}

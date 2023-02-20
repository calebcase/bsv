package control_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/calebcase/bsv/control"
	"github.com/calebcase/oops"
)

func shortName(i int, data []byte) string {
	sb := &strings.Builder{}

	sb.WriteString(fmt.Sprintf("%02d/", i))

	if len(data) == 0 {
		sb.WriteString("(len=0)")

		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("%02x", data[0]))
	prev := data[0]
	var dots bool

	for i, b := range data[1:] {
		if len(data) > 16 && prev == b {
			if !dots {
				if (i+1)%2 == 0 {
					sb.WriteString("_")
				}

				sb.WriteString("..")
				dots = true
			}

			continue
		}

		if (i+1)%2 == 0 {
			sb.WriteString("_")
		}

		sb.WriteString(fmt.Sprintf("%02x", b))
		prev = b
		dots = false
	}

	sb.WriteString(fmt.Sprintf("(len=%d)", len(data)))

	return sb.String()
}

func TestEncoder(t *testing.T) {
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
				Output: []byte{0b_0100_0001, 0b_0010_0000, 0b_0000_0000},
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
				output := &bytes.Buffer{}
				e := control.NewEncoder(output)

				err := e.Data(tc.Input)
				require.NoError(t, err, tc.Mark)
				require.Equal(t, len(tc.Output), len(output.Bytes()), tc.Mark)
				require.Equal(t, tc.Output, output.Bytes(), tc.Mark)
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
				output := &bytes.Buffer{}
				e := control.NewEncoder(output)

				err := e.Bound(tc.Input)
				require.NoError(t, err, tc.Mark)
				require.Equal(t, len(tc.Output), len(output.Bytes()), tc.Mark)
				require.Equal(t, tc.Output, output.Bytes(), tc.Mark)
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
				output := &bytes.Buffer{}
				e := control.NewEncoder(output)

				err := e.Unbound(func(e control.Encoder) (err error) {
					return e.Data(tc.Input)
				})
				require.NoError(t, err, tc.Mark)
				require.Equal(t, len(tc.Output), len(output.Bytes()), tc.Mark)
				require.Equal(t, tc.Output, output.Bytes(), tc.Mark)
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
				output := &bytes.Buffer{}
				e := control.NewEncoder(output)

				err := e.Skip(tc.Amount)
				require.NoError(t, err, tc.Mark)
				require.Equal(t, len(tc.Output), len(output.Bytes()), tc.Mark)
				require.Equal(t, tc.Output, output.Bytes(), tc.Mark)
			})
		}
	})

	t.Run("symmetric", func(t *testing.T) {
		type TC struct {
			Fn     func(control.Encoder) error
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
				Output: []byte{
					0b_0000_0000,
				},
				Mark: oops.New("unexpected"),
			},
		}

		for i, tc := range tcs {
			t.Run(shortName(i, tc.Output), func(t *testing.T) {
				output := &bytes.Buffer{}
				e := control.NewEncoder(output)

				err := e.Symmetric(tc.Fn)
				require.NoError(t, err, tc.Mark)
				t.Logf("len(expect)=%d", len(tc.Output))
				t.Logf("len(output)=%d", len(output.Bytes()))
				require.Equal(t, tc.Output, output.Bytes(), tc.Mark)
			})
		}
	})

	t.Run("empty", func(t *testing.T) {
		output := &bytes.Buffer{}
		e := control.NewEncoder(output)

		err := e.Empty()
		require.NoError(t, err)
		require.Equal(t, 1, len(output.Bytes()))
		require.Equal(t, []byte{0b_0000_0001}, output.Bytes())
	})

	t.Run("null", func(t *testing.T) {
		output := &bytes.Buffer{}
		e := control.NewEncoder(output)

		err := e.Null()
		require.NoError(t, err)
		require.Equal(t, 1, len(output.Bytes()))
		require.Equal(t, []byte{0b_0000_0000}, output.Bytes())
	})
}

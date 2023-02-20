package control

import (
	"io"
	"math"
	"math/big"
)

type Encoder interface {
	Data(data []byte) (err error)
	Bound(bsv []byte) (err error)
	Unbound(fn func(Encoder) error) (err error)
	Symmetric(fn func(Encoder) error) (err error)
	Skip(amount uint64) (err error)
	Empty() (err error)
	Null() (err error)
}

type encoder struct {
	w io.Writer

	// symmetric is true when the encoder is in symmetric mode. In this
	// mode the encoder automatically writes the trailing blocks for a
	// symmetric field. It also tracks whether a field has been written. If
	// it has, then attempts to write another field must fail.
	symmetric bool
	written   bool
}

func NewEncoder(w io.Writer) Encoder {
	e := &encoder{
		w: w,
	}

	return e
}

func (e *encoder) Data(data []byte) (err error) {
	if e.symmetric && e.written {
		return Error.New("invalid: symmetric field already written")
	}
	defer func() {
		e.written = true
	}()

	size := len(data)

	switch {
	case size == 0:
		return Error.New("invalid: size=0")
	case size == 1 && data[0]&Data.Mask == data[0]:
		_, err = e.w.Write([]byte{
			Data.Prefix | data[0],
		})
		if err != nil {
			return err
		}

		return nil
	case size == 1:
		_, err = e.w.Write([]byte{
			DataSize.Prefix,
			data[0],
		})
		if err != nil {
			return err
		}

		if e.symmetric {
			_, err = e.w.Write([]byte{
				DataSize.Prefix,
			})
			if err != nil {
				return err
			}
		}
	case size == 2 && data[0]&Data1.Mask == data[0]:
		_, err = e.w.Write([]byte{
			Data1.Prefix | data[0],
			data[1],
		})
		if err != nil {
			return err
		}

		if e.symmetric {
			_, err = e.w.Write([]byte{
				Data1.Prefix | data[0],
			})
			if err != nil {
				return err
			}
		}

		return nil
	case size == 2:
		_, err = e.w.Write([]byte{
			DataSize.Prefix | (2 - 1),
			data[0],
			data[1],
		})
		if err != nil {
			return err
		}

		if e.symmetric {
			_, err = e.w.Write([]byte{
				DataSize.Prefix | (2 - 1),
			})
			if err != nil {
				return err
			}
		}
	case size == 3 && data[0]&Data2.Mask == data[0]:
		_, err = e.w.Write([]byte{
			Data2.Prefix | data[0],
			data[1],
			data[2],
		})
		if err != nil {
			return err
		}

		if e.symmetric {
			_, err = e.w.Write([]byte{
				Data2.Prefix | data[0],
			})
			if err != nil {
				return err
			}
		}

		return nil
	case size == 3:
		_, err = e.w.Write([]byte{
			DataSize.Prefix | (3 - 1),
			data[0],
			data[1],
			data[2],
		})
		if err != nil {
			return err
		}

		if e.symmetric {
			_, err = e.w.Write([]byte{
				DataSize.Prefix | (3 - 1),
			})
			if err != nil {
				return err
			}
		}
	case size <= 64:
		_, err = e.w.Write(append(
			[]byte{DataSize.Prefix | byte(size-1)},
			data...,
		))
		if err != nil {
			return err
		}

		if e.symmetric {
			_, err = e.w.Write([]byte{
				DataSize.Prefix | byte(size-1),
			})
			if err != nil {
				return err
			}
		}
	case size <= math.MaxInt64:
		s := new(big.Int).SetUint64(uint64(size - 1))
		sb := s.Bytes()
		if len(sb) == 0 {
			sb = []byte{0b_0000_0000}
		}

		_, err = e.w.Write([]byte{DataSizeSize.Prefix | byte(len(sb)-1)})
		if err != nil {
			return err
		}

		_, err = e.w.Write(sb)
		if err != nil {
			return err
		}

		_, err = e.w.Write(data)
		if err != nil {
			return err
		}

		if e.symmetric {
			_, err = e.w.Write(sb)
			if err != nil {
				return err
			}

			_, err = e.w.Write([]byte{DataSizeSize.Prefix | byte(len(sb)-1)})
			if err != nil {
				return err
			}
		}
	default:
		return Error.New("unimplemented: size>2^64")
	}

	return nil
}

func (e *encoder) Bound(bsv []byte) (err error) {
	if e.symmetric && e.written {
		return Error.New("invalid: symmetric field already written")
	}
	defer func() {
		e.written = true
	}()

	size := uint64(len(bsv))

	if size == 0 {
		return Error.New("invalid: size=0")
	}

	_, err = e.w.Write([]byte{
		0b_0000_0101,
	})
	if err != nil {
		return Error.Trace(err)
	}

	s := new(big.Int).SetUint64(size)
	s.Sub(s, big.NewInt(1))

	writeSize := func(e Encoder) (err error) {
		sizeBytes := s.Bytes()
		if len(sizeBytes) == 0 {
			sizeBytes = []byte{0b_0000_0000}
		}

		err = e.Data(sizeBytes)
		if err != nil {
			return err
		}

		return nil
	}

	if e.symmetric {
		err = e.Symmetric(writeSize)
		if err != nil {
			return err
		}

		e.written = false
	} else {
		err = writeSize(e)
		if err != nil {
			return err
		}
	}

	_, err = e.w.Write(bsv)
	if err != nil {
		return Error.Trace(err)
	}

	if e.symmetric {
		err = e.Symmetric(writeSize)
		if err != nil {
			return err
		}

		e.written = false

		_, err = e.w.Write([]byte{
			0b_0000_0101,
		})
		if err != nil {
			return Error.Trace(err)
		}
	}

	return err
}

func (e *encoder) Unbound(fn func(Encoder) error) (err error) {
	if e.symmetric && e.written {
		return Error.New("invalid: symmetric field already written")
	}
	defer func() {
		e.written = true
	}()

	_, err = e.w.Write([]byte{
		0b_0000_0110,
	})
	if err != nil {
		return err
	}

	err = fn(e)
	if err != nil {
		return err
	}

	_, err = e.w.Write([]byte{
		0b_0000_0100,
	})
	if err != nil {
		return err
	}

	return err
}

type symmetric struct {
	e *encoder
}

func (se *symmetric) Data(data []byte) (err error) {
	if len(data) == 1 {
		return se.e.Data(data)
	}

	_, err = se.e.w.Write([]byte{
		0b_0000_0111,
	})
	if err != nil {
		return err
	}

	err = se.e.Data(data)
	if err != nil {
		return err
	}

	_, err = se.e.w.Write([]byte{
		0b_0000_0111,
	})
	if err != nil {
		return err
	}

	return nil
}

func (se *symmetric) Bound(bsv []byte) (err error) {
	_, err = se.e.w.Write([]byte{
		0b_0000_0111,
	})
	if err != nil {
		return err
	}

	err = se.e.Bound(bsv)
	if err != nil {
		return err
	}

	_, err = se.e.w.Write([]byte{
		0b_0000_0111,
	})
	if err != nil {
		return err
	}

	return nil
}

func (se *symmetric) Unbound(fn func(Encoder) error) (err error) {
	return se.e.Unbound(fn)
}

func (se *symmetric) Symmetric(fn func(Encoder) error) (err error) {
	_, err = se.e.w.Write([]byte{
		0b_0000_0111,
	})
	if err != nil {
		return err
	}

	err = se.e.Symmetric(fn)
	if err != nil {
		return err
	}

	_, err = se.e.w.Write([]byte{
		0b_0000_0111,
	})
	if err != nil {
		return err
	}

	return nil
}

func (se *symmetric) Skip(amount uint64) (err error) {
	_, err = se.e.w.Write([]byte{
		0b_0000_0111,
	})
	if err != nil {
		return err
	}

	err = se.e.Skip(amount)
	if err != nil {
		return err
	}

	_, err = se.e.w.Write([]byte{
		0b_0000_0111,
	})
	if err != nil {
		return err
	}

	return nil
}

func (se *symmetric) Empty() (err error) {
	return se.e.Empty()
}

func (se *symmetric) Null() (err error) {
	return se.e.Null()
}

func (e *encoder) Symmetric(fn func(Encoder) error) (err error) {
	if e.symmetric && e.written {
		return Error.New("invalid: symmetric field already written")
	}
	defer func() {
		e.written = true
	}()

	se := &symmetric{&encoder{
		w:         e.w,
		symmetric: true,
	}}

	err = fn(se)
	if err != nil {
		return err
	}

	if !se.e.written {
		return Error.New("invalid: symmetric field empty")
	}

	return err
}

func (e *encoder) Skip(amount uint64) (err error) {
	if e.symmetric && e.written {
		return Error.New("invalid: symmetric field already written")
	}
	defer func() {
		e.written = true
	}()

	a := new(big.Int).SetUint64(amount)
	a.Sub(a, big.NewInt(1))

	amountBytes := a.Bytes()
	if len(amountBytes) == 0 {
		amountBytes = []byte{0b_0000_0000}
	}

	switch len(amountBytes) {
	case 1:
		_, err = e.w.Write([]byte{
			0b_0000_0010,
		})
		if err != nil {
			return err
		}
	case 2:
		_, err = e.w.Write([]byte{
			0b_0000_0011,
		})
		if err != nil {
			return err
		}
	default:
		return Error.New("invalid: amount=%d len=%d", amount, len(amountBytes))
	}

	_, err = e.w.Write(amountBytes)
	if err != nil {
		return err
	}

	if e.symmetric {
		switch len(amountBytes) {
		case 1:
			_, err = e.w.Write([]byte{
				0b_0000_0010,
			})
			if err != nil {
				return err
			}
		case 2:
			_, err = e.w.Write([]byte{
				0b_0000_0011,
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (e *encoder) Empty() (err error) {
	if e.symmetric && e.written {
		return Error.New("invalid: symmetric field already written")
	}
	defer func() {
		e.written = true
	}()

	_, err = e.w.Write([]byte{
		0b_0000_0001,
	})
	if err != nil {
		return err
	}

	return err
}

func (e *encoder) Null() (err error) {
	if e.symmetric && e.written {
		return Error.New("invalid: symmetric field already written")
	}
	defer func() {
		e.written = true
	}()

	_, err = e.w.Write([]byte{
		0b_0000_0000,
	})
	if err != nil {
		return err
	}

	return err
}

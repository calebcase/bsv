package control

import (
	"encoding/binary"
	"errors"
	"io"
	"math/big"

	"github.com/calebcase/oops"
)

var ErrInvalidOperation = Error.New("invalid operation")

type Decoder interface {
	Seek() (err error)
	Next() (ok bool)
	Err() (err error)

	Type() Type
	Depth() int
	Stack() Stack
	Consumed() uint64

	Size() (_ uint64, err error)
	Data() (data []byte, err error)
	Enter() (err error)
	BSV() (bsv []byte, err error)
	Amount() (_ uint64, err error)
}

type decoder struct {
	r io.Reader
	s io.Seeker

	consumed uint64

	stack *Stack

	value    [1]byte
	t        Type
	finished bool

	size   uint64
	data   []byte
	amount uint64

	err error
}

func NewDecoder(r io.Reader) Decoder {
	d := &decoder{
		r:     r,
		stack: &Stack{},
	}

	d.s, _ = r.(io.Seeker)

	return d
}

// seek moves the input stream's current position using an io.Seeker if
// available otherwise it falls back to a discarding copy.
func (d *decoder) seek(size uint64) (err error) {
	if d.s != nil {
		n, err := d.s.Seek(int64(size), io.SeekCurrent)
		if err != nil {
			return Error.Trace(err)
		}

		d.consumed += uint64(n)
		err = d.stack.Consume(uint64(n))
		if err != nil {
			return err
		}

		return nil
	}

	n, err := io.CopyN(io.Discard, d.r, int64(size))
	if err != nil {
		return Error.Trace(err)
	}

	d.consumed += uint64(n)
	err = d.stack.Consume(uint64(n))
	if err != nil {
		return err
	}

	return nil
}

// Seek moves the reading position to the end of the current field.
func (d *decoder) Seek() (err error) {
	defer func() {
		if err != nil {
			d.err = err
		}
	}()

	if d.consumed == 0 {
		return nil
	}

	switch d.t {
	case Data:
		// No additional bytes need to be read.
	case DataSize:
		// Seek past the data if we haven't read it yet.
		if len(d.data) == 0 {
			size, err := d.Size()
			if err != nil {
				return err
			}

			err = d.seek(size)
			if err != nil {
				return err
			}
		}
	case Data1:
		// Small enough to just read directly.
		_, err := d.Data()
		if err != nil {
			return err
		}
	case Data2:
		// Small enough to just read directly.
		_, err := d.Data()
		if err != nil {
			return err
		}
	case DataSizeSize:
		size, err := d.Size()
		if err != nil {
			return err
		}

		// Seek past the data if we haven't read it yet.
		if len(d.data) == 0 {
			err = d.seek(size)
			if err != nil {
				return err
			}
		}
	case ContainerSymmetric:
		if !d.finished {
			d.finished = true

			// Move to the embedded field.
			ok := d.Next()
			if d.Err() != nil {
				return d.Err()
			}
			if !ok {
				return
			}

			d.stack.Top().Subtype = d.t

			// Seek past the embedded field.
			err = d.Seek()
			if err != nil {
				return err
			}
		}
	case ContainerBounded:
		size, err := d.Size()
		if err != nil {
			return err
		}

		// Seek past the bsv if we haven't read it yet.
		if len(d.data) == 0 {
			err = d.seek(size)
			if err != nil {
				return err
			}
		}
	case ContainerUnbounded:
		// Read tokens until the matching ContainerEnd is found. Depth
		// will be one less than our current.
		target := d.Depth() - 1
		d.finished = true

		for d.Next() {
			t := d.Type()
			if t == ContainerEnd && target == d.Depth() {
				break
			}
		}
		err = d.Err()
		if err != nil {
			return err
		}
	case ContainerEnd:
		// No additional bytes need to be read.
	case SkipSize:
		_, err := d.Amount()
		if err != nil {
			return err
		}
	case Empty:
		// No additional bytes need to be read.
	case Null:
		// No additional bytes need to be read.
	default:
		return Error.New("unknown field %q: %0b", d.t.Abbr, d.value)
	}

	// Skip over trailing symmetric control blocks.
	if top := d.stack.Top(); top != nil &&
		top.Type == ContainerSymmetric &&
		top.Subtype != Unknown {

		err = d.seek(top.Count)
		if err != nil {
			return err
		}

		err = d.stack.Pop()
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *decoder) Next() (ok bool) {
	// Ensure current field was fully read before moving on...
	if !d.finished {
		d.err = d.Seek()
		if d.err != nil {
			return false
		}
	}

	// Reset state for next field.
	d.value[0] = 0
	d.t = Unknown

	d.size = 0
	d.data = d.data[:0]
	d.amount = 0
	d.finished = false

	// Read the field control block.
	_, d.err = io.ReadFull(d.r, d.value[:])
	if d.err != nil {
		if errors.Is(d.err, io.EOF) {
			d.err = nil
			return false
		}

		d.err = Error.Trace(d.err)

		return false
	}

	d.consumed += 1
	d.stack.Consume(1)

	t, ok := Types.Match(d.value[0])
	if !ok {
		d.err = Error.New("unexpected byte: %0b", d.value[0])

		return false
	}

	switch t {
	case Data, Empty, Null:
		// These are single byte fields that don't add to the symmetric
		// control block count and are fully read.
	case ContainerUnbounded, ContainerEnd:
		// These are already symmetric and do not add to the symmetric
		// control block count.
	default:
		d.stack.Count(1)
	}

	if top := d.stack.Top(); top != nil &&
		top.Type == ContainerSymmetric &&
		top.Subtype == Unknown {

		top.Subtype = t
	}

	switch t {
	case Data:
		d.finished = true
	case ContainerSymmetric:
		d.stack.Push(&Frame{
			Type:  t,
			Count: 1,
		})
	case ContainerBounded:
		d.stack.Push(&Frame{
			Type: t,
		})
	case ContainerUnbounded:
		d.stack.Push(&Frame{
			Type: t,
		})
	case ContainerEnd:
		top := d.stack.Top()
		if top == nil {
			d.err = Error.New("unexpected container end (not in a container)")

			return false
		}

		if top.Type != ContainerUnbounded {
			d.err = Error.New(
				"unexpected container end (container not unbounded): %s",
				top.Type.Abbr,
			)

			return false
		}

		d.err = d.stack.Pop()
		if d.err != nil {
			return false
		}

		d.finished = true
	case Empty:
		d.finished = true
	case Null:
		d.finished = true
	}

	d.t = t

	return true
}

func (d *decoder) Err() error {
	return d.err
}

func (d *decoder) Type() Type {
	return d.t
}

func (d *decoder) Depth() int {
	return len(*d.stack)
}

func (d *decoder) Stack() Stack {
	return *d.stack
}

func (d *decoder) Consumed() uint64 {
	return d.consumed
}

func (d *decoder) Size() (_ uint64, err error) {
	defer func() {
		if err != nil {
			d.size = 0
			d.err = err
		}
	}()

	if d.size != 0 {
		return d.size, nil
	}

	switch d.t {
	case Data:
		d.size = 1
	case DataSize:
		d.size = uint64(d.value[0]&d.t.Mask) + 1
	case Data1:
		d.size = 2
	case Data2:
		d.size = 3
	case DataSizeSize:
		sizeSize := uint64(d.value[0]&d.t.Mask) + 1

		sizeBytes := make([]byte, sizeSize)
		_, err = io.ReadFull(d.r, sizeBytes)
		if err != nil {
			return 0, Error.Trace(err)
		}

		d.consumed += sizeSize
		d.stack.Consume(sizeSize)
		d.stack.Count(sizeSize)

		size := new(big.Int).SetBytes(sizeBytes)
		size.Add(size, big.NewInt(1))
		if !size.IsUint64() {
			return 0, Error.New("unimplemented: size >= 2^64")
		}

		d.size = size.Uint64()
	case ContainerBounded:
		zd := NewDecoder(d.r)

		ok := zd.Next()
		if !ok {
			return 0, Error.New("unabled to read container bounded size")
		}
		if zd.Err() != nil {
			return 0, zd.Err()
		}

		switch zd.Type() {
		case ContainerSymmetric, ContainerBounded, ContainerUnbounded:
			err = zd.Enter()
			if err != nil {
				return 0, err
			}

			ok := zd.Next()
			if !ok {
				return 0, Error.New("unabled to read container bounded size")
			}
			if zd.Err() != nil {
				return 0, zd.Err()
			}
		}

		sizeBytes, err := zd.Data()
		if err != nil {
			return 0, err
		}

		d.consumed += zd.Consumed()
		d.stack.Consume(zd.Consumed())
		d.stack.Count(zd.Consumed())

		size := new(big.Int).SetBytes(sizeBytes)
		size.Add(size, big.NewInt(1))
		if !size.IsUint64() {
			return 0, Error.New("unimplemented: size >= 2^64")
		}

		d.size = size.Uint64()

		top := d.stack.Top()
		top.Size = d.size
		top.Remaining = d.size

		d.finished = true
	case SkipSize:
		d.size = uint64(d.value[0]&d.t.Mask) + 1
	default:
		return 0, oops.Trace(ErrInvalidOperation)
	}

	return d.size, nil
}

// Data reads data bits and bytes from the field. If the field does not contain
// data it returns nil and ErrInvalidOperation.
func (d *decoder) Data() (data []byte, err error) {
	defer func() {
		if err != nil {
			d.data = d.data[:0]
			d.err = err
		}
	}()

	if d.t != Data && d.t != Data1 && d.t != Data2 && d.t != DataSize && d.t != DataSizeSize {
		return nil, oops.Trace(ErrInvalidOperation)
	}

	if len(d.data) != 0 {
		return d.data, nil
	}

	_, err = d.Size()
	if err != nil {
		return d.data[:0], err
	}

	switch d.t {
	case Data:
		d.data = []byte{d.value[0] & d.t.Mask}
	case DataSize:
		d.data = make([]byte, d.size)

		_, err = io.ReadFull(d.r, d.data)
		if err != nil {
			return nil, Error.Trace(err)
		}

		d.consumed += d.size
		d.stack.Consume(d.size)

		d.finished = true
	case Data1:
		d.data = make([]byte, d.size)
		d.data[0] = d.value[0] & d.t.Mask

		_, err = io.ReadFull(d.r, d.data[1:])
		if err != nil {
			return nil, Error.Trace(err)
		}

		d.consumed += 1
		d.stack.Consume(1)

		d.finished = true
	case Data2:
		d.data = make([]byte, d.size)
		d.data[0] = d.value[0] & d.t.Mask

		_, err = io.ReadFull(d.r, d.data[1:])
		if err != nil {
			return nil, Error.Trace(err)
		}

		d.consumed += 2
		d.stack.Consume(2)

		d.finished = true
	case DataSizeSize:
		d.data = make([]byte, d.size)

		_, err = io.ReadFull(d.r, d.data)
		if err != nil {
			return nil, Error.Trace(err)
		}

		d.consumed += d.size
		d.stack.Consume(d.size)

		d.finished = true
	}

	// Skip over trailing symmetric control blocks.
	if top := d.stack.Top(); top != nil &&
		top.Type == ContainerSymmetric &&
		top.Subtype != Unknown {

		err = d.seek(top.Count)
		if err != nil {
			return nil, err
		}

		err = d.stack.Pop()
		if err != nil {
			return nil, err
		}
	}

	return d.data, nil
}

// Enter informs decoder that the ContainerBounded or ContainerUnbounded field
// should be entered.  If the current field type is not ContainerBounded or
// ContainerUnbounded, then it returns ErrInvalidOperation.
func (d *decoder) Enter() (err error) {
	defer func() {
		if err != nil {
			d.err = err
		}
	}()

	switch d.t {
	case ContainerSymmetric:
		d.finished = true
	case ContainerBounded:
		_, err := d.Size()
		if err != nil {
			return err
		}
	case ContainerUnbounded:
		d.finished = true
	default:
		return oops.Trace(ErrInvalidOperation)
	}

	return nil
}

// BSV returns the embedded BSV. If the current field type is not
// ContainerBounded, then it returns nil and ErrInvalidOperation.
func (d *decoder) BSV() (bsv []byte, err error) {
	defer func() {
		if err != nil {
			d.data = d.data[:0]
			d.err = err
		}
	}()

	if d.t != ContainerBounded {
		return nil, oops.Trace(ErrInvalidOperation)
	}

	err = d.Enter()
	if err != nil {
		return nil, err
	}

	size, err := d.Size()
	if err != nil {
		return nil, err
	}

	d.data = make([]byte, size)

	_, err = io.ReadFull(d.r, d.data)
	if err != nil {
		return nil, Error.Trace(err)
	}

	d.consumed += size
	d.stack.Consume(size)

	// Skip over trailing symmetric control blocks.
	if top := d.stack.Top(); top != nil &&
		top.Type == ContainerSymmetric &&
		top.Subtype != Unknown {

		err = d.seek(top.Count)
		if err != nil {
			return nil, err
		}

		err = d.stack.Pop()
		if err != nil {
			return nil, err
		}
	}

	return d.data, nil
}

// Amount returns the skip amount. If the current field type is not SkipSize,
// then it returns 0 and ErrInvalidOperation.
func (d *decoder) Amount() (_ uint64, err error) {
	defer func() {
		if err != nil {
			d.amount = 0
			d.err = err
		}
	}()

	if d.t != SkipSize {
		return 0, oops.Trace(ErrInvalidOperation)
	}

	if d.amount != 0 {
		return d.amount, nil
	}

	_, err = d.Size()
	if err != nil {
		return 0, err
	}

	amountBytes := make([]byte, d.size)

	_, err = io.ReadFull(d.r, amountBytes)
	if err != nil {
		return 0, Error.Trace(err)
	}

	d.consumed += d.size
	d.stack.Consume(d.size)

	padSize := 8 - len(amountBytes)
	padBytes := make([]byte, padSize)

	amountBytes = append(padBytes, amountBytes...)

	d.amount = binary.BigEndian.Uint64(amountBytes) + 1

	d.finished = true

	// Skip over trailing symmetric control blocks.
	if top := d.stack.Top(); top != nil &&
		top.Type == ContainerSymmetric &&
		top.Subtype != Unknown {

		err = d.seek(top.Count)
		if err != nil {
			return 0, err
		}

		err = d.stack.Pop()
		if err != nil {
			return 0, err
		}
	}

	return d.amount, nil
}

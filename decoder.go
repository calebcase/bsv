package bsv

import "io"

type Block struct {
	Type  Type
	Value uint8
	Data  []byte
}

type Decoder struct {
	r io.Reader
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		r: r,
	}
}

func (d *Decoder) Read(blk *Block) (n int, err error) {
	buf := make([]byte, 1)

	n, err = r.Read(buf)
	if err != nil {
		return
	}

	t, v, err := Parse(buf[0])
	if err != nil {
		return
	}

	return 1, nil
}

package control

type Frame struct {
	Type Type

	// If the type is ContainerBounded then Size is the total Size of the
	// container and remaining indicates how many bytes remain to be read
	// from this container.
	Size      uint64
	Remaining uint64

	// If the type is ContainerSymmetric then the Subtype is the control
	// block type of the symmetric Subtype and Count is the number of
	// control blocks encountered.
	Subtype Type
	Count   uint64
}

type Stack []*Frame

func (s *Stack) Push(f *Frame) {
	*s = append(*s, f)
}

func (s *Stack) Top() *Frame {
	if len(*s) == 0 {
		return nil
	}

	return (*s)[len(*s)-1]
}

func (s *Stack) Pop() (err error) {
	top := s.Top()
	if top == nil {
		Error.New("no frame on stack")
	}

	switch top.Type {
	case ContainerBounded:
		if top.Remaining != 0 {
			return Error.New(
				"data remaining in bounded: size=%d remaining=%d",
				top.Size,
				top.Remaining,
			)
		}
	}

	*s = (*s)[:len(*s)-1]

	return nil
}

func (s *Stack) Count(blocks uint64) {
	top := s.Top()
	if top == nil {
		return
	}

	if top.Type != ContainerSymmetric && len(*s) >= 2 {
		parent := (*s)[len(*s)-2]
		if parent.Type != ContainerSymmetric {
			panic("invalid attempt to count")
		}

		parent.Count += blocks

		return
	}

	top.Count += blocks
}

func (s *Stack) Consume(size uint64) (err error) {
	top := s.Top()
	if top == nil {
		return
	}

	for i, f := range *s {
		if f.Type != ContainerBounded {
			continue
		}

		if size > f.Remaining {
			return Error.New(
				"exceeded bounded: depth=%d/%d size=%d remaining=%d consuming=%d",
				i,
				len(*s),
				f.Size,
				f.Remaining,
				size,
			)
		}

		f.Remaining -= size
	}

	for {
		top = s.Top()
		if top == nil {
			break
		}

		if top.Type == ContainerBounded && top.Remaining == 0 {
			err = s.Pop()
			if err != nil {
				return err
			}

			continue
		}

		break
	}

	return nil
}

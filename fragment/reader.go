package fragment

import (
	"fmt"
	"io"
)

type Fragment struct {
	Offset int64
	Length int
}

type Reader struct {
	src       io.ReadSeeker
	fragments []Fragment
	idx       int
	remaining int
}

func NewReader(src io.ReadSeeker, fragments []Fragment) *Reader {
	return &Reader{src: src, fragments: fragments, idx: -1, remaining: 0}
}

func (r *Reader) Read(p []byte) (n int, err error) {
	if r.idx >= len(r.fragments) {
		return 0, io.EOF
	}

	if len(p) == 0 {
		return 0, nil
	}

	if r.remaining == 0 {
		r.idx++
		if r.idx >= len(r.fragments) {
			return 0, io.EOF
		}
		next := r.fragments[r.idx]
		r.remaining = next.Length
		seeked, err := r.src.Seek(next.Offset, 0)
		if err != nil {
			return 0, fmt.Errorf("unable to seek to next offset %d: %v", next.Offset, err)
		}
		if seeked != next.Offset {
			return 0, fmt.Errorf("wanted to seek to %d but reached %d", next.Offset, seeked)
		}
	}

	target := p
	if len(p) > r.remaining {
		target = p[:r.remaining]
	}

	n, err = io.ReadFull(r.src, target)
	r.remaining -= n
	return n, err
}

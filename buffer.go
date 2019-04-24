package jingo

// buffer.go manages Buffer and its responsibilities.
// Its existence is justified by the performance increase gained
// over more general implementations and the fact that we allow
// these buffers to be pooled, which reduces our allocation
// profile quite significantly.

import (
	"io"
	"sync"
	"unsafe"
)

// Buffer is used to pass on to the encoders Marshal methods.
type Buffer struct {
	Bytes []byte
}

var _ io.Writer = &Buffer{} // commit to compatibility with io.Writer

// Write a chunk of bytes to the buffer
func (b *Buffer) Write(v []byte) (int, error) {
	b.Bytes = append(b.Bytes, v...)
	return len(v), nil
}

// WriteByte writes a single byte into the output buffer
func (b *Buffer) WriteByte(v byte) {
	b.Bytes = append(b.Bytes, v)
}

// Reset allows this to be reused by emptying
func (b *Buffer) Reset() {
	b.Bytes = b.Bytes[:0]
}

func (b *Buffer) String() string {
	return *(*string)(unsafe.Pointer(&b.Bytes))
}

// WriteTo writes the contents of our buffer to an io.Writer
func (b *Buffer) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(b.Bytes)
	return int64(n), err
}

var bufpool = sync.Pool{
	New: func() interface{} { return &Buffer{} },
}

// NewBufferFromPool returns a pointer to a zerod Buffer. This may be retrieved from a
// pool. When you're done with it, call 'ReturnToPool'.
func NewBufferFromPool() *Buffer {
	b := bufpool.Get().(*Buffer)
	b.Reset()
	return b
}

// NewBufferFromPoolWithCap returns a pointer to a zero'd Buffer with its underlying
// capacity set. This may be retrieved from a pool. When you're done with it, call 'ReturnToPool'.
func NewBufferFromPoolWithCap(size int) *Buffer {
	b := bufpool.Get().(*Buffer)

	if c := cap(b.Bytes); c < size {
		b.Bytes = make([]byte, 0, size)
	} else if c > 0 {
		b.Reset()
	}

	return b
}

// ReturnToPool puts this instance back in the underlying pool. Reading from or using this instance
// in any way after calling this is invalid.
func (b *Buffer) ReturnToPool() {
	bufpool.Put(b)
}

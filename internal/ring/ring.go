package ring

type Buffer struct {
	chunks [][]byte
	size   int
}

func New(size int) *Buffer {
	return &Buffer{size: size}
}

func (b *Buffer) Push(chunk []byte) {
	cp := make([]byte, len(chunk))
	copy(cp, chunk)
	if len(b.chunks) >= b.size {
		b.chunks = b.chunks[1:]
	}
	b.chunks = append(b.chunks, cp)
}

func (b *Buffer) Flush() [][]byte {
	out := make([][]byte, len(b.chunks))
	copy(out, b.chunks)
	b.chunks = b.chunks[:0]
	return out
}

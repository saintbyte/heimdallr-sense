package ring

import (
	"bytes"
	"testing"
)

func TestNew(t *testing.T) {
	b := New(3)
	if b == nil {
		t.Fatal("New returned nil")
	}
	if len(b.chunks) != 0 {
		t.Errorf("new buffer has %d chunks, want 0", len(b.chunks))
	}
}

func TestPush(t *testing.T) {
	b := New(3)
	b.Push([]byte{1, 2})
	b.Push([]byte{3, 4})
	b.Push([]byte{5, 6})

	if len(b.chunks) != 3 {
		t.Errorf("len = %d, want 3", len(b.chunks))
	}

	for i, want := range [][]byte{{1, 2}, {3, 4}, {5, 6}} {
		if !bytes.Equal(b.chunks[i], want) {
			t.Errorf("chunks[%d] = %v, want %v", i, b.chunks[i], want)
		}
	}
}

func TestPushOverflow(t *testing.T) {
	b := New(2)
	b.Push([]byte{1})
	b.Push([]byte{2})
	b.Push([]byte{3}) // evicts {1}

	if len(b.chunks) != 2 {
		t.Errorf("len = %d, want 2", len(b.chunks))
	}
	if !bytes.Equal(b.chunks[0], []byte{2}) {
		t.Errorf("chunks[0] = %v, want [2]", b.chunks[0])
	}
	if !bytes.Equal(b.chunks[1], []byte{3}) {
		t.Errorf("chunks[1] = %v, want [3]", b.chunks[1])
	}
}

func TestPushCopiesData(t *testing.T) {
	b := New(2)
	orig := []byte{1, 2, 3}
	b.Push(orig)
	orig[0] = 99

	if b.chunks[0][0] != 1 {
		t.Errorf("buffer was modified after original changed: got %d, want 1", b.chunks[0][0])
	}
}

func TestFlush(t *testing.T) {
	b := New(3)
	b.Push([]byte{1})
	b.Push([]byte{2})
	b.Push([]byte{3})

	out := b.Flush()

	if len(out) != 3 {
		t.Fatalf("flush returned %d items, want 3", len(out))
	}
	for i, want := range [][]byte{{1}, {2}, {3}} {
		if !bytes.Equal(out[i], want) {
			t.Errorf("out[%d] = %v, want %v", i, out[i], want)
		}
	}

	if len(b.chunks) != 0 {
		t.Errorf("buffer not cleared: len = %d, want 0", len(b.chunks))
	}
}

func TestFlushEmpty(t *testing.T) {
	b := New(3)
	out := b.Flush()
	if len(out) != 0 {
		t.Errorf("flush on empty returned %d items, want 0", len(out))
	}
}

func TestFlushCopiesData(t *testing.T) {
	b := New(2)
	b.Push([]byte{1, 2})
	out := b.Flush()
	out[0][0] = 99

	if len(b.chunks) != 0 {
		t.Errorf("buffer should be empty after flush, len = %d", len(b.chunks))
	}

	b.Push([]byte{1, 2})
	if b.chunks[0][0] != 1 {
		t.Errorf("original data was mutated: got %d, want 1", b.chunks[0][0])
	}
}

func TestSizeOne(t *testing.T) {
	b := New(1)
	b.Push([]byte{10})
	b.Push([]byte{20})
	b.Push([]byte{30})

	if len(b.chunks) != 1 {
		t.Errorf("len = %d, want 1", len(b.chunks))
	}
	if !bytes.Equal(b.chunks[0], []byte{30}) {
		t.Errorf("chunks[0] = %v, want [30]", b.chunks[0])
	}
}

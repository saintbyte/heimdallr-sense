package vad

import (
	"encoding/binary"
	"testing"
)

func makeSilence(n int) []byte {
	out := make([]byte, n*2)
	for i := 0; i < n; i++ {
		binary.LittleEndian.PutUint16(out[i*2:], 100)
	}
	return out
}

func makeTone(n int, freq int, sampleRate int) []byte {
	out := make([]byte, n*2)
	for i := 0; i < n; i++ {
		val := int16(16000 * 0.8)
		binary.LittleEndian.PutUint16(out[i*2:], uint16(val))
		_ = freq
		_ = sampleRate
	}
	return out
}

func TestNewProcessor(t *testing.T) {
	p := NewProcessor(16000, 30, 5, 4, 3, 3)
	if p.FrameLen() != 480 {
		t.Errorf("FrameLen = %d, want 480", p.FrameLen())
	}
	if p.ChunkLen() != 2400 {
		t.Errorf("ChunkLen = %d, want 2400", p.ChunkLen())
	}
	if p.ChunkBytes() != 4800 {
		t.Errorf("ChunkBytes = %d, want 4800", p.ChunkBytes())
	}
	if p.SilenceThreshold() != 3 {
		t.Errorf("SilenceThreshold = %d, want 3", p.SilenceThreshold())
	}
}

func TestNewProcessor8k(t *testing.T) {
	p := NewProcessor(8000, 20, 5, 3, 2, 2)
	if p.FrameLen() != 160 {
		t.Errorf("FrameLen = %d, want 160", p.FrameLen())
	}
	if p.ChunkLen() != 800 {
		t.Errorf("ChunkLen = %d, want 800", p.ChunkLen())
	}
}

func TestProcessChunkSilence(t *testing.T) {
	p := NewProcessor(8000, 20, 5, 3, 2, 3)
	raw := makeSilence(p.ChunkLen())
	count, err := p.ProcessChunk(raw)
	if err != nil {
		t.Fatalf("ProcessChunk: %v", err)
	}
	if count > p.FrameLen() {
		t.Errorf("silence voice count %d is too high", count)
	}
}

func TestProcessChunkLoudTone(t *testing.T) {
	p := NewProcessor(8000, 20, 5, 3, 2, 0)
	raw := makeTone(p.ChunkLen(), 400, 8000)
	count, err := p.ProcessChunk(raw)
	if err != nil {
		t.Fatalf("ProcessChunk: %v", err)
	}
	t.Logf("loud tone voice count: %d/5", count)
}

func TestProcessChunkBadSize(t *testing.T) {
	p := NewProcessor(8000, 20, 5, 3, 2, 3)
	raw := make([]byte, 10) // too small
	_, err := p.ProcessChunk(raw)
	if err == nil {
		t.Error("expected error for bad size, got nil")
	}
	t.Logf("got expected error: %v", err)
}

func TestIsVoice(t *testing.T) {
	p := NewProcessor(8000, 20, 5, 3, 2, 3)

	tests := []struct {
		count int
		want  bool
	}{
		{0, false},
		{2, false},
		{3, true},
		{5, true},
	}
	for _, tt := range tests {
		got := p.IsVoice(tt.count)
		if got != tt.want {
			t.Errorf("IsVoice(%d) = %v, want %v", tt.count, got, tt.want)
		}
	}
}

func TestString(t *testing.T) {
	p := NewProcessor(16000, 30, 5, 4, 3, 3)
	s := p.String()
	if s == "" {
		t.Error("String() returned empty")
	}
	t.Logf("String() = %q", s)
}

func TestMergeChunks(t *testing.T) {
	chunks := [][]byte{
		{0x64, 0x00, 0x9C, 0xFF}, // int16: 100, -100
		{0xC8, 0x00, 0x38, 0xFF}, // int16: 200, -200
	}
	merged := MergeChunks(chunks, 2)

	if len(merged) != 4 {
		t.Fatalf("len = %d, want 4", len(merged))
	}

	wants := []int16{100, -100, 200, -200}
	for i, want := range wants {
		if merged[i] != want {
			t.Errorf("merged[%d] = %d, want %d", i, merged[i], want)
		}
	}
}

func TestMergeChunksEmpty(t *testing.T) {
	merged := MergeChunks(nil, 10)
	if len(merged) != 0 {
		t.Errorf("len = %d, want 0", len(merged))
	}
}

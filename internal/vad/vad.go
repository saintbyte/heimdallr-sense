package vad

import (
	"encoding/binary"
	"fmt"

	webvad "github.com/rolandhe/go-vad"
)

type Processor struct {
	vad            *webvad.VAD
	frameLen       int
	chunkLen       int
	framesPerChunk int
	voiceThresh    int
	silenceThresh  int
}

func NewProcessor(sampleRate, vadFrameMs, framesPerChunk, voiceThresh, silenceThresh, vadMode int) *Processor {
	frameLen := sampleRate * vadFrameMs / 1000
	chunkLen := frameLen * framesPerChunk

	v := webvad.New()
	v.SetSampleRate(webvad.SampleRate(sampleRate))
	v.SetMode(webvad.Mode(vadMode))

	return &Processor{
		vad:            v,
		frameLen:       frameLen,
		chunkLen:       chunkLen,
		framesPerChunk: framesPerChunk,
		voiceThresh:    voiceThresh,
		silenceThresh:  silenceThresh,
	}
}

func (p *Processor) FrameLen() int   { return p.frameLen }
func (p *Processor) ChunkLen() int   { return p.chunkLen }
func (p *Processor) ChunkBytes() int { return p.chunkLen * 2 }

func (p *Processor) ProcessChunk(raw []byte) (int, error) {
	frame := make([]int16, p.frameLen)
	voiceCount := 0
	for f := 0; f < p.framesPerChunk; f++ {
		off := f * p.frameLen * 2
		for i := range frame {
			frame[i] = int16(binary.LittleEndian.Uint16(raw[off+i*2:]))
		}
		result, err := p.vad.Process(frame)
		if err != nil {
			return 0, fmt.Errorf("vad process frame %d: %w", f, err)
		}
		if result == webvad.ResultVoice {
			voiceCount++
		}
	}
	return voiceCount, nil
}

func (p *Processor) IsVoice(voiceCount int) bool {
	return voiceCount >= p.voiceThresh
}

func (p *Processor) SilenceThreshold() int {
	return p.silenceThresh
}

func (p *Processor) String() string {
	return fmt.Sprintf("frame=%dms, frames=%d, voice≥%d, silence≥%d",
		p.frameLen*1000/p.chunkLen*p.framesPerChunk/p.framesPerChunk,
		p.framesPerChunk, p.voiceThresh, p.silenceThresh)
}

func MergeChunks(chunks [][]byte, samplesPerChunk int) []int16 {
	total := len(chunks) * samplesPerChunk
	out := make([]int16, 0, total)
	for _, c := range chunks {
		for i := 0; i < len(c)-1; i += 2 {
			out = append(out, int16(binary.LittleEndian.Uint16(c[i:])))
		}
	}
	return out
}

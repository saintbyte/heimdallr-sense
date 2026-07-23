package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	cfg := Load("/nonexistent/path/config.yaml")
	if cfg.SampleRate != 16000 {
		t.Errorf("SampleRate = %d, want 16000", cfg.SampleRate)
	}
	if cfg.VadFrameMs != 30 {
		t.Errorf("VadFrameMs = %d, want 30", cfg.VadFrameMs)
	}
	if cfg.FramesPerChunk != 5 {
		t.Errorf("FramesPerChunk = %d, want 5", cfg.FramesPerChunk)
	}
	if cfg.VoiceThreshold != 4 {
		t.Errorf("VoiceThreshold = %d, want 4", cfg.VoiceThreshold)
	}
	if cfg.SilenceThreshold != 3 {
		t.Errorf("SilenceThreshold = %d, want 3", cfg.SilenceThreshold)
	}
	if cfg.VadMode != 3 {
		t.Errorf("VadMode = %d, want 3", cfg.VadMode)
	}
	if cfg.AudioSource != "pw-cat" {
		t.Errorf("AudioSource = %q, want %q", cfg.AudioSource, "pw-cat")
	}
	if cfg.RecordMode != "none" {
		t.Errorf("RecordMode = %q, want %q", cfg.RecordMode, "none")
	}
	if cfg.MinChunks != 3 {
		t.Errorf("MinChunks = %d, want 3", cfg.MinChunks)
	}
	if cfg.LogEnabled != true {
		t.Errorf("LogEnabled = %v, want true", cfg.LogEnabled)
	}
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	content := `
sample_rate: 8000
vad_frame_ms: 20
frames_per_chunk: 10
voice_threshold: 7
silence_threshold: 2
vad_mode: 1
audio_source: arecord
record_mode: file
record_dir: /tmp/recordings
pre_buffer_chunks: 5
https_url: https://example.com/upload
http_timeout: 30
min_chunks: 5
tls_skip_verify: true
log_enabled: false
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := Load(path)

	if cfg.SampleRate != 8000 {
		t.Errorf("SampleRate = %d, want 8000", cfg.SampleRate)
	}
	if cfg.VadFrameMs != 20 {
		t.Errorf("VadFrameMs = %d, want 20", cfg.VadFrameMs)
	}
	if cfg.FramesPerChunk != 10 {
		t.Errorf("FramesPerChunk = %d, want 10", cfg.FramesPerChunk)
	}
	if cfg.VoiceThreshold != 7 {
		t.Errorf("VoiceThreshold = %d, want 7", cfg.VoiceThreshold)
	}
	if cfg.SilenceThreshold != 2 {
		t.Errorf("SilenceThreshold = %d, want 2", cfg.SilenceThreshold)
	}
	if cfg.VadMode != 1 {
		t.Errorf("VadMode = %d, want 1", cfg.VadMode)
	}
	if cfg.AudioSource != "arecord" {
		t.Errorf("AudioSource = %q, want %q", cfg.AudioSource, "arecord")
	}
	if cfg.RecordMode != "file" {
		t.Errorf("RecordMode = %q, want %q", cfg.RecordMode, "file")
	}
	if cfg.RecordDir != "/tmp/recordings" {
		t.Errorf("RecordDir = %q, want %q", cfg.RecordDir, "/tmp/recordings")
	}
	if cfg.PreBufferChunks != 5 {
		t.Errorf("PreBufferChunks = %d, want 5", cfg.PreBufferChunks)
	}
	if cfg.HTTPSUrl != "https://example.com/upload" {
		t.Errorf("HTTPSUrl = %q, want %q", cfg.HTTPSUrl, "https://example.com/upload")
	}
	if cfg.HTTPTimeout != 30 {
		t.Errorf("HTTPTimeout = %d, want 30", cfg.HTTPTimeout)
	}
	if cfg.MinChunks != 5 {
		t.Errorf("MinChunks = %d, want 5", cfg.MinChunks)
	}
	if cfg.TLSSkipVerify != true {
		t.Errorf("TLSSkipVerify = %v, want true", cfg.TLSSkipVerify)
	}
	if cfg.LogEnabled != false {
		t.Errorf("LogEnabled = %v, want false", cfg.LogEnabled)
	}
}

func TestValidateSampleRate(t *testing.T) {
	valid := []int{8000, 16000, 32000, 48000}
	for _, sr := range valid {
		cfg := defaults
		cfg.SampleRate = sr
		if err := validate(&cfg); err != nil {
			t.Errorf("sample_rate %d: unexpected error: %v", sr, err)
		}
	}

	cfg := defaults
	cfg.SampleRate = 44100
	if err := validate(&cfg); err == nil {
		t.Error("sample_rate 44100: expected error, got nil")
	}
}

func TestValidateVadFrameMs(t *testing.T) {
	valid := []int{10, 20, 30}
	for _, ms := range valid {
		cfg := defaults
		cfg.VadFrameMs = ms
		if err := validate(&cfg); err != nil {
			t.Errorf("vad_frame_ms %d: unexpected error: %v", ms, err)
		}
	}

	cfg := defaults
	cfg.VadFrameMs = 15
	if err := validate(&cfg); err == nil {
		t.Error("vad_frame_ms 15: expected error, got nil")
	}
}

func TestValidateFramesPerChunk(t *testing.T) {
	cfg := defaults
	cfg.FramesPerChunk = 0
	if err := validate(&cfg); err == nil {
		t.Error("frames_per_chunk 0: expected error, got nil")
	}
}

func TestValidateVoiceThreshold(t *testing.T) {
	cfg := defaults
	cfg.VoiceThreshold = 0
	if err := validate(&cfg); err == nil {
		t.Error("voice_threshold 0: expected error, got nil")
	}

	cfg = defaults
	cfg.VoiceThreshold = 10
	cfg.FramesPerChunk = 5
	if err := validate(&cfg); err == nil {
		t.Error("voice_threshold > frames_per_chunk: expected error, got nil")
	}
}

func TestValidateSilenceThreshold(t *testing.T) {
	cfg := defaults
	cfg.SilenceThreshold = 0
	if err := validate(&cfg); err == nil {
		t.Error("silence_threshold 0: expected error, got nil")
	}
}

func TestValidateVadMode(t *testing.T) {
	for _, mode := range []int{0, 1, 2, 3} {
		cfg := defaults
		cfg.VadMode = mode
		if err := validate(&cfg); err != nil {
			t.Errorf("vad_mode %d: unexpected error: %v", mode, err)
		}
	}

	cfg := defaults
	cfg.VadMode = 4
	if err := validate(&cfg); err == nil {
		t.Error("vad_mode 4: expected error, got nil")
	}
}

func TestValidateAudioSource(t *testing.T) {
	for _, src := range []string{"pw-cat", "arecord", "custom"} {
		cfg := defaults
		cfg.AudioSource = src
		if src == "custom" {
			cfg.AudioCommand = "some command"
		}
		if err := validate(&cfg); err != nil {
			t.Errorf("audio_source %q: unexpected error: %v", src, err)
		}
	}

	cfg := defaults
	cfg.AudioSource = "invalid"
	if err := validate(&cfg); err == nil {
		t.Error("audio_source invalid: expected error, got nil")
	}
}

func TestValidateCustomRequiresCommand(t *testing.T) {
	cfg := defaults
	cfg.AudioSource = "custom"
	cfg.AudioCommand = ""
	if err := validate(&cfg); err == nil {
		t.Error("custom without command: expected error, got nil")
	}
}

func TestValidateMinChunks(t *testing.T) {
	cfg := defaults
	cfg.MinChunks = 0
	if err := validate(&cfg); err == nil {
		t.Error("min_chunks 0: expected error, got nil")
	}
}

func TestValidateHTTPTimeout(t *testing.T) {
	cfg := defaults
	cfg.HTTPTimeout = 0
	if err := validate(&cfg); err == nil {
		t.Error("http_timeout 0: expected error, got nil")
	}
}

func TestBuildCommandPwCat(t *testing.T) {
	cfg := defaults
	cfg.AudioSource = "pw-cat"
	cfg.SampleRate = 8000

	name, args := cfg.BuildCommand()
	if name != "pw-cat" {
		t.Errorf("name = %q, want %q", name, "pw-cat")
	}
	expected := []string{"-r", "--format", "s16", "--rate", "8000", "--channels", "1", "-"}
	if len(args) != len(expected) {
		t.Fatalf("args len = %d, want %d", len(args), len(expected))
	}
	for i, want := range expected {
		if args[i] != want {
			t.Errorf("args[%d] = %q, want %q", i, args[i], want)
		}
	}
}

func TestBuildCommandArecord(t *testing.T) {
	cfg := defaults
	cfg.AudioSource = "arecord"
	cfg.SampleRate = 16000

	name, args := cfg.BuildCommand()
	if name != "arecord" {
		t.Errorf("name = %q, want %q", name, "arecord")
	}
	expected := []string{"-q", "-f", "S16_LE", "-r", "16000", "-c", "1", "-t", "raw", "-"}
	if len(args) != len(expected) {
		t.Fatalf("args len = %d, want %d", len(args), len(expected))
	}
	for i, want := range expected {
		if args[i] != want {
			t.Errorf("args[%d] = %q, want %q", i, args[i], want)
		}
	}
}

func TestBuildCommandCustom(t *testing.T) {
	cfg := defaults
	cfg.AudioSource = "custom"
	cfg.AudioCommand = "ssh host arecord -f S16_LE -r 8000 -c 1 -"

	name, args := cfg.BuildCommand()
	if name != "ssh" {
		t.Errorf("name = %q, want %q", name, "ssh")
	}
	expected := []string{"host", "arecord", "-f", "S16_LE", "-r", "8000", "-c", "1", "-"}
	if len(args) != len(expected) {
		t.Fatalf("args len = %d, want %d", len(args), len(expected))
	}
	for i, want := range expected {
		if args[i] != want {
			t.Errorf("args[%d] = %q, want %q", i, args[i], want)
		}
	}
}

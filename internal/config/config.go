package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	SampleRate       int    `yaml:"sample_rate"`
	VadFrameMs       int    `yaml:"vad_frame_ms"`
	FramesPerChunk   int    `yaml:"frames_per_chunk"`
	VoiceThreshold   int    `yaml:"voice_threshold"`
	SilenceThreshold int    `yaml:"silence_threshold"`
	VadMode          int    `yaml:"vad_mode"`
	AudioSource      string `yaml:"audio_source"`
	AudioCommand     string `yaml:"audio_command"`
	RecordMode       string `yaml:"record_mode"`
	RecordDir        string `yaml:"record_dir"`
	PreBufferChunks  int    `yaml:"pre_buffer_chunks"`
	HTTPSUrl         string `yaml:"https_url"`
	HTTPTimeout      int    `yaml:"http_timeout"`
	MinChunks        int    `yaml:"min_chunks"`
	TLSSkipVerify    bool   `yaml:"tls_skip_verify"`
	LogEnabled       bool   `yaml:"log_enabled"`
}

var defaults = Config{
	SampleRate:       16000,
	VadFrameMs:       30,
	FramesPerChunk:   5,
	VoiceThreshold:   4,
	SilenceThreshold: 3,
	VadMode:          3,
	AudioSource:      "pw-cat",
	AudioCommand:     "",
	RecordMode:       "none",
	RecordDir:        "./recordings",
	PreBufferChunks:  3,
	HTTPSUrl:         "",
	HTTPTimeout:      10,
	MinChunks:        3,
	TLSSkipVerify:    false,
	LogEnabled:       true,
}

func Load(path string) Config {
	cfg := defaults

	paths := []string{
		"/etc/heimdallr-sense/config.yaml",
		path,
	}

	var data []byte
	var found string
	for _, p := range paths {
		d, err := os.ReadFile(p)
		if err == nil {
			data = d
			found = p
			break
		}
	}

	if data == nil {
		fmt.Println("no config file found, using defaults")
		return cfg
	}

	fmt.Printf("config loaded from %s\n", found)
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "parse config error: %v\n", err)
		os.Exit(1)
	}

	if err := validate(&cfg); err != nil {
		fmt.Fprintf(os.Stderr, "config validation error: %v\n", err)
		os.Exit(1)
	}

	return cfg
}

func validate(cfg *Config) error {
	if cfg.SampleRate != 8000 && cfg.SampleRate != 16000 && cfg.SampleRate != 32000 && cfg.SampleRate != 48000 {
		return fmt.Errorf("sample_rate must be 8000, 16000, 32000, or 48000, got %d", cfg.SampleRate)
	}
	if cfg.VadFrameMs != 10 && cfg.VadFrameMs != 20 && cfg.VadFrameMs != 30 {
		return fmt.Errorf("vad_frame_ms must be 10, 20, or 30, got %d", cfg.VadFrameMs)
	}
	if cfg.FramesPerChunk < 1 {
		return fmt.Errorf("frames_per_chunk must be >= 1, got %d", cfg.FramesPerChunk)
	}
	if cfg.VoiceThreshold < 1 || cfg.VoiceThreshold > cfg.FramesPerChunk {
		return fmt.Errorf("voice_threshold must be 1..%d, got %d", cfg.FramesPerChunk, cfg.VoiceThreshold)
	}
	if cfg.SilenceThreshold < 1 {
		return fmt.Errorf("silence_threshold must be >= 1, got %d", cfg.SilenceThreshold)
	}
	if cfg.VadMode < 0 || cfg.VadMode > 3 {
		return fmt.Errorf("vad_mode must be 0..3, got %d", cfg.VadMode)
	}
	if cfg.AudioSource != "pw-cat" && cfg.AudioSource != "arecord" && cfg.AudioSource != "custom" {
		return fmt.Errorf("audio_source must be pw-cat, arecord, or custom, got %q", cfg.AudioSource)
	}
	if cfg.AudioSource == "custom" && cfg.AudioCommand == "" {
		return fmt.Errorf("audio_command is required when audio_source is custom")
	}
	if cfg.MinChunks < 1 {
		return fmt.Errorf("min_chunks must be >= 1, got %d", cfg.MinChunks)
	}
	if cfg.HTTPTimeout < 1 {
		return fmt.Errorf("http_timeout must be >= 1, got %d", cfg.HTTPTimeout)
	}
	return nil
}

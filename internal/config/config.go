package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	SampleRate       int    `yaml:"sample_rate"`
	VadFrameMs       int    `yaml:"vad_frame_ms"`
	FramesPerChunk   int    `yaml:"frames_per_chunk"`
	VoiceThreshold   int    `yaml:"voice_threshold"`
	SilenceThreshold int    `yaml:"silence_threshold"`
	VadMode          int    `yaml:"vad_mode"`
	AudioSource      string `yaml:"audio_source"`    // pw-cat, arecord, custom
	AudioCommand     string `yaml:"audio_command"`   // кастомная команда
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
		log.Fatal("parse config:", err)
	}
	return cfg
}

func (c *Config) BuildCommand() (name string, args []string) {
	rate := fmt.Sprint(c.SampleRate)

	switch c.AudioSource {
	case "arecord":
		return "arecord", []string{
			"-q",
			"-f", "S16_LE",
			"-r", rate,
			"-c", "1",
			"-t", "raw",
			"-",
		}
	case "custom":
		if c.AudioCommand == "" {
			log.Fatal("audio_command is empty but audio_source is custom")
		}
		parts := strings.Fields(c.AudioCommand)
		return parts[0], parts[1:]
	default: // pw-cat
		return "pw-cat", []string{
			"-r",
			"--format", "s16",
			"--rate", rate,
			"--channels", "1",
			"-",
		}
	}
}

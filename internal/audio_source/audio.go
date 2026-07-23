package audio_source

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/saintbyte/heimdallr-sense/internal/config"
)

type Source struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
}

func New(cfg *config.Config) *Source {
	name, args := buildCommand(cfg)
	return &Source{
		cmd: exec.Command(name, args...),
	}
}

func buildCommand(cfg *config.Config) (name string, args []string) {
	rate := fmt.Sprint(cfg.SampleRate)

	switch cfg.AudioSource {
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
		parts := strings.Fields(cfg.AudioCommand)
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

func (s *Source) Start() error {
	if _, err := exec.LookPath(s.cmd.Path); err != nil {
		return fmt.Errorf("audio source not found: %w", err)
	}

	s.cmd.Stderr = os.Stderr

	stdin, err := s.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("stdin pipe: %w", err)
	}
	s.stdin = stdin

	stdout, err := s.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}
	s.stdout = stdout

	if err := s.cmd.Start(); err != nil {
		return fmt.Errorf("start audio source: %w", err)
	}

	return nil
}

func (s *Source) Stdin() io.WriteCloser {
	return s.stdin
}

func (s *Source) Stdout() io.ReadCloser {
	return s.stdout
}

func (s *Source) Stop() error {
	if s.cmd.Process == nil {
		return nil
	}
	return s.cmd.Process.Signal(syscall.SIGTERM)
}

func (s *Source) Wait() error {
	if s.cmd.Process == nil {
		return nil
	}
	return s.cmd.Wait()
}

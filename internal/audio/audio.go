package audio

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
)

type Source struct {
	cmd    *exec.Cmd
	stdout io.ReadCloser
}

func New(name string, args []string) *Source {
	return &Source{
		cmd: exec.Command(name, args...),
	}
}

func (s *Source) Start() error {
	if _, err := exec.LookPath(s.cmd.Path); err != nil {
		return fmt.Errorf("audio source not found: %w", err)
	}

	s.cmd.Stderr = os.Stderr

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

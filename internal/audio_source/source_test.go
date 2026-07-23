package audio_source

import (
	"io"
	"testing"

	"github.com/saintbyte/heimdallr-sense/internal/config"
)

func cfgPwCat() *config.Config {
	return &config.Config{
		SampleRate:  8000,
		AudioSource: "pw-cat",
	}
}

func cfgArecord() *config.Config {
	return &config.Config{
		SampleRate:  16000,
		AudioSource: "arecord",
	}
}

func cfgCustom() *config.Config {
	return &config.Config{
		SampleRate:   8000,
		AudioSource:  "custom",
		AudioCommand: "cat -",
	}
}

func TestStartCat(t *testing.T) {
	s := New(cfgCustom())
	if err := s.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer s.Stop()

	stdout := s.Stdout()
	if stdout == nil {
		t.Fatal("Stdout() returned nil")
	}

	go func() {
		s.Stdin().Write([]byte("hello\n"))
		s.Stdin().Close()
	}()

	out, err := io.ReadAll(stdout)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if string(out) != "hello\n" {
		t.Errorf("output = %q, want %q", string(out), "hello\n")
	}
}

func TestStop(t *testing.T) {
	s := New(&config.Config{
		SampleRate:   8000,
		AudioSource:  "custom",
		AudioCommand: "sleep 100",
	})
	if err := s.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}

	if err := s.Stop(); err != nil {
		t.Errorf("Stop: %v", err)
	}

	err := s.Wait()
	if err == nil {
		t.Error("Wait: expected error from killed process, got nil")
	}
}

func TestWaitNotStarted(t *testing.T) {
	s := New(&config.Config{
		SampleRate:   8000,
		AudioSource:  "custom",
		AudioCommand: "sleep 1",
	})
	err := s.Wait()
	if err != nil {
		t.Errorf("Wait on unstarted: expected nil, got %v", err)
	}
}

func TestStopNotStarted(t *testing.T) {
	s := New(&config.Config{
		SampleRate:   8000,
		AudioSource:  "custom",
		AudioCommand: "sleep 1",
	})
	err := s.Stop()
	if err != nil {
		t.Errorf("Stop on unstarted: expected nil, got %v", err)
	}
}

func TestNotFound(t *testing.T) {
	s := New(&config.Config{
		SampleRate:   8000,
		AudioSource:  "custom",
		AudioCommand: "nonexistent_binary_12345",
	})
	err := s.Start()
	if err == nil {
		s.Stop()
		t.Error("Start: expected error for missing binary, got nil")
	}
}

func TestStartTrue(t *testing.T) {
	s := New(&config.Config{
		SampleRate:   8000,
		AudioSource:  "custom",
		AudioCommand: "true",
	})
	if err := s.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if err := s.Wait(); err != nil {
		t.Errorf("Wait: %v", err)
	}
}

func TestBuildCommandPwCat(t *testing.T) {
	s := New(cfgPwCat())
	name, _ := buildCommand(cfgPwCat())
	if name != "pw-cat" {
		t.Errorf("name = %q, want %q", name, "pw-cat")
	}
	if s == nil {
		t.Fatal("New returned nil")
	}
}

func TestBuildCommandArecord(t *testing.T) {
	name, args := buildCommand(cfgArecord())
	if name != "arecord" {
		t.Errorf("name = %q, want %q", name, "arecord")
	}
	if len(args) == 0 {
		t.Error("args is empty")
	}
}

func TestBuildCommandCustom(t *testing.T) {
	name, args := buildCommand(cfgCustom())
	if name != "cat" {
		t.Errorf("name = %q, want %q", name, "cat")
	}
	if len(args) != 1 || args[0] != "-" {
		t.Errorf("args = %v, want [-]", args)
	}
}

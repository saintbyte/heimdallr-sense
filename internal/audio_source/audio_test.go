package audio_source

import (
	"io"
	"testing"
)

func TestStartCat(t *testing.T) {
	s := New("cat", []string{"-"})
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
	s := New("sleep", []string{"100"})
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
	s := New("sleep", []string{"1"})
	err := s.Wait()
	if err != nil {
		t.Errorf("Wait on unstarted: expected nil, got %v", err)
	}
}

func TestStopNotStarted(t *testing.T) {
	s := New("sleep", []string{"1"})
	err := s.Stop()
	if err != nil {
		t.Errorf("Stop on unstarted: expected nil, got %v", err)
	}
}

func TestNotFound(t *testing.T) {
	s := New("nonexistent_binary_12345", []string{})
	err := s.Start()
	if err == nil {
		s.Stop()
		t.Error("Start: expected error for missing binary, got nil")
	}
}

func TestStartTrue(t *testing.T) {
	s := New("true", []string{})
	if err := s.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if err := s.Wait(); err != nil {
		t.Errorf("Wait: %v", err)
	}
}

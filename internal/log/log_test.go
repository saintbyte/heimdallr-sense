package log

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func captureOutput(fn func()) string {
	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestInitDisabled(t *testing.T) {
	Init(false)
	if enabled {
		t.Error("expected enabled=false after Init(false)")
	}
}

func TestInitEnabled(t *testing.T) {
	out := captureOutput(func() {
		Init(true)
		if !enabled {
			t.Error("expected enabled=true after Init(true)")
		}
		Info("test message", "key", "value")
	})

	if !strings.Contains(out, "test message") {
		t.Errorf("output missing message: %q", out)
	}
	if !strings.Contains(out, "key") {
		t.Errorf("output missing key: %q", out)
	}
	if !strings.Contains(out, "value") {
		t.Errorf("output missing value: %q", out)
	}
}

func TestInfoDisabled(t *testing.T) {
	Init(false)
	out := captureOutput(func() {
		Info("should not appear")
	})
	if out != "" {
		t.Errorf("expected no output when disabled, got: %q", out)
	}
}

func TestErrorDisabled(t *testing.T) {
	Init(false)
	out := captureOutput(func() {
		Error("should not appear")
	})
	if out != "" {
		t.Errorf("expected no output when disabled, got: %q", out)
	}
}

func TestErrorEnabled(t *testing.T) {
	out := captureOutput(func() {
		Init(true)
		Error("something failed", "code", 500)
	})

	if !strings.Contains(out, "something failed") {
		t.Errorf("output missing message: %q", out)
	}
	if !strings.Contains(out, "ERROR") {
		t.Errorf("output missing ERROR level: %q", out)
	}
}

func TestInfoJSONFormat(t *testing.T) {
	out := captureOutput(func() {
		Init(true)
		Info("hello", "num", 42)
	})

	if !strings.Contains(out, `"msg":"hello"`) {
		t.Errorf("output missing msg field: %q", out)
	}
	if !strings.Contains(out, `"num":42`) {
		t.Errorf("output missing num field: %q", out)
	}
}

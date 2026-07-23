package recorder

import (
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUploadHTTPS(t *testing.T) {
	var receivedFields map[string]string
	var receivedFile string
	var receivedFilename string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedFields = make(map[string]string)

		contentType := r.Header.Get("Content-Type")
		_, params, err := mime.ParseMediaType(contentType)
		if err != nil {
			t.Errorf("parse media type: %v", err)
			w.WriteHeader(400)
			return
		}

		mr := multipart.NewReader(r.Body, params["boundary"])
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Errorf("next part: %v", err)
				break
			}

			data, _ := io.ReadAll(p)
			if p.FileName() != "" {
				receivedFile = string(data)
				receivedFilename = p.FileName()
			} else {
				receivedFields[p.FormName()] = string(data)
			}
		}
		w.WriteHeader(200)
	}))
	defer srv.Close()

	samples := []int16{100, -100, 200, -200, 300, -300}
	err := UploadHTTPS(srv.URL, samples, 8000, 5, false, "test-recording.wav")
	if err != nil {
		t.Fatalf("upload failed: %v", err)
	}

	if receivedFields["filename"] != "test-recording.wav" {
		t.Errorf("filename field = %q, want %q", receivedFields["filename"], "test-recording.wav")
	}
	if receivedFields["sample_rate"] != "8000" {
		t.Errorf("sample_rate field = %q, want %q", receivedFields["sample_rate"], "8000")
	}
	if receivedFields["duration"] != "0.00" {
		t.Errorf("duration field = %q, want %q", receivedFields["duration"], "0.00")
	}
	if receivedFilename != "test-recording.wav" {
		t.Errorf("file filename = %q, want %q", receivedFilename, "test-recording.wav")
	}
	if len(receivedFile) != 44+len(samples)*2 {
		t.Errorf("file size = %d, want %d", len(receivedFile), 44+len(samples)*2)
	}

	t.Logf("fields: %v", receivedFields)
	t.Logf("file: name=%s size=%d", receivedFilename, len(receivedFile))
}

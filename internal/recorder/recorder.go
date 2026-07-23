package recorder

import (
	"bytes"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

func writeLE(w io.Writer, data any) error {
	return binary.Write(w, binary.LittleEndian, data)
}

func WriteWAV(filename string, data []int16, sampleRate int) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	dataSize := len(data) * 2
	fileSize := 36 + dataSize

	if err := writeLE(f, []byte("RIFF")); err != nil {
		return fmt.Errorf("write RIFF: %w", err)
	}
	if err := writeLE(f, uint32(fileSize)); err != nil {
		return fmt.Errorf("write file size: %w", err)
	}
	if err := writeLE(f, []byte("WAVE")); err != nil {
		return fmt.Errorf("write WAVE: %w", err)
	}
	if err := writeLE(f, []byte("fmt ")); err != nil {
		return fmt.Errorf("write fmt: %w", err)
	}
	if err := writeLE(f, uint32(16)); err != nil {
		return fmt.Errorf("write fmt size: %w", err)
	}
	if err := writeLE(f, uint16(1)); err != nil {
		return fmt.Errorf("write audio format: %w", err)
	}
	if err := writeLE(f, uint16(1)); err != nil {
		return fmt.Errorf("write channels: %w", err)
	}
	if err := writeLE(f, uint32(sampleRate)); err != nil {
		return fmt.Errorf("write sample rate: %w", err)
	}
	if err := writeLE(f, uint32(sampleRate*2)); err != nil {
		return fmt.Errorf("write byte rate: %w", err)
	}
	if err := writeLE(f, uint16(2)); err != nil {
		return fmt.Errorf("write block align: %w", err)
	}
	if err := writeLE(f, uint16(16)); err != nil {
		return fmt.Errorf("write bits per sample: %w", err)
	}
	if err := writeLE(f, []byte("data")); err != nil {
		return fmt.Errorf("write data header: %w", err)
	}
	if err := writeLE(f, uint32(dataSize)); err != nil {
		return fmt.Errorf("write data size: %w", err)
	}

	for i, s := range data {
		if err := writeLE(f, s); err != nil {
			return fmt.Errorf("write sample %d: %w", i, err)
		}
	}
	return nil
}

func UploadHTTPS(rawURL string, data []int16, sampleRate int, timeout int, skipVerify bool, filename string) error {
	var buf bytes.Buffer
	duration := float64(len(data)) / float64(sampleRate)

	// field: filename
	buf.WriteString("--boundary\r\n")
	buf.WriteString("Content-Disposition: form-data; name=\"filename\"\r\n\r\n")
	buf.WriteString(filename)
	buf.WriteString("\r\n")

	// field: sample_rate
	buf.WriteString("--boundary\r\n")
	buf.WriteString("Content-Disposition: form-data; name=\"sample_rate\"\r\n\r\n")
	buf.WriteString(fmt.Sprintf("%d", sampleRate))
	buf.WriteString("\r\n")

	// field: duration
	buf.WriteString("--boundary\r\n")
	buf.WriteString("Content-Disposition: form-data; name=\"duration\"\r\n\r\n")
	buf.WriteString(fmt.Sprintf("%.2f", duration))
	buf.WriteString("\r\n")

	// field: file
	buf.WriteString("--boundary\r\n")
	buf.WriteString("Content-Disposition: form-data; name=\"file\"; filename=\"")
	buf.WriteString(filename)
	buf.WriteString("\"\r\nContent-Type: audio/wav\r\n\r\n")

	wavHeader := make([]byte, 44)
	dataSize := len(data) * 2
	fileSize := 36 + dataSize

	binary.LittleEndian.PutUint32(wavHeader[0:4], 0x46464952)
	binary.LittleEndian.PutUint32(wavHeader[4:8], uint32(fileSize))
	copy(wavHeader[8:12], "WAVE")
	copy(wavHeader[12:16], "fmt ")
	binary.LittleEndian.PutUint32(wavHeader[16:20], 16)
	binary.LittleEndian.PutUint16(wavHeader[20:22], 1)
	binary.LittleEndian.PutUint16(wavHeader[22:24], 1)
	binary.LittleEndian.PutUint32(wavHeader[24:28], uint32(sampleRate))
	binary.LittleEndian.PutUint32(wavHeader[28:32], uint32(sampleRate*2))
	binary.LittleEndian.PutUint16(wavHeader[32:34], 2)
	binary.LittleEndian.PutUint16(wavHeader[34:36], 16)
	copy(wavHeader[36:40], "data")
	binary.LittleEndian.PutUint32(wavHeader[40:44], uint32(dataSize))

	buf.Write(wavHeader)
	for _, s := range data {
		binary.Write(&buf, binary.LittleEndian, s)
	}
	buf.WriteString("\r\n--boundary--\r\n")

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipVerify},
	}
	client := &http.Client{
		Timeout:   time.Duration(timeout) * time.Second,
		Transport: transport,
	}

	uploadURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("parse URL: %w", err)
	}
	q := uploadURL.Query()
	q.Set("filename", filename)
	uploadURL.RawQuery = q.Encode()

	resp, err := client.Post(uploadURL.String(), "multipart/form-data; boundary=boundary", &buf)
	if err != nil {
		return fmt.Errorf("HTTPS POST: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTPS upload status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func SaveFile(recordDir, filename string, samples []int16, sampleRate int) (string, error) {
	path := filepath.Join(recordDir, filename)
	if err := WriteWAV(path, samples, sampleRate); err != nil {
		return "", err
	}
	return path, nil
}

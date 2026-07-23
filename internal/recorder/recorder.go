package recorder

import (
	"bytes"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

func WriteWAV(filename string, data []int16, sampleRate int) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	dataSize := len(data) * 2
	fileSize := 36 + dataSize

	wav := []byte{
		'R', 'I', 'F', 'F',
		byte(fileSize), byte(fileSize >> 8), byte(fileSize >> 16), byte(fileSize >> 24),
		'W', 'A', 'V', 'E',
		'f', 'm', 't', ' ',
		0x10, 0x00, 0x00, 0x00,
		0x01, 0x00,
		0x01, 0x00,
		byte(sampleRate), byte(sampleRate >> 8), byte(sampleRate >> 16), byte(sampleRate >> 24),
		byte(sampleRate * 2), byte((sampleRate * 2) >> 8), byte((sampleRate * 2) >> 16), byte((sampleRate * 2) >> 24),
		0x02, 0x00,
		0x10, 0x00,
		'd', 'a', 't', 'a',
		byte(dataSize), byte(dataSize >> 8), byte(dataSize >> 16), byte(dataSize >> 24),
	}

	if _, err := f.Write(wav); err != nil {
		return fmt.Errorf("write wav header: %w", err)
	}

	for _, s := range data {
		if err := binary.Write(f, binary.LittleEndian, s); err != nil {
			return fmt.Errorf("write sample: %w", err)
		}
	}
	return nil
}

func UploadHTTPS(rawURL string, data []int16, sampleRate int, timeout int, skipVerify bool, filename string) error {
	duration := float64(len(data)) / float64(sampleRate)

	var body bytes.Buffer
	w := multipart.NewWriter(&body)

	// text fields
	w.WriteField("filename", filename)
	w.WriteField("sample_rate", fmt.Sprintf("%d", sampleRate))
	w.WriteField("duration", fmt.Sprintf("%.2f", duration))

	// file field
	part, err := w.CreateFormFile("file", filename)
	if err != nil {
		return fmt.Errorf("create file part: %w", err)
	}

	wav := buildWAV(data, sampleRate)
	if _, err := part.Write(wav); err != nil {
		return fmt.Errorf("write wav to multipart: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("close multipart writer: %w", err)
	}

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

	resp, err := client.Post(uploadURL.String(), w.FormDataContentType(), &body)
	if err != nil {
		return fmt.Errorf("HTTPS POST: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTPS upload status %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}

func buildWAV(data []int16, sampleRate int) []byte {
	dataSize := len(data) * 2
	fileSize := 36 + dataSize

	buf := make([]byte, 44+dataSize)
	binary.LittleEndian.PutUint32(buf[0:4], 0x46464952) // RIFF
	binary.LittleEndian.PutUint32(buf[4:8], uint32(fileSize))
	copy(buf[8:12], "WAVE")
	copy(buf[12:16], "fmt ")
	binary.LittleEndian.PutUint32(buf[16:20], 16)
	binary.LittleEndian.PutUint16(buf[20:22], 1) // PCM
	binary.LittleEndian.PutUint16(buf[22:24], 1) // mono
	binary.LittleEndian.PutUint32(buf[24:28], uint32(sampleRate))
	binary.LittleEndian.PutUint32(buf[28:32], uint32(sampleRate*2))
	binary.LittleEndian.PutUint16(buf[32:34], 2)
	binary.LittleEndian.PutUint16(buf[34:36], 16)
	copy(buf[36:40], "data")
	binary.LittleEndian.PutUint32(buf[40:44], uint32(dataSize))

	for i, s := range data {
		binary.LittleEndian.PutUint16(buf[44+i*2:], uint16(s))
	}
	return buf
}

func SaveFile(recordDir, filename string, samples []int16, sampleRate int) (string, error) {
	path := filepath.Join(recordDir, filename)
	if err := WriteWAV(path, samples, sampleRate); err != nil {
		return "", err
	}
	return path, nil
}

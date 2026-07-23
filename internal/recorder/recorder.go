package recorder

import (
	"bytes"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func WriteWAV(filename string, data []int16, sampleRate int) error {
	tmp := filename + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}

	dataSize := len(data) * 2
	fileSize := 36 + dataSize

	binary.Write(f, binary.LittleEndian, []byte("RIFF"))
	binary.Write(f, binary.LittleEndian, uint32(fileSize))
	binary.Write(f, binary.LittleEndian, []byte("WAVE"))
	binary.Write(f, binary.LittleEndian, []byte("fmt "))
	binary.Write(f, binary.LittleEndian, uint32(16))
	binary.Write(f, binary.LittleEndian, uint16(1))
	binary.Write(f, binary.LittleEndian, uint16(1))
	binary.Write(f, binary.LittleEndian, uint32(sampleRate))
	binary.Write(f, binary.LittleEndian, uint32(sampleRate*2))
	binary.Write(f, binary.LittleEndian, uint16(2))
	binary.Write(f, binary.LittleEndian, uint16(16))
	binary.Write(f, binary.LittleEndian, []byte("data"))
	binary.Write(f, binary.LittleEndian, uint32(dataSize))

	for _, s := range data {
		binary.Write(f, binary.LittleEndian, s)
	}

	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}

	return os.Rename(tmp, filename)
}

func UploadHTTPS(url string, data []int16, sampleRate int, timeout int, skipVerify bool, filename string) error {
	var buf bytes.Buffer

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

	resp, err := client.Post(url, "multipart/form-data; boundary=boundary", &buf)
	if err != nil {
		return fmt.Errorf("HTTPS upload failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTPS upload: status %d", resp.StatusCode)
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

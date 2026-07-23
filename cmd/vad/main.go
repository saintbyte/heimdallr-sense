package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/saintbyte/heimdallr-sense/internal/audio_source"
	"github.com/saintbyte/heimdallr-sense/internal/config"
	"github.com/saintbyte/heimdallr-sense/internal/log"
	"github.com/saintbyte/heimdallr-sense/internal/ring"
	"github.com/saintbyte/heimdallr-sense/internal/recorder"
	"github.com/saintbyte/heimdallr-sense/internal/vad"
)

func main() {
	cfg := config.Load("config.yaml")
	log.Init(cfg.LogEnabled)

	proc := vad.NewProcessor(cfg.SampleRate, cfg.VadFrameMs, cfg.FramesPerChunk,
		cfg.VoiceThreshold, cfg.SilenceThreshold, cfg.VadMode)

	if cfg.RecordMode != "none" && cfg.RecordMode != "" {
		if err := os.MkdirAll(cfg.RecordDir, 0755); err != nil {
			log.Fatal("create record dir", "error", err, "dir", cfg.RecordDir)
		}
	}

	name, args := cfg.BuildCommand()
	src := audio_source.New(name, args)
	if err := src.Start(); err != nil {
		log.Fatal("start audio source", "error", err, "cmd", name)
	}

	log.Info("listening",
		"rate", cfg.SampleRate,
		"frame_ms", cfg.VadFrameMs,
		"frames", cfg.FramesPerChunk,
		"voice_threshold", cfg.VoiceThreshold,
		"silence_threshold", cfg.SilenceThreshold,
		"record_mode", cfg.RecordMode,
		"audio_source", cfg.AudioSource,
	)

	preBuffer := ring.New(cfg.PreBufferChunks)
	wasVoice := false
	silentChunks := 0
	var recording [][]byte
	var recordStart time.Time

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		<-ch
		log.Info("stopping")
		if err := src.Stop(); err != nil {
			log.Error("signal process", "error", err)
		}
		os.Exit(0)
	}()

	for {
		raw := make([]byte, proc.ChunkBytes())
		if _, err := io.ReadFull(src.Stdout(), raw); err != nil {
			if err == io.EOF {
				log.Info("audio source exited")
			} else {
				log.Error("read audio", "error", err)
			}
			break
		}

		voiceCount, err := proc.ProcessChunk(raw)
		if err != nil {
			log.Error("process chunk", "error", err)
			continue
		}
		isVoice := proc.IsVoice(voiceCount)

		if isVoice {
			silentChunks = 0
			if !wasVoice {
				log.Info("voice start")
				if cfg.RecordMode != "none" && cfg.RecordMode != "" {
					recordStart = time.Now()
					recording = preBuffer.Flush()
					recording = append(recording, raw)
					log.Info("recording started", "pre_buffer_chunks", len(recording)-1)
				}
			} else if cfg.RecordMode != "none" && cfg.RecordMode != "" {
				recording = append(recording, raw)
			}
		} else {
			silentChunks++
			if wasVoice && silentChunks >= proc.SilenceThreshold() {
				log.Info("voice end")
				if cfg.RecordMode != "none" && cfg.RecordMode != "" && recording != nil {
					if len(recording) < cfg.MinChunks {
						log.Info("recording discarded",
							"chunks", len(recording),
							"min_chunks", cfg.MinChunks,
						)
						recording = nil
					} else {
						samples := vad.MergeChunks(recording, proc.ChunkLen())
						filename := recordStart.Format("2006-01-02_15-04-05.000") + ".wav"
						duration := float64(len(samples)) / float64(cfg.SampleRate)

						log.Info("recording ready",
							"chunks", len(recording),
							"duration_s", fmt.Sprintf("%.1f", duration),
							"filename", filename,
						)

						if cfg.RecordMode == "file" || cfg.RecordMode == "both" {
							path, err := recorder.SaveFile(cfg.RecordDir, filename, samples, cfg.SampleRate)
							if err != nil {
								log.Error("save wav", "error", err, "path", path)
							} else {
								log.Info("file saved", "path", path)
							}
						}

						if (cfg.RecordMode == "https" || cfg.RecordMode == "both") && cfg.HTTPSUrl != "" {
							if err := recorder.UploadHTTPS(cfg.HTTPSUrl, samples, cfg.SampleRate,
								cfg.HTTPTimeout, cfg.TLSSkipVerify, filename); err != nil {
								log.Error("upload failed", "error", err, "url", cfg.HTTPSUrl)
							} else {
								log.Info("uploaded", "url", cfg.HTTPSUrl, "filename", filename)
							}
						}

						recording = nil
					}
				}
			}
		}

		if cfg.RecordMode != "none" && cfg.RecordMode != "" && !isVoice && recording == nil {
			preBuffer.Push(raw)
		}

		wasVoice = !(silentChunks >= proc.SilenceThreshold())
	}

	if err := src.Wait(); err != nil {
		log.Error("audio source finished", "error", err)
	}
}

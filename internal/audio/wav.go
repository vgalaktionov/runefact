package audio

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
)

// WriteWAV writes float64 samples as a PCM WAV file.
func WriteWAV(path string, samples []float64, sampleRate, bitDepth int) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating WAV file: %w", err)
	}
	defer f.Close()

	channels := 1
	bytesPerSample := bitDepth / 8
	dataSize := len(samples) * bytesPerSample
	fileSize := 36 + dataSize

	// RIFF header.
	f.Write([]byte("RIFF"))
	binary.Write(f, binary.LittleEndian, uint32(fileSize))
	f.Write([]byte("WAVE"))

	// fmt chunk.
	f.Write([]byte("fmt "))
	binary.Write(f, binary.LittleEndian, uint32(16))                                    // chunk size
	binary.Write(f, binary.LittleEndian, uint16(1))                                     // PCM format
	binary.Write(f, binary.LittleEndian, uint16(channels))                               // channels
	binary.Write(f, binary.LittleEndian, uint32(sampleRate))                             // sample rate
	binary.Write(f, binary.LittleEndian, uint32(sampleRate*channels*bytesPerSample))     // byte rate
	binary.Write(f, binary.LittleEndian, uint16(channels*bytesPerSample))                // block align
	binary.Write(f, binary.LittleEndian, uint16(bitDepth))                               // bits per sample

	// data chunk.
	f.Write([]byte("data"))
	binary.Write(f, binary.LittleEndian, uint32(dataSize))

	// Write samples.
	for _, s := range samples {
		// Clamp to [-1, 1].
		if s > 1 {
			s = 1
		} else if s < -1 {
			s = -1
		}

		switch bitDepth {
		case 8:
			// 8-bit WAV is unsigned 0-255.
			v := uint8((s + 1) / 2 * 255)
			binary.Write(f, binary.LittleEndian, v)
		case 16:
			v := int16(s * math.MaxInt16)
			binary.Write(f, binary.LittleEndian, v)
		case 24:
			v := int32(s * 8388607) // 2^23 - 1
			b := [3]byte{byte(v), byte(v >> 8), byte(v >> 16)}
			f.Write(b[:])
		}
	}

	return nil
}

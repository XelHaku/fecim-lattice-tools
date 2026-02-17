// Package utils provides utility functions for the FeCIM application.
// L02: PNG metadata embedding for screenshots.
package utils

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"image"
	"image/png"
	"io"
	"os"
	"time"
)

// PNGMetadata holds metadata to embed in PNG files.
type PNGMetadata struct {
	Title       string
	Author      string
	Description string
	Software    string
	Timestamp   time.Time
	CustomData  map[string]string
}

// SavePNGWithMetadata saves an image as PNG with embedded metadata.
// L02: Embeds tEXt chunks with simulation parameters.
func SavePNGWithMetadata(img image.Image, filename string, meta *PNGMetadata) error {
	// Encode PNG to buffer first
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return err
	}

	// PNG structure: signature + chunks
	// We need to insert tEXt chunks before IEND
	pngData := buf.Bytes()

	// Find IEND chunk (last 12 bytes: length(4) + "IEND"(4) + CRC(4))
	if len(pngData) < 20 {
		// Too small, just write as-is
		return os.WriteFile(filename, pngData, 0644)
	}

	// IEND chunk starts 12 bytes from end
	iendStart := len(pngData) - 12

	// Create output file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write everything before IEND
	if _, err := file.Write(pngData[:iendStart]); err != nil {
		return err
	}

	// Write metadata as tEXt chunks
	if meta != nil {
		if meta.Title != "" {
			writeTextChunk(file, "Title", meta.Title)
		}
		if meta.Author != "" {
			writeTextChunk(file, "Author", meta.Author)
		}
		if meta.Description != "" {
			writeTextChunk(file, "Description", meta.Description)
		}
		if meta.Software != "" {
			writeTextChunk(file, "Software", meta.Software)
		}
		if !meta.Timestamp.IsZero() {
			writeTextChunk(file, "Creation Time", meta.Timestamp.Format(time.RFC3339))
		}
		for key, value := range meta.CustomData {
			writeTextChunk(file, key, value)
		}
	}

	// Write IEND chunk
	if _, err := file.Write(pngData[iendStart:]); err != nil {
		return err
	}

	return nil
}

// writeTextChunk writes a PNG tEXt chunk.
// Format: length(4) + "tEXt"(4) + keyword + null + text + CRC(4)
func writeTextChunk(w io.Writer, keyword, text string) error {
	// tEXt data: keyword + null byte + text
	data := append([]byte(keyword), 0)
	data = append(data, []byte(text)...)

	// Calculate length
	length := uint32(len(data))

	// Write length (big-endian)
	lengthBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBytes, length)
	if _, err := w.Write(lengthBytes); err != nil {
		return err
	}

	// Chunk type + data for CRC calculation
	chunkType := []byte("tEXt")
	chunkData := append(chunkType, data...)

	// Write chunk type
	if _, err := w.Write(chunkType); err != nil {
		return err
	}

	// Write data
	if _, err := w.Write(data); err != nil {
		return err
	}

	// Calculate and write CRC
	crc := crc32.ChecksumIEEE(chunkData)
	crcBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(crcBytes, crc)
	if _, err := w.Write(crcBytes); err != nil {
		return err
	}

	return nil
}

// DefaultScreenshotMetadata creates default metadata for FeCIM screenshots.
func DefaultScreenshotMetadata(moduleName string) *PNGMetadata {
	return &PNGMetadata{
		Title:       "FeCIM Lattice Tools - " + moduleName,
		Author:      "FeCIM Lattice Tools Project",
		Software:    "FeCIM Lattice Tools v1.1 (Go 1.24/Fyne 2.7)",
		Timestamp:   time.Now(),
		Description: "Ferroelectric Compute-in-Memory simulation and hardware validation screenshot",
		CustomData: map[string]string{
			"Module":          moduleName,
			"FeCIM-Levels":    "30",
			"Project-URL":     "github.com/your-org/fecim-lattice-tools",
			"Validation-Gate": "QA-A0 Passed",
		},
	}
}

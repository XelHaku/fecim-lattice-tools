// Package mnist provides MNIST dataset loading utilities.
package mnist

import (
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// LoadMNIST loads the MNIST dataset from the specified directory.
// If train is true, loads training data; otherwise loads test data.
func LoadMNIST(dir string, train bool) (images [][]float64, labels []int, err error) {
	var imageFile, labelFile string
	if train {
		imageFile = filepath.Join(dir, "train-images-idx3-ubyte.gz")
		labelFile = filepath.Join(dir, "train-labels-idx1-ubyte.gz")
	} else {
		imageFile = filepath.Join(dir, "t10k-images-idx3-ubyte.gz")
		labelFile = filepath.Join(dir, "t10k-labels-idx1-ubyte.gz")
	}

	// Try without .gz extension if compressed files not found
	if _, err := os.Stat(imageFile); os.IsNotExist(err) {
		imageFile = imageFile[:len(imageFile)-3]
		labelFile = labelFile[:len(labelFile)-3]
	}

	// Load images
	images, err = loadImages(imageFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load images: %w", err)
	}

	// Load labels
	labels, err = loadLabels(labelFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load labels: %w", err)
	}

	if len(images) != len(labels) {
		return nil, nil, fmt.Errorf("image count (%d) != label count (%d)",
			len(images), len(labels))
	}

	return images, labels, nil
}

func loadImages(filename string) ([][]float64, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var reader io.Reader = f
	if filepath.Ext(filename) == ".gz" {
		gz, err := gzip.NewReader(f)
		if err != nil {
			return nil, err
		}
		defer gz.Close()
		reader = gz
	}

	// Read header
	var magic, numImages, numRows, numCols int32
	binary.Read(reader, binary.BigEndian, &magic)
	binary.Read(reader, binary.BigEndian, &numImages)
	binary.Read(reader, binary.BigEndian, &numRows)
	binary.Read(reader, binary.BigEndian, &numCols)

	if magic != 2051 {
		return nil, fmt.Errorf("invalid magic number: %d (expected 2051)", magic)
	}

	if numRows != 28 || numCols != 28 {
		return nil, fmt.Errorf("unexpected image size: %dx%d", numRows, numCols)
	}

	// Read images
	images := make([][]float64, numImages)
	pixelCount := int(numRows * numCols)
	buf := make([]byte, pixelCount)

	for i := int32(0); i < numImages; i++ {
		_, err := io.ReadFull(reader, buf)
		if err != nil {
			return nil, fmt.Errorf("failed to read image %d: %w", i, err)
		}

		img := make([]float64, pixelCount)
		for j := 0; j < pixelCount; j++ {
			img[j] = float64(buf[j]) / 255.0
		}
		images[i] = img
	}

	return images, nil
}

func loadLabels(filename string) ([]int, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var reader io.Reader = f
	if filepath.Ext(filename) == ".gz" {
		gz, err := gzip.NewReader(f)
		if err != nil {
			return nil, err
		}
		defer gz.Close()
		reader = gz
	}

	// Read header
	var magic, numLabels int32
	binary.Read(reader, binary.BigEndian, &magic)
	binary.Read(reader, binary.BigEndian, &numLabels)

	if magic != 2049 {
		return nil, fmt.Errorf("invalid magic number: %d (expected 2049)", magic)
	}

	// Read labels
	labels := make([]int, numLabels)
	buf := make([]byte, numLabels)
	_, err = io.ReadFull(reader, buf)
	if err != nil {
		return nil, err
	}

	for i := range labels {
		labels[i] = int(buf[i])
	}

	return labels, nil
}

// Package data provides dataset loading utilities for neural network demos.
package data

import (
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const (
	// MNIST dataset URLs
	mnistBaseURL    = "http://yann.lecun.com/exdb/mnist/"
	trainImagesFile = "train-images-idx3-ubyte.gz"
	trainLabelsFile = "train-labels-idx1-ubyte.gz"
	testImagesFile  = "t10k-images-idx3-ubyte.gz"
	testLabelsFile  = "t10k-labels-idx1-ubyte.gz"

	// Image dimensions
	MNISTImageSize = 28
	MNISTPixels    = MNISTImageSize * MNISTImageSize // 784
	MNISTClasses   = 10
)

// MNISTImage represents a single MNIST image.
type MNISTImage struct {
	Pixels []float64 // Normalized pixel values (0-1)
	Label  int       // Digit label (0-9)
}

// MNISTDataset holds the complete MNIST dataset.
type MNISTDataset struct {
	TrainImages []MNISTImage
	TestImages  []MNISTImage
	DataDir     string
}

// NewMNISTDataset creates a new dataset loader.
func NewMNISTDataset(dataDir string) *MNISTDataset {
	return &MNISTDataset{
		DataDir: dataDir,
	}
}

// Download downloads the MNIST dataset files.
func (m *MNISTDataset) Download() error {
	// Create data directory
	if err := os.MkdirAll(m.DataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	files := []string{trainImagesFile, trainLabelsFile, testImagesFile, testLabelsFile}

	for _, filename := range files {
		localPath := filepath.Join(m.DataDir, filename)

		// Skip if already exists
		if _, err := os.Stat(localPath); err == nil {
			fmt.Printf("  %s already exists, skipping\n", filename)
			continue
		}

		// Download file
		url := mnistBaseURL + filename
		fmt.Printf("  Downloading %s...\n", filename)

		if err := downloadFile(localPath, url); err != nil {
			return fmt.Errorf("failed to download %s: %w", filename, err)
		}
	}

	return nil
}

// Load loads the MNIST dataset from disk.
func (m *MNISTDataset) Load() error {
	var err error

	// Load training data
	fmt.Println("Loading training data...")
	m.TrainImages, err = m.loadImageSet(
		filepath.Join(m.DataDir, trainImagesFile),
		filepath.Join(m.DataDir, trainLabelsFile),
	)
	if err != nil {
		return fmt.Errorf("failed to load training data: %w", err)
	}

	// Load test data
	fmt.Println("Loading test data...")
	m.TestImages, err = m.loadImageSet(
		filepath.Join(m.DataDir, testImagesFile),
		filepath.Join(m.DataDir, testLabelsFile),
	)
	if err != nil {
		return fmt.Errorf("failed to load test data: %w", err)
	}

	fmt.Printf("Loaded %d training and %d test images\n",
		len(m.TrainImages), len(m.TestImages))

	return nil
}

// loadImageSet loads images and labels from gzipped IDX files.
func (m *MNISTDataset) loadImageSet(imagePath, labelPath string) ([]MNISTImage, error) {
	// Load images
	imageFile, err := os.Open(imagePath)
	if err != nil {
		return nil, err
	}
	defer imageFile.Close()

	imageGz, err := gzip.NewReader(imageFile)
	if err != nil {
		return nil, err
	}
	defer imageGz.Close()

	// Read image header
	var imageMagic, numImages, rows, cols int32
	binary.Read(imageGz, binary.BigEndian, &imageMagic)
	binary.Read(imageGz, binary.BigEndian, &numImages)
	binary.Read(imageGz, binary.BigEndian, &rows)
	binary.Read(imageGz, binary.BigEndian, &cols)

	if imageMagic != 2051 {
		return nil, fmt.Errorf("invalid image magic number: %d", imageMagic)
	}

	// Load labels
	labelFile, err := os.Open(labelPath)
	if err != nil {
		return nil, err
	}
	defer labelFile.Close()

	labelGz, err := gzip.NewReader(labelFile)
	if err != nil {
		return nil, err
	}
	defer labelGz.Close()

	// Read label header
	var labelMagic, numLabels int32
	binary.Read(labelGz, binary.BigEndian, &labelMagic)
	binary.Read(labelGz, binary.BigEndian, &numLabels)

	if labelMagic != 2049 {
		return nil, fmt.Errorf("invalid label magic number: %d", labelMagic)
	}

	if numImages != numLabels {
		return nil, fmt.Errorf("image/label count mismatch: %d vs %d", numImages, numLabels)
	}

	// Read all images and labels
	images := make([]MNISTImage, numImages)
	pixelBuf := make([]byte, MNISTPixels)
	labelBuf := make([]byte, 1)

	for i := int32(0); i < numImages; i++ {
		// Read pixels
		if _, err := io.ReadFull(imageGz, pixelBuf); err != nil {
			return nil, fmt.Errorf("failed to read image %d: %w", i, err)
		}

		// Read label
		if _, err := io.ReadFull(labelGz, labelBuf); err != nil {
			return nil, fmt.Errorf("failed to read label %d: %w", i, err)
		}

		// Convert to normalized float64
		pixels := make([]float64, MNISTPixels)
		for j, p := range pixelBuf {
			pixels[j] = float64(p) / 255.0
		}

		images[i] = MNISTImage{
			Pixels: pixels,
			Label:  int(labelBuf[0]),
		}
	}

	return images, nil
}

// GetBatch returns a batch of images starting at the given index.
func (m *MNISTDataset) GetBatch(start, batchSize int, train bool) []MNISTImage {
	var source []MNISTImage
	if train {
		source = m.TrainImages
	} else {
		source = m.TestImages
	}

	end := start + batchSize
	if end > len(source) {
		end = len(source)
	}

	return source[start:end]
}

// ToOneHot converts a label to one-hot encoded vector.
func ToOneHot(label int) []float64 {
	oneHot := make([]float64, MNISTClasses)
	if label >= 0 && label < MNISTClasses {
		oneHot[label] = 1.0
	}
	return oneHot
}

// ArgMax returns the index of the maximum value in a slice.
func ArgMax(values []float64) int {
	maxIdx := 0
	maxVal := values[0]
	for i, v := range values {
		if v > maxVal {
			maxVal = v
			maxIdx = i
		}
	}
	return maxIdx
}

// downloadFile downloads a file from URL to local path.
func downloadFile(filepath, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

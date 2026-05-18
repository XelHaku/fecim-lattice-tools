package mnist

type MNISTState struct {
	Accuracy      float64   `json:"accuracy"`
	NumLevels     int       `json:"num_levels"`
	TotalImages   int       `json:"total_images"`
	CorrectImages int       `json:"correct_images"`
	SweepLevels   []int     `json:"sweep_levels"`
	SweepAccuracy []float64 `json:"sweep_accuracy"`
}

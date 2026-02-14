package crossbar

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
)

type kchDerived struct {
	IsumRow   [][]float64
	IsumCol   [][]float64
	VdropsRow [][]float64
	VdropsCol [][]float64
}

func deriveKCH(r *ParasiticMVMResult, rpRow, rpCol float64) *kchDerived {
	rows := len(r.DeviceCurrents)
	cols := len(r.DeviceCurrents[0])

	IsumCol := make([][]float64, rows)
	IsumRow := make([][]float64, rows)
	VdropsCol := make([][]float64, rows)
	VdropsRow := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		IsumCol[i] = make([]float64, cols)
		IsumRow[i] = make([]float64, cols)
		VdropsCol[i] = make([]float64, cols)
		VdropsRow[i] = make([]float64, cols)
	}

	for j := 0; j < cols; j++ {
		IsumCol[0][j] = r.DeviceCurrents[0][j]
		for i := 1; i < rows; i++ {
			IsumCol[i][j] = IsumCol[i-1][j] + r.DeviceCurrents[i][j]
		}
	}

	for i := 0; i < rows; i++ {
		IsumRow[i][cols-1] = r.DeviceCurrents[i][cols-1]
		for j := cols - 2; j >= 0; j-- {
			IsumRow[i][j] = IsumRow[i][j+1] + r.DeviceCurrents[i][j]
		}
	}

	for j := 0; j < cols; j++ {
		VdropsCol[0][j] = 0
		for i := 1; i < rows; i++ {
			VdropsCol[i][j] = VdropsCol[i-1][j] + rpCol*IsumCol[i-1][j]
		}
	}

	for i := 0; i < rows; i++ {
		VdropsRow[i][0] = 0
		for j := 1; j < cols; j++ {
			VdropsRow[i][j] = VdropsRow[i][j-1] + rpRow*IsumRow[i][j]
		}
	}

	return &kchDerived{IsumRow: IsumRow, IsumCol: IsumCol, VdropsRow: VdropsRow, VdropsCol: VdropsCol}
}

func TestSolverKirchhoff_M2KCH01_KCLResidual(t *testing.T) {
	sizes := []int{2, 4, 8, 16, 32}
	rng := rand.New(rand.NewSource(1))

	const (
		gUniform = 50e-6
		rpSeg    = 2.5
		vIn      = 0.2
	)

	for _, n := range sizes {
		t.Run(fmt.Sprintf("N=%dx%d/uniform", n, n), func(t *testing.T) {
			cfg := DefaultSORConfig()
			s, err := NewParasiticSolver(n, n, cfg)
			if err != nil {
				t.Fatalf("NewParasiticSolver: %v", err)
			}
			s.SetParasitics(rpSeg, rpSeg)
			s.SetConductances(makeUniformG(n, n, gUniform))

			res, err := s.SolveMVM(makeApplied(n, vIn))
			if err != nil && !res.Converged {
				t.Fatalf("SolveMVM: %v (iters=%d maxErr=%g)", err, res.Iterations, res.MaxError)
			}

			d := deriveKCH(res, rpSeg, rpSeg)

			maxKCL := 0.0

			for i := 0; i < n; i++ {
				for j := 0; j < n-1; j++ {
					r := d.IsumRow[i][j] - d.IsumRow[i][j+1] - res.DeviceCurrents[i][j]
					if a := math.Abs(r); a > maxKCL {
						maxKCL = a
					}
				}
				r := d.IsumRow[i][n-1] - res.DeviceCurrents[i][n-1]
				if a := math.Abs(r); a > maxKCL {
					maxKCL = a
				}
			}
			for j := 0; j < n; j++ {
				r0 := d.IsumCol[0][j] - res.DeviceCurrents[0][j]
				if a := math.Abs(r0); a > maxKCL {
					maxKCL = a
				}
				for i := 1; i < n; i++ {
					r := d.IsumCol[i][j] - d.IsumCol[i-1][j] - res.DeviceCurrents[i][j]
					if a := math.Abs(r); a > maxKCL {
						maxKCL = a
					}
				}
			}

			t.Logf("max KCL residual = %.3e A (iters=%d, converged=%v, maxVerr=%.3e V)", maxKCL, res.Iterations, res.Converged, res.MaxError)
			if maxKCL >= 1e-10 {
				t.Fatalf("KCL residual too large: %.3e A (want < 1e-10)", maxKCL)
			}
		})

		t.Run(fmt.Sprintf("N=%dx%d/random", n, n), func(t *testing.T) {
			cfg := DefaultSORConfig()
			s, err := NewParasiticSolver(n, n, cfg)
			if err != nil {
				t.Fatalf("NewParasiticSolver: %v", err)
			}
			s.SetParasitics(rpSeg, rpSeg)
			s.SetConductances(makeRandomG(n, n, 10e-6, 100e-6, rng))

			res, err := s.SolveMVM(makeApplied(n, vIn))
			if err != nil && !res.Converged {
				t.Fatalf("SolveMVM: %v (iters=%d maxErr=%g)", err, res.Iterations, res.MaxError)
			}

			d := deriveKCH(res, rpSeg, rpSeg)
			maxKCL := 0.0
			for i := 0; i < n; i++ {
				for j := 0; j < n-1; j++ {
					r := d.IsumRow[i][j] - d.IsumRow[i][j+1] - res.DeviceCurrents[i][j]
					if a := math.Abs(r); a > maxKCL {
						maxKCL = a
					}
				}
				r := d.IsumRow[i][n-1] - res.DeviceCurrents[i][n-1]
				if a := math.Abs(r); a > maxKCL {
					maxKCL = a
				}
			}
			for j := 0; j < n; j++ {
				r0 := d.IsumCol[0][j] - res.DeviceCurrents[0][j]
				if a := math.Abs(r0); a > maxKCL {
					maxKCL = a
				}
				for i := 1; i < n; i++ {
					r := d.IsumCol[i][j] - d.IsumCol[i-1][j] - res.DeviceCurrents[i][j]
					if a := math.Abs(r); a > maxKCL {
						maxKCL = a
					}
				}
			}

			t.Logf("max KCL residual = %.3e A (iters=%d, converged=%v, maxVerr=%.3e V)", maxKCL, res.Iterations, res.Converged, res.MaxError)
			if maxKCL >= 1e-10 {
				t.Fatalf("KCL residual too large: %.3e A (want < 1e-10)", maxKCL)
			}
		})
	}
}

func TestSolverKirchhoff_M2KCH02_KVLLoop(t *testing.T) {
	sizes := []int{2, 4, 8, 16, 32}
	rng := rand.New(rand.NewSource(2))

	const (
		gUniform = 50e-6
		rpSeg    = 2.5
		vIn      = 0.2
		kvlTol   = 1e-6
	)

	cases := []struct {
		name string
		g    func(n int) [][]float64
	}{
		{"uniform", func(n int) [][]float64 { return makeUniformG(n, n, gUniform) }},
		{"random", func(n int) [][]float64 { return makeRandomG(n, n, 10e-6, 100e-6, rng) }},
	}

	for _, n := range sizes {
		for _, tc := range cases {
			t.Run(fmt.Sprintf("N=%dx%d/%s", n, n, tc.name), func(t *testing.T) {
				cfg := DefaultSORConfig()
				s, err := NewParasiticSolver(n, n, cfg)
				if err != nil {
					t.Fatalf("NewParasiticSolver: %v", err)
				}
				s.SetParasitics(rpSeg, rpSeg)
				s.SetConductances(tc.g(n))

				applied := makeApplied(n, vIn)
				res, err := s.SolveMVM(applied)
				if err != nil && !res.Converged {
					t.Fatalf("SolveMVM: %v (iters=%d maxErr=%g)", err, res.Iterations, res.MaxError)
				}

				d := deriveKCH(res, rpSeg, rpSeg)
				maxKVL := 0.0
				for i := 0; i < n; i++ {
					for j := 0; j < n; j++ {
						r := applied[j] - (d.VdropsRow[i][j] + d.VdropsCol[i][j] + res.DeviceVoltages[i][j])
						if a := math.Abs(r); a > maxKVL {
							maxKVL = a
						}
					}
				}
				t.Logf("max KVL residual = %.3e V (iters=%d, converged=%v, maxVerr=%.3e V)", maxKVL, res.Iterations, res.Converged, res.MaxError)
				if maxKVL >= kvlTol {
					t.Fatalf("KVL residual too large: %.3e V (want < %.1e)", maxKVL, kvlTol)
				}
			})
		}
	}
}

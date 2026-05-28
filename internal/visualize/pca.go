package visualize

import (
	"fmt"

	"gonum.org/v1/gonum/mat"
)

type PCAModel struct {
	Components *mat.Dense
	Mean       []float64
}

func (m *PCAModel) Project(vec []float32) ([3]float64, error) {
	if m.Components == nil {
		return [3]float64{}, fmt.Errorf("PCA model not initialized")
	}

	numDims := len(vec)
	if numDims != len(m.Mean) {
		return [3]float64{}, fmt.Errorf("dimension mismatch: expected %d, got %d", len(m.Mean), numDims)
	}

	centered := make([]float64, numDims)
	for i, val := range vec {
		centered[i] = float64(val) - m.Mean[i]
	}

	v := mat.NewDense(1, numDims, centered)
	var projected mat.Dense
	projected.Mul(v, m.Components)

	var result [3]float64
	cols := m.Components.RawMatrix().Cols
	for i := 0; i < cols && i < 3; i++ {
		result[i] = projected.At(0, i)
	}

	return result, nil
}

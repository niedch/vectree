package visualize

import (
	"fmt"

	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat"
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

func ReduceTo3D(embeddings [][]float32) ([][3]float64, *PCAModel, error) {
	if len(embeddings) == 0 {
		return nil, nil, nil
	}

	numRecords := len(embeddings)
	numDims := len(embeddings[0])

	mean := make([]float64, numDims)
	data := make([]float64, numRecords*numDims)
	for i, record := range embeddings {
		for j, val := range record {
			fval := float64(val)
			data[i*numDims+j] = fval
			mean[j] += fval
		}
	}
	for j := range mean {
		mean[j] /= float64(numRecords)
	}

	for i := 0; i < numRecords; i++ {
		for j := 0; j < numDims; j++ {
			data[i*numDims+j] -= mean[j]
		}
	}

	m := mat.NewDense(numRecords, numDims, data)

	var pc stat.PC
	ok := pc.PrincipalComponents(m, nil)
	if !ok {
		return nil, nil, fmt.Errorf("PCA failed")
	}

	k := 3
	if numDims < 3 {
		k = numDims
	}

	var allComponents mat.Dense
	pc.VectorsTo(&allComponents)

	_, availComponents := allComponents.Dims()
	if k > availComponents {
		k = availComponents
	}

	var topK mat.Dense
	topK.CloneFrom(allComponents.Slice(0, numDims, 0, k))

	var reduced mat.Dense
	reduced.Mul(m, &topK)

	results := make([][3]float64, numRecords)
	for i := 0; i < numRecords; i++ {
		var point [3]float64
		for j := 0; j < k; j++ {
			point[j] = reduced.At(i, j)
		}
		results[i] = point
	}

	model := &PCAModel{
		Components: &topK,
		Mean:       mean,
	}

	return results, model, nil
}

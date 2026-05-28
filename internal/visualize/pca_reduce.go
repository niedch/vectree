package visualize

import (
	"fmt"

	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat"
)

func ReduceTo3D(embeddings [][]float32) ([][3]float64, *PCAModel, error) {
	if len(embeddings) == 0 {
		return nil, nil, nil
	}

	numRecords := len(embeddings)
	numDims := len(embeddings[0])

	mean := computeMean(embeddings, numRecords, numDims)

	data := make([]float64, numRecords*numDims)
	for i, record := range embeddings {
		for j, val := range record {
			data[i*numDims+j] = float64(val)
		}
	}

	centerData(data, mean, numRecords, numDims)

	m := mat.NewDense(numRecords, numDims, data)

	var pc stat.PC
	ok := pc.PrincipalComponents(m, nil)
	if !ok {
		return nil, nil, fmt.Errorf("PCA failed")
	}

	k := min(numDims, 3)

	var allComponents mat.Dense
	pc.VectorsTo(&allComponents)

	_, availComponents := allComponents.Dims()
	if k > availComponents {
		k = availComponents
	}

	if k == 0 {
		return nil, nil, fmt.Errorf("no principal components available")
	}

	var topK mat.Dense
	topK.CloneFrom(allComponents.Slice(0, numDims, 0, k))

	var reduced mat.Dense
	reduced.Mul(m, &topK)

	results := extractPoints(&reduced, numRecords, k)

	model := &PCAModel{
		Components: &topK,
		Mean:       mean,
	}

	return results, model, nil
}

func computeMean(embeddings [][]float32, numRecords, numDims int) []float64 {
	mean := make([]float64, numDims)

	for i := range numRecords {
		for j := range numDims {
			mean[j] += float64(embeddings[i][j])
		}
	}

	for j := range mean {
		mean[j] /= float64(numRecords)
	}

	return mean
}

func centerData(data []float64, mean []float64, numRecords, numDims int) {
	for i := range numRecords {
		for j := range numDims {
			data[i*numDims+j] -= mean[j]
		}
	}
}

func extractPoints(reduced *mat.Dense, numRecords, k int) [][3]float64 {
	results := make([][3]float64, numRecords)
	for i := range numRecords {
		var point [3]float64
		for j := range k {
			point[j] = reduced.At(i, j)
		}
		results[i] = point
	}
	return results
}

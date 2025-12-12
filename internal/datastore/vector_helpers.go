package datastore

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"
)

// VectorSearchResult represents a result from a vector similarity search
type VectorSearchResult struct {
	Rowid    int64
	Distance float64
}

// SerializeFloat32Vector converts a slice of float32 values into a byte slice
// that can be stored in the embeddings table
func SerializeFloat32Vector(vector []float32) []byte {
	buf := make([]byte, len(vector)*4)
	for i, v := range vector {
		binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(v))
	}
	return buf
}

// DeserializeFloat32Vector converts a byte slice back into a slice of float32 values
func DeserializeFloat32Vector(data []byte) ([]float32, error) {
	if len(data)%4 != 0 {
		return nil, fmt.Errorf("invalid vector data length: %d", len(data))
	}
	vector := make([]float32, len(data)/4)
	for i := range vector {
		bits := binary.LittleEndian.Uint32(data[i*4:])
		vector[i] = math.Float32frombits(bits)
	}
	return vector, nil
}

// SearchSimilarVectors performs a vector similarity search using sqlite-vec
// This uses raw SQL because sqlite-vec's MATCH syntax is not supported by sqlc
func (q *Queries) SearchSimilarVectors(ctx context.Context, queryVector []byte, k int) ([]VectorSearchResult, error) {
	// sqlite-vec uses the MATCH operator for vector search
	// The query vector is passed as a blob, and k specifies the number of results
	query := `
		SELECT 
			rowid,
			distance
		FROM embeddings
		WHERE embedding MATCH ?
		ORDER BY distance
		LIMIT ?
	`

	rows, err := q.db.QueryContext(ctx, query, queryVector, k)
	if err != nil {
		return nil, fmt.Errorf("failed to search vectors: %w", err)
	}
	defer rows.Close()

	var results []VectorSearchResult
	for rows.Next() {
		var result VectorSearchResult
		if err := rows.Scan(&result.Rowid, &result.Distance); err != nil {
			return nil, fmt.Errorf("failed to scan result: %w", err)
		}
		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating results: %w", err)
	}

	return results, nil
}

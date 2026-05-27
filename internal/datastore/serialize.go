package datastore

import (
	"encoding/binary"
	"fmt"
	"math"
)

func SerializeFloat32(v []float32) ([]byte, error) {
	buf := make([]byte, len(v)*4)
	for i, f := range v {
		binary.LittleEndian.PutUint32(buf[i*4:(i+1)*4], math.Float32bits(f))
	}
	return buf, nil
}

func deserializeFloat32(data []byte) ([]float32, error) {
	if len(data)%4 != 0 {
		return nil, fmt.Errorf("invalid data length: %d (must be multiple of 4)", len(data))
	}

	count := len(data) / 4
	result := make([]float32, count)
	for i := range count {
		bits := binary.LittleEndian.Uint32(data[i*4 : (i+1)*4])
		result[i] = math.Float32frombits(bits)
	}
	return result, nil
}

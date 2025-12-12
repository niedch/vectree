package stages

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBatcher(t *testing.T) {
	size := 10
	b := NewBatcher[string](size)
	assert.Equal(t, size, b.Size, "NewBatcher() size")
}

func TestBatcher_Run(t *testing.T) {
	tests := []struct {
		name      string
		batchSize int
		inputs    []string
		want      [][]string
	}{
		{
			name:      "exact batches",
			batchSize: 2,
			inputs:    []string{"a", "b", "c", "d"},
			want:      [][]string{{"a", "b"}, {"c", "d"}},
		},
		{
			name:      "incomplete final batch",
			batchSize: 3,
			inputs:    []string{"a", "b", "c", "d", "e"},
			want:      [][]string{{"a", "b", "c"}, {"d", "e"}},
		},
		{
			name:      "no input",
			batchSize: 2,
			inputs:    []string{},
			want:      nil,
		},
		{
			name:      "single full batch",
			batchSize: 3,
			inputs:    []string{"a", "b", "c"},
			want:      [][]string{{"a", "b", "c"}},
		},
		{
			name:      "single partial batch",
			batchSize: 3,
			inputs:    []string{"a", "b"},
			want:      [][]string{{"a", "b"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBatcher[string](tt.batchSize)
			in := make(chan string)
			out := b.Run(context.Background(), in)

			go func() {
				defer close(in)
				for _, item := range tt.inputs {
					in <- item
				}
			}()

			var got [][]string
			for batch := range out {
				got = append(got, batch)
			}

			assert.Equal(t, tt.want, got, "Batcher.Run()")
		})
	}
}

func TestBatcher_Run_DefaultSize(t *testing.T) {
	b := NewBatcher[string](0)
	in := make(chan string, 33)
	out := b.Run(context.Background(), in)

	for i := 0; i < 33; i++ {
		in <- "a"
	}
	close(in)

	count := 0
	for batch := range out {
		count++
		if count == 1 {
			assert.Len(t, batch, 32, "first batch size")
		}
		if count == 2 {
			assert.Len(t, batch, 1, "second batch size")
		}
	}
	assert.Equal(t, 2, count, "expected 2 batches")
}

func TestBatcher_Run_ContextCancellation(t *testing.T) {
	b := NewBatcher[string](2)
	in := make(chan string)
	ctx, cancel := context.WithCancel(context.Background())
	out := b.Run(ctx, in)

	go func() {
		defer close(in)
		for i := 0; i < 5; i++ {
			select {
			case in <- "a":
			case <-ctx.Done():
				return
			}
		}
	}()

	// Read one batch
	d, ok := <-out
	assert.True(t, ok, "expected to read a batch")
	assert.Len(t, d, 2, "expected batch of size 2")

	cancel()

	// After cancellation, the output channel should be closed.
	// Depending on timing, there might be one more batch in the pipe.
	for range out {
	}

	// Check that it is actually closed
	_, ok = <-out
	assert.False(t, ok, "output channel should be closed after context cancellation")
}
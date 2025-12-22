package stages

import (
	"context"
)

type Batcher[T any] struct {
	Size int
}

func NewBatcher[T any](size int) *Batcher[T] {
	return &Batcher[T]{
		Size: size,
	}
}

func (b Batcher[T]) Run(ctx context.Context, in <-chan T) <-chan []T {
	if b.Size <= 0 {
		b.Size = 32
	}

	out := make(chan []T)
	go func() {
		defer close(out)

		batch := make([]T, 0, b.Size)
		flush := func() bool {
			if len(batch) == 0 {
				return true
			}

			select {
			case out <- batch:
				batch = make([]T, 0, b.Size)
				return true
			case <-ctx.Done():
				return false
			}
		}

		for msg := range in {
			chunk := msg
			batch = append(batch, chunk)
			if len(batch) >= b.Size && !flush() {
				return
			}
		}

		_ = flush()
	}()

	return out
}

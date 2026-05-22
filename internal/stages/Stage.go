package stages

import (
	"context"
	"sync"
)

type Stage[IN any, OUT any] interface {
	Run(ctx context.Context, in <-chan IN) <-chan OUT
}

type UntypedStage struct {
	Runner func(ctx context.Context, in <-chan any) <-chan any
}

func (u UntypedStage) Run(ctx context.Context, in <-chan any) <-chan any {
	return u.Runner(ctx, in)
}

// WorkerPoolStage processes items from the input channel using a fixed pool of workers.
// The processor function can produce zero or more output items for each input item.
func WorkerPoolStage[IN any, OUT any](ctx context.Context, in <-chan IN, workers int, processor func(context.Context, IN, chan<- OUT) error) <-chan OUT {
	if workers <= 0 {
		workers = 1
	}

	out := make(chan OUT)
	var wg sync.WaitGroup
	wg.Add(workers)

	worker := func() {
		defer wg.Done()

		for item := range in {
			if err := processor(ctx, item, out); err != nil {
				continue
			}
		}
	}

	for i := 0; i < workers; i++ {
		go worker()
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

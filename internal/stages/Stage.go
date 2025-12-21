package stages

import (
	"context"
	"sync"
)

type Stage[IN any, OUT any] interface {
	Run(ctx context.Context, in <-chan IN) <-chan OUT
}

// ParallelStage processes items from the input channel concurrently using the provided processor function
func ParallelStage[IN any, OUT any](ctx context.Context, in <-chan IN, concurrency int, processor func(context.Context, IN) (OUT, bool)) <-chan OUT {
	out := make(chan OUT)

	go func() {
		defer close(out)

		sem := make(chan struct{}, concurrency)

		done := make(chan struct{})
		activeCount := 0

		for item := range in {
			activeCount++

			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				return
			}

			go func(item IN) {
				defer func() {
					<-sem
					done <- struct{}{}
				}()

				result, ok := processor(ctx, item)

				if !ok {
					return
				}

				select {
				case out <- result:
				case <-ctx.Done():
					return
				}
			}(item)
		}

		for i := 0; i < activeCount; i++ {
			<-done
		}
	}()

	return out
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

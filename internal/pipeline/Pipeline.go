package pipeline

import (
	"context"

	"broadcom.com/vertex-ingestor/internal/stages"
)

type Pipeline[OUT any] struct {
	runner func(ctx context.Context, firstIn <-chan any) <-chan OUT
}

func New[OUT any](stage stages.Stage[any, OUT]) *Pipeline[OUT] {
	return &Pipeline[OUT]{
		runner: func(ctx context.Context, firstIn <-chan any) <-chan OUT {
			return stage.Run(ctx, firstIn)
		},
	}
}

func AddStage[PREV_OUT any, NEW_OUT any](p *Pipeline[PREV_OUT], stage stages.Stage[PREV_OUT, NEW_OUT]) *Pipeline[NEW_OUT] {
	return &Pipeline[NEW_OUT]{
		runner: func(ctx context.Context, firstIn <-chan any) <-chan NEW_OUT {
			// Get the output channel from the previously constructed pipeline part.
			prevOutChan := p.runner(ctx, firstIn)
			return stage.Run(ctx, prevOutChan)
		},
	}
}

func (p *Pipeline[OUT]) Run(ctx context.Context) {
	in := make(chan any)
	close(in)

	finalOut := p.runner(ctx, in)
	for range finalOut {
	}
}

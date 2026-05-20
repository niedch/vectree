package pipeline

import (
	"context"
	"sync"

	"github.com/niedch/tree-rag/internal/stages"
)

// MergeStage merges the outputs of two pipelines into a single output channel.
// It reads from both input channels concurrently and forwards all items to the output.
// The output channel is closed when both input channels are closed.
type MergeStage struct{}

func NewMergeStage() *MergeStage {
	return &MergeStage{}
}

// Run implements the Stage interface for MergeStage.
// It takes a channel of channels (where each item is a channel to merge)
// and returns a single merged output channel.
func (m *MergeStage) Run(ctx context.Context, in <-chan any) <-chan any {
	out := make(chan any)

	go func() {
		defer close(out)

		for item := range in {
			// Each item should be a channel
			if ch, ok := item.(<-chan any); ok {
				for v := range ch {
					select {
					case out <- v:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return out
}

// Merge merges two channels into a single output channel.
// This is a helper function that can be used independently of the Stage interface.
func Merge[T any](ctx context.Context, ch1, ch2 <-chan T) <-chan T {
	return MergeAll(ctx, ch1, ch2)
}

// MergeAll merges multiple channels into a single output channel.
func MergeAll[T any](ctx context.Context, channels ...<-chan T) <-chan T {
	out := make(chan T)

	go func() {
		defer close(out)

		var wg sync.WaitGroup
		for _, ch := range channels {
			wg.Add(1)
			go func(ch <-chan T) {
				defer wg.Done()
				for v := range ch {
					select {
					case out <- v:
					case <-ctx.Done():
						return
					}
				}
			}(ch)
		}

		wg.Wait()
	}()

	return out
}

// MergePipelines creates a stage that merges the outputs of multiple pipelines.
func MergePipelines(pipelines []*Pipeline) stages.UntypedStage {
	return stages.UntypedStage{
		Runner: func(ctx context.Context, in <-chan any) <-chan any {
			var outs []<-chan any
			for _, p := range pipelines {
				outs = append(outs, p.Run(ctx))
			}
			return MergeAll(ctx, outs...)
		},
	}
}

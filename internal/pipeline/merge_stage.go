package pipeline

import (
	"context"

	"broadcom.com/vertex-ingestor/internal/stages"
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
	out := make(chan T)

	go func() {
		defer close(out)

		// Use a done channel to track when both inputs are exhausted
		done := make(chan struct{}, 2)

		// Forward from first channel
		go func() {
			defer func() { done <- struct{}{} }()
			for v := range ch1 {
				select {
				case out <- v:
				case <-ctx.Done():
					return
				}
			}
		}()

		// Forward from second channel
		go func() {
			defer func() { done <- struct{}{} }()
			for v := range ch2 {
				select {
				case out <- v:
				case <-ctx.Done():
					return
				}
			}
		}()

		// Wait for both channels to be exhausted
		<-done
		<-done
	}()

	return out
}

// MergePipelines creates a stage that merges the outputs of two pipelines.
func MergePipelines(pipeline1, pipeline2 *Pipeline) stages.UntypedStage {
	return stages.UntypedStage{
		Runner: func(ctx context.Context, in <-chan any) <-chan any {
			// Run both pipelines
			out1 := pipeline1.Run(ctx)
			out2 := pipeline2.Run(ctx)

			// Merge their outputs
			return Merge(ctx, out1, out2)
		},
	}
}

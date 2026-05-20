package pipeline

import (
	"context"

	"github.com/niedch/tree-rag/internal/stages"
)

func TypedStage[I any, O any](stage stages.Stage[I, O]) stages.UntypedStage {
	return stages.UntypedStage{
		Runner: func(ctx context.Context, in <-chan any) <-chan any {
			inTyped := make(chan I)

			// Convert untyped input to typed input
			go func() {
				defer close(inTyped)
				for v := range in {
					inTyped <- v.(I)
				}
			}()

			outTyped := stage.Run(ctx, inTyped)

			// Convert typed output to untyped output
			out := make(chan any)
			go func() {
				defer close(out)
				for v := range outTyped {
					out <- v
				}
			}()

			return out
		},
	}
}


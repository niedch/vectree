package stages

import "context"

type Stage[IN any, OUT any] interface {
	Run(ctx context.Context, in <-chan IN) <-chan OUT
}

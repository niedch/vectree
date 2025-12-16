package stages

import (
	"context"
	"fmt"
	"log"
)

type DebugStage struct{}

func NewDebugStage() *DebugStage {
	return &DebugStage{}
}

func (l DebugStage) Run(ctx context.Context, in <-chan string) <-chan string {
	out := make(chan string)

	go func() {
		defer close(out)

		for str := range in {
			log.Println(str)

			select {
			case out <- str:
			case <-ctx.Done():
				return
			}
		}
	}()

	return out
}

type DebugGenericStage[T any] struct {
	name string
}

func NewDebugGenericStage[T any](name string) *DebugGenericStage[T] {
	return &DebugGenericStage[T]{name: name}
}

func (l DebugGenericStage[T]) Run(ctx context.Context, in <-chan T) <-chan T {
	out := make(chan T)

	go func() {
		defer close(out)

		count := 0
		for item := range in {
			count++
			log.Printf("[%s] Item %d: %v", l.name, count, fmt.Sprintf("%+v", item))

			select {
			case out <- item:
			case <-ctx.Done():
				return
			}
		}
		log.Printf("[%s] Total items: %d", l.name, count)
	}()

	return out
}

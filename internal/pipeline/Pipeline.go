package pipeline

import (
	"context"

	"broadcom.com/vertex-ingestor/internal/stages"
)

type Pipeline struct {
	stages []stages.Stage[any, any]
}

func NewPipeline() *Pipeline {
	return &Pipeline{}
}

func (p *Pipeline) AddStage(stage stages.Stage[any, any]) *Pipeline {
	p.stages = append(p.stages, stage)
	return p
}

func (p *Pipeline) Run(ctx context.Context) <-chan any {
	// Create and close initial empty channel to start the pipeline
	initialCh := make(chan any)
	close(initialCh)

	var ch <-chan any = initialCh
	for _, stage := range p.stages {
		ch = stage.Run(ctx, ch)
	}

	return ch
}


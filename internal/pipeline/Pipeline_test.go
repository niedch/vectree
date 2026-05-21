package pipeline_test

import (
	"context"
	"testing"

	"github.com/niedch/tree-rag/internal/pipeline"
	"github.com/niedch/tree-rag/internal/stages"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewAndRunFirstStage(t *testing.T) {
	ctx := context.Background()
	stage1 := stages.NewMockStage[any, int](t)
	var stage1Ran bool

	outCh := make(chan int)
	close(outCh)

	stage1.EXPECT().Run(mock.Anything, mock.Anything).Return(outCh).Run(func(ctx context.Context, in <-chan any) {
		stage1Ran = true
	})

	p := pipeline.NewPipeline()
	p.AddStage(pipeline.TypedStage(stage1))
	out := p.Run(ctx)
	for range out {
		// Pipeline execution happens as we consume the output
	}

	assert.True(t, stage1Ran, "expected the stage to be run, but it was not")
}

func TestAddStagesAndExecution(t *testing.T) {
	ctx := context.Background()
	stage1 := stages.NewMockStage[any, int](t)
	stage2 := stages.NewMockStage[int, string](t)
	var stage1Ran, stage2Ran bool
	var stage2Received int

	out1 := make(chan int, 1)
	out2 := make(chan string)
	close(out2)

	stage1.EXPECT().Run(mock.Anything, mock.Anything).Return(out1).Run(func(ctx context.Context, in <-chan any) {
		stage1Ran = true
		out1 <- 42
		close(out1)
	})

	stage2.EXPECT().Run(mock.Anything, mock.Anything).Return(out2).Run(func(ctx context.Context, in <-chan int) {
		for val := range in {
			stage2Ran = true
			stage2Received = val
		}
	})

	p := pipeline.NewPipeline()
	p.AddStage(pipeline.TypedStage(stage1))
	p.AddStage(pipeline.TypedStage(stage2))
	out := p.Run(ctx)
	for range out {
		// Pipeline execution happens as we consume the output
	}

	assert.True(t, stage1Ran, "expected stage1 to be run, but it was not")
	assert.True(t, stage2Ran, "expected stage2 to be run, but it was not")
	assert.Equal(t, 42, stage2Received, "expected stage2 to receive 42")
}

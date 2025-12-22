package pipeline_test

import (
	"context"
	"sort"
	"testing"
	"time"

	"broadcom.com/vertex-ingestor/internal/pipeline"
	"broadcom.com/vertex-ingestor/internal/stages"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestMergeFunction tests the standalone Merge function
func TestMergeFunction(t *testing.T) {
	ctx := context.Background()

	// Create two channels with test data
	ch1 := make(chan int, 3)
	ch2 := make(chan int, 3)

	ch1 <- 1
	ch1 <- 2
	ch1 <- 3
	close(ch1)

	ch2 <- 4
	ch2 <- 5
	ch2 <- 6
	close(ch2)

	// Merge the channels
	merged := pipeline.Merge(ctx, ch1, ch2)

	// Collect results
	var results []int
	for v := range merged {
		results = append(results, v)
	}

	// Sort to make test deterministic (merge order is not guaranteed)
	sort.Ints(results)

	expected := []int{1, 2, 3, 4, 5, 6}
	assert.Equal(t, expected, results, "merged channel should contain all items from both channels")
}

// TestMergeWithContext tests that Merge respects context cancellation
func TestMergeWithContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// Create two channels with test data
	ch1 := make(chan int)
	ch2 := make(chan int)

	go func() {
		for i := 0; i < 100; i++ {
			ch1 <- i
			time.Sleep(10 * time.Millisecond)
		}
		close(ch1)
	}()

	go func() {
		for i := 100; i < 200; i++ {
			ch2 <- i
			time.Sleep(10 * time.Millisecond)
		}
		close(ch2)
	}()

	// Merge the channels
	merged := pipeline.Merge(ctx, ch1, ch2)

	// Cancel context after a short time
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	// Collect results
	var results []int
	for v := range merged {
		results = append(results, v)
	}

	// Should have received some but not all items
	assert.Greater(t, len(results), 0, "should have received some items")
	assert.Less(t, len(results), 200, "should not have received all items due to cancellation")
}

// TestMergePipelines tests merging two complete pipelines
func TestMergePipelines(t *testing.T) {
	ctx := context.Background()

	// Create first pipeline that emits numbers 1-3
	pipeline1 := pipeline.NewPipeline()
	stage1 := stages.NewMockStage[any, int](t)
	out1 := make(chan int, 3)
	out1 <- 1
	out1 <- 2
	out1 <- 3
	close(out1)
	stage1.EXPECT().Run(ctx, mock.Anything).Return(out1)
	pipeline1.AddStage(pipeline.TypedStage(stage1))

	// Create second pipeline that emits numbers 4-6
	pipeline2 := pipeline.NewPipeline()
	stage2 := stages.NewMockStage[any, int](t)
	out2 := make(chan int, 3)
	out2 <- 4
	out2 <- 5
	out2 <- 6
	close(out2)
	stage2.EXPECT().Run(ctx, mock.Anything).Return(out2)
	pipeline2.AddStage(pipeline.TypedStage(stage2))

	// Create a main pipeline with the merge stage
	mainPipeline := pipeline.NewPipeline()
	mainPipeline.AddStage(pipeline.MergePipelines(ctx, pipeline1, pipeline2))

	// Run and collect results
	out := mainPipeline.Run(ctx)
	var results []int
	for v := range out {
		results = append(results, v.(int))
	}

	// Sort to make test deterministic
	sort.Ints(results)

	expected := []int{1, 2, 3, 4, 5, 6}
	assert.Equal(t, expected, results, "merged pipeline should contain all items from both pipelines")
}

// TestMergeStage tests the MergeStage directly
func TestMergeStage(t *testing.T) {
	ctx := context.Background()

	// Create input channel that will contain channels to merge
	input := make(chan any, 2)

	ch1 := make(chan any, 2)
	ch1 <- "hello"
	ch1 <- "world"
	close(ch1)

	ch2 := make(chan any, 2)
	ch2 <- "foo"
	ch2 <- "bar"
	close(ch2)

	input <- (<-chan any)(ch1)
	input <- (<-chan any)(ch2)
	close(input)

	// Create and run merge stage
	mergeStage := pipeline.NewMergeStage()
	output := mergeStage.Run(ctx, input)

	// Collect results
	var results []string
	for v := range output {
		results = append(results, v.(string))
	}

	// Sort to make test deterministic
	sort.Strings(results)

	expected := []string{"bar", "foo", "hello", "world"}
	assert.Equal(t, expected, results, "merge stage should contain all items from all input channels")
}

// TestMergeEmptyChannels tests merging empty channels
func TestMergeEmptyChannels(t *testing.T) {
	ctx := context.Background()

	ch1 := make(chan int)
	ch2 := make(chan int)
	close(ch1)
	close(ch2)

	merged := pipeline.Merge(ctx, ch1, ch2)

	var results []int
	for v := range merged {
		results = append(results, v)
	}

	assert.Empty(t, results, "merging empty channels should produce empty output")
}

// TestMergeOneEmptyChannel tests merging where one channel is empty
func TestMergeOneEmptyChannel(t *testing.T) {
	ctx := context.Background()

	ch1 := make(chan int, 3)
	ch1 <- 1
	ch1 <- 2
	ch1 <- 3
	close(ch1)

	ch2 := make(chan int)
	close(ch2)

	merged := pipeline.Merge(ctx, ch1, ch2)

	var results []int
	for v := range merged {
		results = append(results, v)
	}

	sort.Ints(results)
	expected := []int{1, 2, 3}
	assert.Equal(t, expected, results, "should receive all items from non-empty channel")
}

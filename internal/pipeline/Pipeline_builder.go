package pipeline

import (
	"fmt"

	"github.com/niedch/vectree/internal/conf"
	"github.com/niedch/vectree/internal/stages"
)

type PipelineBuilder struct {
	cfg *conf.Config
}

func NewPipelineBuilder(cfg *conf.Config) *PipelineBuilder {
	return &PipelineBuilder{cfg: cfg}
}

func (b *PipelineBuilder) BuildAll() ([]*Pipeline, error) {
	pipelines := make([]*Pipeline, 0, len(b.cfg.Sources))

	for name := range b.cfg.Sources {
		p, err := b.BuildForSource(name)
		if err != nil {
			return nil, fmt.Errorf("source %q: %w", name, err)
		}

		pipelines = append(pipelines, p)
	}
	return pipelines, nil
}

func (b *PipelineBuilder) BuildForSource(name string) (*Pipeline, error) {
	src, ok := b.cfg.Sources[name]
	if !ok {
		return nil, fmt.Errorf("source %q not found", name)
	}

	switch src.Type {
	case conf.HTTP_SOURCE_TYPE:
		return BuildHttpPipeline(src.AsHttp(), b.cfg.Pipeline), nil
	case conf.MARKDOWN_SOURCE_TYPE:
		return BuildMarkdownPipeline(src.AsMarkdown()), nil
	default:
		return nil, fmt.Errorf("unknown source type %q for source %q", src.Type, name)
	}
}

func BuildHttpPipeline(cfg *conf.HttpSourceConfig, pipelineCfg conf.Pipeline) *Pipeline {
	p := NewPipeline()
	p.AddStage(TypedStage(stages.NewDocTocLoader(cfg.Name, cfg.URL)))
	p.AddStage(TypedStage(stages.NewContentLoader(pipelineCfg.EmbedderWorkers)))
	return p
}

func BuildMarkdownPipeline(cfg *conf.MarkdownSourceConfig) *Pipeline {
	p := NewPipeline()
	p.AddStage(TypedStage(stages.NewDirLoader(cfg.Location)))
	p.AddStage(TypedStage(stages.NewFileLoader()))
	return p
}

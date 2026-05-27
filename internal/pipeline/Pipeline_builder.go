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

func (b *PipelineBuilder) BuildChunking() stages.Stage[string, stages.Section] {
	var splitter stages.Stage[string, stages.Section]
	switch b.cfg.Chunking.Strategy {
	case conf.HEADER_STRATEGY:
		splitter = stages.NewHeaderSplitter()
	case conf.LINE_STRATEGY:
		splitter = stages.NewLineSplitter()
	default:
		splitter = stages.NewMdAstSplitter()
	}
	return splitter
}

func (b *PipelineBuilder) BuildSources() ([]*Pipeline, error) {
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
	case conf.GITHUB_SOURCE_TYPE:
		return BuildGithubPipeline(src.AsGithub()), nil
	default:
		return nil, fmt.Errorf("unknown source type %q for source %q", src.Type, name)
	}
}

func BuildHttpPipeline(cfg *conf.HttpSourceConfig, pipelineCfg conf.Pipeline) *Pipeline {
	p := NewPipeline()
	p.AddStage(TypedStage(stages.NewWebCrawler(cfg.Name, cfg.URL, cfg.MaxDepth, cfg.Selector, pipelineCfg.CrawlerWorkers)))
	return p
}

func BuildMarkdownPipeline(cfg *conf.MarkdownSourceConfig) *Pipeline {
	p := NewPipeline()
	p.AddStage(TypedStage(stages.NewDirLoader(cfg.Location)))
	p.AddStage(TypedStage(stages.NewFileLoader(cfg.Name)))
	return p
}

func BuildGithubPipeline(cfg *conf.GithubSourceConfig) *Pipeline {
	p := NewPipeline()
	p.AddStage(TypedStage(stages.NewGitHubCloner(cfg.Name, cfg.Repo, cfg.Branch, cfg.Token, cfg.Subdir)))
	p.AddStage(TypedStage(stages.NewDirLoader("")))
	p.AddStage(TypedStage(stages.NewFileLoader(cfg.Name)))
	return p
}

package conf

const (
	HTTP_SOURCE_TYPE     SourceType = "http"
	MARKDOWN_SOURCE_TYPE SourceType = "markdown"
)

type SourceType string

type SourceBase struct {
	Name string
	Type SourceType
}

type Source struct {
	Name     string     `koanf:"-"`
	Type     SourceType `koanf:"type"`
	URL      string     `koanf:"url"`
	Location string     `koanf:"location"`
}

func (s Source) AsHttp() *HttpSourceConfig {
	if s.Type != HTTP_SOURCE_TYPE {
		return nil
	}
	return &HttpSourceConfig{
		SourceBase: SourceBase{Name: s.Name, Type: s.Type},
		URL:        s.URL,
	}
}

func (s Source) AsMarkdown() *MarkdownSourceConfig {
	if s.Type != MARKDOWN_SOURCE_TYPE {
		return nil
	}
	return &MarkdownSourceConfig{
		SourceBase: SourceBase{Name: s.Name, Type: s.Type},
		Location:   s.Location,
	}
}

type HttpSourceConfig struct {
	SourceBase
	URL string
}

type MarkdownSourceConfig struct {
	SourceBase
	Location string
}

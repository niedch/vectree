package conf

const (
	HTTP_SOURCE_TYPE     SourceType = "http"
	MARKDOWN_SOURCE_TYPE SourceType = "markdown"
	GITHUB_SOURCE_TYPE   SourceType = "github"
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
	MaxDepth int        `koanf:"max_depth"`
	Selector string     `koanf:"selector"`
	Repo     string     `koanf:"repo"`
	Branch   string     `koanf:"branch"`
	Token    string     `koanf:"token"`
	Subdir   string     `koanf:"subdir"`
}

type HttpSourceConfig struct {
	SourceBase
	URL      string `validate:"required,url"`
	MaxDepth int
	Selector string
}

func (s Source) AsHttp() *HttpSourceConfig {
	if s.Type != HTTP_SOURCE_TYPE {
		return nil
	}
	return &HttpSourceConfig{
		SourceBase: SourceBase{Name: s.Name, Type: s.Type},
		URL:        s.URL,
		MaxDepth:   s.MaxDepth,
		Selector:   s.Selector,
	}
}

type MarkdownSourceConfig struct {
	SourceBase
	Location string `validate:"required,dirpath"`
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

type GithubSourceConfig struct {
	SourceBase
	Repo   string `validate:"required,url"`
	Branch string
	Token  string
	Subdir string
}

func (s Source) AsGithub() *GithubSourceConfig {
	if s.Type != GITHUB_SOURCE_TYPE {
		return nil
	}
	return &GithubSourceConfig{
		SourceBase: SourceBase{Name: s.Name, Type: s.Type},
		Repo:       s.Repo,
		Branch:     s.Branch,
		Token:      s.Token,
		Subdir:     s.Subdir,
	}
}

package prompt

type Prompt struct {
	Name        string
	Description string
	Arguments   []Argument
	Source      string
}

type Argument struct {
	Name        string
	Required    bool
	Description string
}

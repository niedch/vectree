package mdast

type TokenType string

const (
	TokenHeading TokenType = "Heading"
	TokenText    TokenType = "Text"
	TokenBlank   TokenType = "Blank"
)

type Token struct {
	Type  TokenType
	Value string
	Level int // for headings
}

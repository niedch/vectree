package mdast

import "strings"

func Lex(input string) []Token {
	lines := strings.Split(input, "\n")
	tokens := make([]Token, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimRight(line, " ")

		if line == "" {
			tokens = append(tokens, Token{Type: TokenBlank})
			continue
		}

		if strings.HasPrefix(line, "#") {
			level := 0
			for level < len(line) && line[level] == '#' {
				level++
			}

			if level <= 6 && len(line) > level && line[level] == ' ' {
				tokens = append(tokens, Token{
					Type:  TokenHeading,
					Value: strings.TrimSpace(line[level:]),
					Level: level,
				})
				continue
			}
		}

		tokens = append(tokens, Token{
			Type:  TokenText,
			Value: line,
		})
	}

	return tokens
}

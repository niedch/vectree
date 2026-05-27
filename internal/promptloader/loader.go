package promptloader

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/dotprompt/go/dotprompt"
)

type LoadedPrompt struct {
	Name        string
	Description string
	Arguments   []PromptArgument
	Source      string
	FilePath    string
}

type PromptArgument struct {
	Name        string
	Required    bool
	Description string
}

func LoadDir(dirPath string) ([]LoadedPrompt, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("reading prompts directory %q: %w", dirPath, err)
	}

	var prompts []LoadedPrompt
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := filepath.Ext(entry.Name())
		if ext != ".prompt" && ext != ".dotprompt" {
			continue
		}
		fullPath := filepath.Join(dirPath, entry.Name())
		p, err := loadFile(fullPath)
		if err != nil {
			log.Printf("warning: skipping prompt file %q: %v", fullPath, err)
			continue
		}
		prompts = append(prompts, p)
	}

	return prompts, nil
}

func loadFile(path string) (LoadedPrompt, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return LoadedPrompt{}, err
	}
	source := string(data)

	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	var desc string
	var args []PromptArgument

	parsed, err := dotprompt.ParseDocument(source)
	if err == nil {
		if parsed.Name != "" {
			name = parsed.Name
		}
		desc = parsed.Description
		args = parseSchema(parsed.Input.Schema)
	}

	return LoadedPrompt{
		Name:        name,
		Description: desc,
		Arguments:   args,
		Source:      source,
		FilePath:    path,
	}, nil
}

func parseSchema(schema dotprompt.Schema) []PromptArgument {
	schemaMap, ok := schema.(map[string]any)
	if !ok {
		return nil
	}
	var args []PromptArgument
	for rawKey, val := range schemaMap {
		required := true
		name := rawKey
		if strings.HasSuffix(rawKey, "?") {
			required = false
			name = strings.TrimSuffix(rawKey, "?")
		}
		desc := fmt.Sprintf("%v", val)
		args = append(args, PromptArgument{
			Name:        name,
			Required:    required,
			Description: desc,
		})
	}
	return args
}

var sourceHasFrontmatter = func(source string) bool {
	s := strings.TrimSpace(source)
	return strings.HasPrefix(s, "---")
}

func RenderPrompt(source string, args map[string]string) (string, error) {
	dp := dotprompt.NewDotprompt(nil)

	input := make(map[string]any, len(args))
	for k, v := range args {
		input[k] = v
	}

	renderSource := source
	if !sourceHasFrontmatter(renderSource) {
		renderSource = "---\n---\n" + renderSource
	}

	rendered, err := dp.Render(renderSource, &dotprompt.DataArgument{
		Input: input,
	}, nil)
	if err != nil {
		return "", fmt.Errorf("rendering prompt: %w", err)
	}

	var parts []string
	for _, msg := range rendered.Messages {
		for _, part := range msg.Content {
			if text, ok := part.(*dotprompt.TextPart); ok {
				parts = append(parts, text.Text)
			}
		}
	}

	return strings.TrimSpace(strings.Join(parts, "\n")), nil
}

package prompt

import (
	"fmt"
	"log"

	"github.com/google/dotprompt/go/dotprompt"
	"github.com/invopop/jsonschema"
)

func LoadDir(dirPath string) ([]Prompt, error) {
	store, err := dotprompt.NewDirStore(dirPath)
	if err != nil {
		return nil, fmt.Errorf("creating prompt store for %q: %w", dirPath, err)
	}

	refs, err := store.List(dotprompt.ListPromptsOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing prompts in %q: %w", dirPath, err)
	}

	dp := dotprompt.NewDotprompt(nil)

	var prompts []Prompt
	for _, ref := range refs.Items {
		data, err := store.Load(ref.Name, dotprompt.LoadPromptOptions{Variant: ref.Variant})
		if err != nil {
			log.Printf("warning: skipping prompt %q: %v", ref.Name, err)
			continue
		}

		p, err := LoadFromSource(dp, data.Source, data.Name)
		if err != nil {
			log.Printf("warning: skipping prompt %q: %v", ref.Name, err)
			continue
		}
		prompts = append(prompts, p)
	}

	return prompts, nil
}

func LoadFromSource(dp *dotprompt.Dotprompt, source string, name string) (Prompt, error) {
	parsed, err := dp.Parse(source)
	if err != nil {
		return Prompt{}, fmt.Errorf("parsing prompt %q: %w", name, err)
	}

	if parsed.Name != "" {
		name = parsed.Name
	}

	resolved, err := dp.RenderPicoschema(parsed.PromptMetadata)
	if err != nil {
		return Prompt{}, fmt.Errorf("resolving schema for prompt %q: %w", name, err)
	}

	return Prompt{
		Name:        name,
		Description: resolved.Description,
		Arguments:   extractArguments(resolved.Input.Schema),
		Source:      parsed.Template,
	}, nil
}

func extractArguments(schema dotprompt.Schema) []Argument {
	schemaObj, ok := schema.(*jsonschema.Schema)
	if !ok || schemaObj == nil {
		return nil
	}

	if schemaObj.Properties == nil {
		return nil
	}

	required := make(map[string]bool, len(schemaObj.Required))
	for _, name := range schemaObj.Required {
		required[name] = true
	}

	var args []Argument
	for pair := schemaObj.Properties.Oldest(); pair != nil; pair = pair.Next() {
		desc := pair.Value.Description
		if desc == "" {
			desc = pair.Value.Type
		}
		if desc == "" {
			desc = "any"
		}
		args = append(args, Argument{
			Name:        pair.Key,
			Required:    required[pair.Key],
			Description: desc,
		})
	}
	return args
}
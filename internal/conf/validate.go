package conf

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	validate *validator.Validate = validator.New(validator.WithRequiredStructEnabled())
)

func validateConfig(cfg *Config) error {
	if err := validateSources(cfg.Sources); err != nil {
		return err
	}

	if err := validateChunking(cfg.Chunking); err != nil {
		return err
	}

	if err := validateAI(cfg.AI); err != nil {
		return err
	}

	err := validate.Struct(cfg)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	return nil
}

func validateChunking(c Chunking) error {
	switch c.Strategy {
	case MDAST_STRATEGY, HEADER_STRATEGY, LINE_STRATEGY:
		return nil
	default:
		return fmt.Errorf("chunking: strategy %q is invalid (must be one of: mdast, header, line)", c.Strategy)
	}
}

func validateAI(ai AI) error {
	var typed any
	switch ai.Provider {
	case GEMINI_PROVIDER:
		typed = ai.AsGeminiProviderConfig()
	case OPENAI_PROVIDER:
		typed = ai.AsOpenAIProviderConfig()
	case OLLAMA_PROVIDER:
		typed = ai.AsOllamaProviderConfig()
	default:
		return fmt.Errorf("ai: provider %q is invalid", ai.Provider)
	}

	if errs := validateTyped("ai", typed); len(errs) > 0 {
		return fmt.Errorf("\n\t%s", strings.Join(errs, "\n\t"))
	}
	return nil
}

func validateSources(sources map[string]Source) error {
	if len(sources) < 1 {
		return fmt.Errorf("no sources defined")
	}

	var errs []string
	for name, src := range sources {
		switch src.Type {
		case HTTP_SOURCE_TYPE:
			errs = append(errs, validateTyped(fmt.Sprintf("source %q", name), src.AsHttp())...)
		case MARKDOWN_SOURCE_TYPE:
			errs = append(errs, validateTyped(fmt.Sprintf("source %q", name), src.AsMarkdown())...)
		case GITHUB_SOURCE_TYPE:
			errs = append(errs, validateTyped(fmt.Sprintf("source %q", name), src.AsGithub())...)
		default:
			errs = append(errs, fmt.Sprintf("source %q: type %q is invalid", name, src.Type))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("\n\t%s", strings.Join(errs, "\n\t"))
	}
	return nil
}

func validateTyped(prefix string, v any) []string {
	err := validate.Struct(v)
	if err == nil {
		return nil
	}
	var errs []string
	for _, fe := range err.(validator.ValidationErrors) {
		errs = append(errs, fmt.Sprintf("%s: %s", prefix, fmtErr(fe)))
	}
	return errs
}

func fmtErr(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", err.Field())
	case "url":
		return fmt.Sprintf("%s must be a valid URL", err.Field())
	case "dirpath":
		return fmt.Sprintf("%s must be a valid Directory", err.Field())
	default:
		return fmt.Sprintf("%s failed validation: %s", err.Field(), err.Tag())
	}
}

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

	err := validate.Struct(cfg)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	return nil
}

func validateSources(sources map[string]Source) error {
	if len(sources) < 1 {
		return fmt.Errorf("no sources defined")
	}

	var errs []string
	for name, src := range sources {
		var err error
		switch src.Type {
		case HTTP_SOURCE_TYPE:
			err = validate.Struct(src.AsHttp())
		case MARKDOWN_SOURCE_TYPE:
			err = validate.Struct(src.AsMarkdown())
		default:
			errs = append(errs, fmt.Sprintf("source %q: type %q is invalid", name, src.Type))
			continue
		}

		if err != nil {
			for _, fe := range err.(validator.ValidationErrors) {
				errs = append(errs, fmt.Sprintf("source %q: %s", name, fmtErr(fe)))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("\n\t%s", strings.Join(errs, "\n\t"))
	}
	return nil
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

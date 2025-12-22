package mcptemplate

import (
	"bytes"
	_ "embed"
	"strings"
	"text/template"

	"broadcom.com/vertex-ingestor/internal/datastore"
)

//go:embed template/mcp_response.tmpl
var mcpResponseTemplate string

// BuildResponseString formats a slice of documents with embeddings into a markdown string
// using the embedded template. Each document result includes metadata (ID and Level)
// and the document content.
func BuildResponseString(docs []datastore.DocumentWithEmbedding) string {
	// Create template with custom functions
	tmpl, err := template.New("mcp_response").Funcs(template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"deref": func(ptr *int) int {
			if ptr == nil {
				return 0
			}
			return *ptr
		},
	}).Parse(mcpResponseTemplate)

	if err != nil {
		// Fallback to simple concatenation if template parsing fails
		return fallbackResponseString(docs)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, docs)
	if err != nil {
		// Fallback to simple concatenation if template execution fails
		return fallbackResponseString(docs)
	}

	return buf.String()
}

// fallbackResponseString provides a simple concatenation fallback
// in case template parsing or execution fails
func fallbackResponseString(docs []datastore.DocumentWithEmbedding) string {
	sb := strings.Builder{}
	for _, doc := range docs {
		sb.WriteString(doc.Document.Document)
		sb.WriteString("\n")
	}
	return sb.String()
}

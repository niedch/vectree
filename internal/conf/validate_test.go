package conf

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateSuccess(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "*.toml")
	require.NoError(t, err)
	defer f.Close()

	config_str := `
[sources.tech-docs]
type = "http"
url = "https://techdocs.broadcom.com/us/en/ca-enterprise-software/valueops/connectall/4-0/jcr:content.toc.html"
`
	_, err = f.WriteString(config_str)
	require.NoError(t, err)

	_, err = loadCustomFile(f.Name())
	require.NoError(t, err)
}

func TestInvalidSourceType(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "*.toml")
	require.NoError(t, err)
	defer f.Close()

	config_str := `
[sources.tech-docs]
type = "invalidtype"
`
	_, err = f.WriteString(config_str)
	require.NoError(t, err)

	_, err = loadCustomFile(f.Name())
	require.Error(t, err)
}

func TestHttpNeedToHaveValidUrl(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "*.toml")
	require.NoError(t, err)
	defer f.Close()

	config_str := `
[sources.tech-docs]
type = "http"
url = "asdfasdf"
	`
	_, err = f.WriteString(config_str)
	require.NoError(t, err)

	_, err = loadCustomFile(f.Name())
	require.Error(t, err)
}

func TestMarkdownNeedsToPointToAFolder(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "*.toml")
	require.NoError(t, err)
	defer f.Close()

	config_str := `
[sources.tech-docs]
type = "markdown"
location = "internal/conf/"
	`
	_, err = f.WriteString(config_str)
	require.NoError(t, err)

	_, err = loadCustomFile(f.Name())
	require.NoError(t, err)
}

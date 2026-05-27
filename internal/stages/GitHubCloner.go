package stages

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

type GitHubCloner struct {
	sourceName string
	repo       string
	branch     string
	token      string
	subdir     string
}

func NewGitHubCloner(sourceName, repo, branch, token, subdir string) *GitHubCloner {
	return &GitHubCloner{
		sourceName: sourceName,
		repo:       repo,
		branch:     branch,
		token:      token,
		subdir:     subdir,
	}
}

func (g *GitHubCloner) Run(ctx context.Context, _ <-chan any) <-chan string {
	out := make(chan string)

	go func() {
		defer close(out)

		tmpDir, err := os.MkdirTemp("", "vectree-github-*")
		if err != nil {
			log.Printf("[%s] Failed to create temp dir: %v", g.sourceName, err)
			return
		}
		defer func() {
			<-ctx.Done()
			os.RemoveAll(tmpDir)
		}()

		token := g.token
		if token == "" {
			token = os.Getenv("GITHUB_TOKEN")
		}

		cloneOpts := &gogit.CloneOptions{
			URL:   g.repo,
			Depth: 1,
		}

		if token != "" {
			if !strings.HasPrefix(g.repo, "https://") {
				log.Printf("[%s] Token auth only supported for HTTPS repos", g.sourceName)
				return
			}
			cloneOpts.Auth = &http.BasicAuth{
				Username: "token",
				Password: token,
			}
		}

		if g.branch != "" {
			cloneOpts.ReferenceName = plumbing.ReferenceName("refs/heads/" + g.branch)
		}

		_, err = gogit.PlainClone(tmpDir, false, cloneOpts)
		if err != nil {
			log.Printf("[%s] Failed to clone repo %s: %v", g.sourceName, g.repo, err)
			return
		}

		log.Printf("[%s] Cloned repo %s", g.sourceName, g.repo)

		root := tmpDir
		if g.subdir != "" {
			root = filepath.Join(tmpDir, g.subdir)
		}

		select {
		case out <- root:
		case <-ctx.Done():
		}
	}()

	return out
}

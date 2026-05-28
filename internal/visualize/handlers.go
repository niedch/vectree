package visualize

import (
	"net/http"
	"strings"
)

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" && r.URL.Path != "/index.html" {
		if r.URL.Path != "/favicon.ico" {
			http.NotFound(w, r)
		}
		return
	}

	content, err := templatesFS.ReadFile("templates/index.html")
	if err != nil {
		http.Error(w, "Failed to load index.html", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(content)
}

func (s *Server) handleComponent(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/components/")
	if name == "" || strings.Contains(name, "/") {
		http.NotFound(w, r)
		return
	}

	content, err := templatesFS.ReadFile("templates/components/" + name)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/javascript")
	w.Write(content)
}

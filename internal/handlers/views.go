package handlers

import (
	"embed"
	"html/template"
	"log"
	"net/http"

	"github.com/stpotter16/coin/internal/handlers/middleware"
)

type viewProps struct {
	CsrfToken string
	CspNonce  string
}

//go:embed templates
var templateFS embed.FS

func loginGet() http.HandlerFunc {
	t := template.Must(
		template.New("base.html").
			ParseFS(
				templateFS,
				"templates/layouts/base.html", "templates/pages/login.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		nonce, err := extractCspNonceOnly(r)
		if err != nil {
			log.Printf("Could not extract csp nonce from ctx: %v", err)
			http.Error(w, "Could not construct session nonce", http.StatusInternalServerError)
			return
		}
		props := viewProps{CsrfToken: "", CspNonce: nonce}
		if err := t.Execute(w, struct {
			viewProps
		}{
			viewProps: props,
		}); err != nil {
			log.Printf("Could not create login page: %v", err)
			http.Error(w, "Server issue - try again later", http.StatusInternalServerError)
		}
	}
}

func indexGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {}
}

func extractCspNonceOnly(r *http.Request) (string, error) {
	cspNonce, err := middleware.NonceFromContext(r.Context())
	if err != nil {
		return "", err
	}
	return cspNonce, nil
}

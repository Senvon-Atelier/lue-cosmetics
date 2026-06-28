package email

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	texttemplate "text/template"
)

//go:embed templates
var templatesFS embed.FS

type Renderer struct {
	html *template.Template
	text *texttemplate.Template
}

func NewRenderer() (*Renderer, error) {
	html, err := template.ParseFS(fs.FS(templatesFS), "templates/*.html.tmpl")
	if err != nil {
		return nil, fmt.Errorf("email: parse html templates: %w", err)
	}
	text, err := texttemplate.ParseFS(fs.FS(templatesFS), "templates/*.txt.tmpl")
	if err != nil {
		return nil, fmt.Errorf("email: parse text templates: %w", err)
	}
	return &Renderer{html: html, text: text}, nil
}

// Render returns (htmlBody, textBody, error). The template name is the
// logical name (e.g., "order_confirmation"); the renderer looks up
// {name}.html.tmpl and {name}.txt.tmpl.
func (r *Renderer) Render(name string, data map[string]any) (string, string, error) {
	var hb bytes.Buffer
	if err := r.html.ExecuteTemplate(&hb, name+".html.tmpl", data); err != nil {
		return "", "", fmt.Errorf("email: html render %s: %w", name, err)
	}
	var tb bytes.Buffer
	if err := r.text.ExecuteTemplate(&tb, name+".txt.tmpl", data); err != nil {
		return "", "", fmt.Errorf("email: text render %s: %w", name, err)
	}
	return hb.String(), tb.String(), nil
}

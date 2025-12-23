package email

import (
	"bytes"
	"html/template"

	"go-reasonable-api/emails/templates"

	"github.com/rotisserie/eris"
)

// Templates holds the parsed email templates
type Templates struct {
	tmpl *template.Template
}

// NewTemplates parses and returns email templates
func NewTemplates() (*Templates, error) {
	tmpl, err := template.ParseFS(templates.EmailsFS, "*.html")
	if err != nil {
		return nil, eris.Wrap(err, "failed to parse email templates")
	}

	return &Templates{tmpl: tmpl}, nil
}

// Render renders an email template with the given data
func (t *Templates) Render(name string, data any) (string, error) {
	var buf bytes.Buffer
	if err := t.tmpl.ExecuteTemplate(&buf, name+".html", data); err != nil {
		return "", eris.Wrapf(err, "failed to render email template %s", name)
	}
	return buf.String(), nil
}

package server

import (
	"bytes"
	"html/template"
	"testing"

	"bros_kiosk/internal/config"
)

func TestTemplateExecution(t *testing.T) {
	tmplStr := `{{range .Sections}}<section id="{{.ID}}"></section>{{end}}`
	tmpl, err := template.New("test").Parse(tmplStr)
	if err != nil {
		t.Fatal(err)
	}

	data := config.Config{
		Sections: []config.Section{
			{ID: "test1", Type: "widget", Style: "default"},
		},
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatal(err)
	}

	expected := `<section id="test1"></section>`
	if buf.String() != expected {
		t.Errorf("expected %q, got %q", expected, buf.String())
	}
}

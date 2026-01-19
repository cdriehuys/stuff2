package templating

import (
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"path/filepath"
	texttemplate "text/template"
)

type LiveLoader struct {
	Logger  *slog.Logger
	BaseDir string
}

func (l *LiveLoader) Render(w io.Writer, page string, data any) error {
	basePath := filepath.Join(l.BaseDir, "base.html")
	partialsPattern := filepath.Join(l.BaseDir, "partials", "*.html")
	pagePath := filepath.Join(l.BaseDir, "pages", page)

	files := []string{basePath}

	partials, err := filepath.Glob(partialsPattern)
	if err != nil {
		return fmt.Errorf("looking for partials: %v", err)
	}

	if partials != nil {
		files = append(files, partials...)
	}

	files = append(files, pagePath)

	t, err := template.ParseFiles(files...)
	if err != nil {
		return fmt.Errorf("parsing files for %q: %v", page, err)
	}

	return t.ExecuteTemplate(w, "main", data)
}

type LiveEmailLoader struct {
	Logger  *slog.Logger
	BaseDir string
}

func (l *LiveEmailLoader) Render(w io.Writer, subject string, data any) error {
	basePath := filepath.Join(l.BaseDir, "base.txt")
	pagePath := filepath.Join(l.BaseDir, "subjects", subject)

	t, err := texttemplate.ParseFiles(basePath, pagePath)
	if err != nil {
		return err
	}

	return t.ExecuteTemplate(w, "main", data)
}

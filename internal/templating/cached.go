package templating

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log/slog"
	"path/filepath"
	texttemplate "text/template"
)

type TemplateCache struct {
	logger *slog.Logger

	cache map[string]*template.Template
}

func NewTemplateCache(logger *slog.Logger, files fs.FS) (*TemplateCache, error) {
	basePath := "base.html"
	pagesPath := "pages"

	var pages []string
	visit := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if filepath.Ext(path) == ".html" {
			pages = append(pages, path)
		}

		return nil
	}

	if err := fs.WalkDir(files, pagesPath, visit); err != nil {
		return nil, fmt.Errorf("failed to collect pages: %v", err)
	}

	cache := make(map[string]*template.Template, len(pages))
	for _, page := range pages {
		name, err := filepath.Rel(pagesPath, page)
		if err != nil {
			return nil, fmt.Errorf("failed to determine relative path for page %q: %v", page, err)
		}

		t, err := template.ParseFS(files, basePath, page)
		if err != nil {
			return nil, fmt.Errorf("failed to construct template for page %q: %v", page, err)
		}

		cache[name] = t
	}

	return &TemplateCache{logger, cache}, nil
}

func (c *TemplateCache) Render(w io.Writer, page string, data any) error {
	t, exists := c.cache[page]
	if !exists {
		return fmt.Errorf("template not found: %v", page)
	}

	return t.ExecuteTemplate(w, "main", data)
}

type EmailTemplateCache struct {
	logger *slog.Logger

	cache map[string]*texttemplate.Template
}

func NewEmailTemplateCache(logger *slog.Logger, files fs.FS) (*EmailTemplateCache, error) {
	basePath := "base.txt"
	subjectsPath := "subjects"

	var subjects []string
	visit := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if filepath.Ext(path) == ".txt" {
			subjects = append(subjects, path)
		}

		return nil
	}

	if err := fs.WalkDir(files, subjectsPath, visit); err != nil {
		return nil, fmt.Errorf("collecting subjects: %v", err)
	}

	cache := make(map[string]*texttemplate.Template, len(subjects))
	for _, subject := range subjects {
		name, err := filepath.Rel(subjectsPath, subject)
		if err != nil {
			return nil, fmt.Errorf("determining relative path for subject %q: %v", subject, err)
		}

		t, err := texttemplate.ParseFS(files, basePath, subject)
		if err != nil {
			return nil, fmt.Errorf("constructing template for subject %q: %v", subject, err)
		}

		cache[name] = t
	}

	return &EmailTemplateCache{logger, cache}, nil
}

func (c *EmailTemplateCache) Render(w io.Writer, subject string, data any) error {
	t, exists := c.cache[subject]
	if !exists {
		return fmt.Errorf("template not found: %v", subject)
	}

	return t.ExecuteTemplate(w, "main", data)
}

package templating_test

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/cdriehuys/stuff2/internal/templating"
)

const standardBaseTemplate = `{{ define "main" }}{{ block "content" . }}{{ end }}{{ end }}`

const helloOutput = "Hello, World!"
const helloTemplate = `{{ define "content" }}Hello, World!{{ end }}`

func TestLiveLoader_Render(t *testing.T) {
	tests := []struct {
		name string

		// setup
		baseTemplate     string
		partialTemplates map[string]string
		pageTemplates    map[string]string

		// parameters
		page string
		data any

		// expectations
		want    string
		wantErr bool
	}{
		{
			name:    "base template missing",
			page:    "foo.html",
			wantErr: true,
		},
		{
			name:         "missing page",
			baseTemplate: standardBaseTemplate,

			page:    "foo.html",
			wantErr: true,
		},
		{
			name:         "single block",
			baseTemplate: standardBaseTemplate,
			pageTemplates: map[string]string{
				"hello.html": helloTemplate,
			},
			page: "hello.html",
			want: helloOutput,
		},
		{
			name:         "uses data",
			baseTemplate: standardBaseTemplate,
			pageTemplates: map[string]string{
				"data.html": `{{ define "content" }}{{ .Content }}{{ end}}`,
			},
			page: "data.html",
			data: map[string]string{"Content": "Refrigerator"},
			want: "Refrigerator",
		},
		{
			name:         "uses partial",
			baseTemplate: standardBaseTemplate,
			partialTemplates: map[string]string{
				"foo.html": `{{ define "foo" }}Hello from foo.{{ end }}`,
			},
			pageTemplates: map[string]string{
				"page.html": `{{ define "content" }}{{ template "foo" . }}{{ end }}`,
			},
			page: "page.html",
			want: "Hello from foo.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			if tt.baseTemplate != "" {
				if err := os.WriteFile(filepath.Join(dir, "base.html"), []byte(tt.baseTemplate), 0o644); err != nil {
					t.Fatalf("failed to create base.html: %v", err)
				}
			}

			if len(tt.partialTemplates) > 0 {
				if err := os.Mkdir(filepath.Join(dir, "partials"), 0o755); err != nil {
					t.Fatalf("failed to create partials dir: %v", err)
				}

				for partial, template := range tt.partialTemplates {
					if err := os.WriteFile(filepath.Join(dir, "partials", partial), []byte(template), 0o644); err != nil {
						t.Fatalf("failed to write partial template %q: %v", partial, err)
					}
				}
			}

			if len(tt.pageTemplates) > 0 {
				if err := os.Mkdir(filepath.Join(dir, "pages"), 0o755); err != nil {
					t.Fatalf("failed to create pages dir: %v", err)
				}

				for page, template := range tt.pageTemplates {
					if err := os.WriteFile(filepath.Join(dir, "pages", page), []byte(template), 0o644); err != nil {
						t.Fatalf("failed to write page %q: %v", page, err)
					}
				}
			}

			l := templating.LiveLoader{Logger: slog.New(slog.DiscardHandler), BaseDir: dir}

			var buffer bytes.Buffer

			gotErr := l.Render(&buffer, tt.page, tt.data)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Render() failed: %v", gotErr)
				}

				return
			}

			if tt.wantErr {
				t.Fatal("Render() succeeded unexpectedly")
			}

			got := buffer.String()
			if got != tt.want {
				t.Errorf("Expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestLiveEmailLoader_Render(t *testing.T) {
	tests := []struct {
		name string

		// setup
		baseTemplate     string
		subjectTemplates map[string]string

		// parameters
		subject string
		data    any

		// expectations
		want    string
		wantErr bool
	}{
		{
			name:    "base template missing",
			subject: "foo.txt",
			wantErr: true,
		},
		{
			name:         "missing page",
			baseTemplate: standardBaseTemplate,

			subject: "foo.txt",
			wantErr: true,
		},
		{
			name:         "single block",
			baseTemplate: standardBaseTemplate,
			subjectTemplates: map[string]string{
				"hello.txt": helloTemplate,
			},
			subject: "hello.txt",
			want:    helloOutput,
		},
		{
			name:         "uses data",
			baseTemplate: standardBaseTemplate,
			subjectTemplates: map[string]string{
				"data.txt": `{{ define "content" }}{{ .Content }}{{ end}}`,
			},
			subject: "data.txt",
			data:    map[string]string{"Content": "Refrigerator"},
			want:    "Refrigerator",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			if tt.baseTemplate != "" {
				if err := os.WriteFile(filepath.Join(dir, "base.txt"), []byte(tt.baseTemplate), 0o644); err != nil {
					t.Fatalf("failed to create base.txt: %v", err)
				}
			}

			if len(tt.subjectTemplates) > 0 {
				if err := os.Mkdir(filepath.Join(dir, "subjects"), 0o755); err != nil {
					t.Fatalf("failed to create subjects dir: %v", err)
				}

				for page, template := range tt.subjectTemplates {
					if err := os.WriteFile(filepath.Join(dir, "subjects", page), []byte(template), 0o644); err != nil {
						t.Fatalf("failed to write subject %q: %v", page, err)
					}
				}
			}

			l := templating.LiveEmailLoader{Logger: slog.New(slog.DiscardHandler), BaseDir: dir}

			var buffer bytes.Buffer

			gotErr := l.Render(&buffer, tt.subject, tt.data)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Render() failed: %v", gotErr)
				}

				return
			}

			if tt.wantErr {
				t.Fatal("Render() succeeded unexpectedly")
			}

			got := buffer.String()
			if got != tt.want {
				t.Errorf("Expected %q, got %q", tt.want, got)
			}
		})
	}
}

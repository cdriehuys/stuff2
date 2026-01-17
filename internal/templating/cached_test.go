package templating_test

import (
	"bytes"
	"log/slog"
	"testing"
	"testing/fstest"

	"github.com/cdriehuys/stuff2/internal/templating"
)

var testFS = fstest.MapFS{
	"base.html": &fstest.MapFile{
		Data: []byte(`{{ define "main" }}{{ block "content" . }}{{ end }}{{ end }}`),
	},
	"pages/content.html": &fstest.MapFile{
		Data: []byte(`{{ define "content" }}{{ .Content }}{{ end }}`),
	},
	"pages/hello.html": &fstest.MapFile{
		Data: []byte(`{{ define "content" }}Hello{{ end }}`),
	},
}

func TestTemplateCache_Render(t *testing.T) {
	tests := []struct {
		name string

		page string
		data any

		want    string
		wantErr bool
	}{
		{
			name: "static page",
			page: "hello.html",
			want: "Hello",
		},
		{
			name: "inject content",
			page: "content.html",
			data: map[string]string{"Content": "custom content"},
			want: "custom content",
		},
		{
			name:    "missing page",
			page:    "missing.html",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := templating.NewTemplateCache(slog.New(slog.DiscardHandler), testFS)
			if err != nil {
				t.Fatalf("could not construct template cache: %v", err)
			}

			var buffer bytes.Buffer
			gotErr := c.Render(&buffer, tt.page, tt.data)
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

var testEmailFS = fstest.MapFS{
	"base.txt": &fstest.MapFile{
		Data: []byte(`{{ define "main" }}{{ block "content" . }}{{ end }}{{ end }}`),
	},
	"subjects/content.txt": &fstest.MapFile{
		Data: []byte(`{{ define "content" }}{{ .Content }}{{ end }}`),
	},
	"subjects/hello.txt": &fstest.MapFile{
		Data: []byte(`{{ define "content" }}Hello{{ end }}`),
	},
}

func TestEmailTemplateCache_Render(t *testing.T) {
	tests := []struct {
		name string

		subject string
		data    any

		want    string
		wantErr bool
	}{
		{
			name:    "static page",
			subject: "hello.txt",
			want:    "Hello",
		},
		{
			name:    "inject content",
			subject: "content.txt",
			data:    map[string]string{"Content": "custom content"},
			want:    "custom content",
		},
		{
			name:    "missing subject",
			subject: "missing.txt",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := templating.NewEmailTemplateCache(slog.New(slog.DiscardHandler), testEmailFS)
			if err != nil {
				t.Fatalf("could not construct template cache: %v", err)
			}

			var buffer bytes.Buffer
			gotErr := c.Render(&buffer, tt.subject, tt.data)
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

package main

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"testing"
)

const minimalDocxXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
    <w:p><w:r><w:t>Участник должен предоставить лицензию</w:t></w:r></w:p>
    <w:p><w:r><w:t>Второй абзац</w:t></w:r></w:p>
  </w:body>
</w:document>`

func writeDocx(t *testing.T, path string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create docx: %v", err)
	}
	defer f.Close()
	zw := zip.NewWriter(f)
	w, err := zw.Create("word/document.xml")
	if err != nil {
		t.Fatalf("create entry: %v", err)
	}
	if _, err := w.Write([]byte(minimalDocxXML)); err != nil {
		t.Fatalf("write entry: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
}

func TestParseDocx(t *testing.T) {
	tests := []struct {
		name      string
		wantNames []string
		wantText  []string
	}{
		{
			name:      "two paragraphs",
			wantNames: []string{"paragraph_1", "paragraph_2"},
			wantText:  []string{"Участник должен предоставить лицензию", "Второй абзац"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "requirements.docx")
			writeDocx(t, path)

			pkg, err := parseDocx(context.Background(), path)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(pkg.Documents) != len(tt.wantNames) {
				t.Fatalf("got %d documents, want %d", len(pkg.Documents), len(tt.wantNames))
			}
			for i, d := range pkg.Documents {
				if d.Name != tt.wantNames[i] {
					t.Errorf("doc %d name = %q, want %q", i, d.Name, tt.wantNames[i])
				}
				if d.Content != tt.wantText[i] {
					t.Errorf("doc %d content = %q, want %q", i, d.Content, tt.wantText[i])
				}
			}
		})
	}
}

func TestParseDocxMissingDocumentXML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "empty.docx")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create docx: %v", err)
	}
	zw := zip.NewWriter(f)
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close file: %v", err)
	}

	if _, err := parseDocx(context.Background(), path); err == nil {
		t.Fatal("expected error, got nil")
	}
}

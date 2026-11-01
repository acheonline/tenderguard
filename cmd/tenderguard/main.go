package main

import (
	"archive/zip"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/acheonline/tenderguard/internal/analysis"
)

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		slog.Error("tenderguard failed", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: tenderguard <input.json|input.docx>")
	}
	var pkg analysis.Package
	switch strings.ToLower(filepath.Ext(args[0])) {
	case ".docx":
		p, err := parseDocx(ctx, args[0])
		if err != nil {
			return err
		}
		pkg = p
	case ".json":
		data, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("read input: %w", err)
		}
		if err := json.Unmarshal(data, &pkg); err != nil {
			return fmt.Errorf("parse input: %w", err)
		}
	default:
		return fmt.Errorf("unsupported input extension %q", filepath.Ext(args[0]))
	}
	report, err := analysis.Analyze(ctx, pkg)
	if err != nil {
		return fmt.Errorf("analyze: %w", err)
	}
	fmt.Print(report.String())
	return nil
}

func parseDocx(ctx context.Context, path string) (analysis.Package, error) {
	if err := ctx.Err(); err != nil {
		return analysis.Package{}, fmt.Errorf("parse docx: %w", err)
	}
	zr, err := zip.OpenReader(path)
	if err != nil {
		return analysis.Package{}, fmt.Errorf("open docx: %w", err)
	}
	defer zr.Close()

	var docXML io.ReadCloser
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			rc, err := f.Open()
			if err != nil {
				return analysis.Package{}, fmt.Errorf("open document.xml: %w", err)
			}
			docXML = rc
			break
		}
	}
	if docXML == nil {
		return analysis.Package{}, fmt.Errorf("parse docx: word/document.xml not found")
	}
	defer docXML.Close()

	paragraphs, err := extractParagraphs(docXML)
	if err != nil {
		return analysis.Package{}, fmt.Errorf("parse docx: %w", err)
	}

	docs := make([]analysis.Document, 0, len(paragraphs))
	for i, text := range paragraphs {
		docs = append(docs, analysis.Document{
			Name:    fmt.Sprintf("paragraph_%d", i+1),
			Content: text,
		})
	}
	return analysis.Package{Documents: docs}, nil
}

func extractParagraphs(r io.Reader) ([]string, error) {
	dec := xml.NewDecoder(r)
	var (
		paragraphs []string
		current    strings.Builder
		inText     bool
		inPara     bool
	)
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("decode xml: %w", err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "p":
				inPara = true
				current.Reset()
			case "t":
				inText = true
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "p":
				if inPara {
					paragraphs = append(paragraphs, current.String())
				}
				inPara = false
			case "t":
				inText = false
			}
		case xml.CharData:
			if inText {
				current.Write(t)
			}
		}
	}
	return paragraphs, nil
}

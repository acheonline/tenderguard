package analysis

import (
	"archive/zip"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// ParseDocx extracts paragraphs and tables from a .docx file into a Package.
func ParseDocx(ctx context.Context, path string) (Package, error) {
	if err := ctx.Err(); err != nil {
		return Package{}, fmt.Errorf("parse docx: %w", err)
	}
	zr, err := zip.OpenReader(path)
	if err != nil {
		return Package{}, fmt.Errorf("open docx: %w", err)
	}
	defer zr.Close()

	var docXML io.ReadCloser
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			rc, err := f.Open()
			if err != nil {
				return Package{}, fmt.Errorf("open document.xml: %w", err)
			}
			docXML = rc
			break
		}
	}
	if docXML == nil {
		return Package{}, fmt.Errorf("parse docx: word/document.xml not found")
	}
	defer docXML.Close()

	docs, err := extractBlocks(docXML)
	if err != nil {
		return Package{}, fmt.Errorf("parse docx: %w", err)
	}
	return Package{Documents: docs}, nil
}

func extractBlocks(r io.Reader) ([]Document, error) {
	dec := xml.NewDecoder(r)
	var (
		docs       []Document
		paraCount  int
		tableCount int
		current    strings.Builder
		inText     bool
		inPara     bool
		inTable    bool
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
			case "tbl":
				inTable = true
				current.Reset()
			case "p":
				inPara = true
				if !inTable {
					current.Reset()
				}
			case "t":
				inText = true
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "tbl":
				if inTable {
					tableCount++
					docs = append(docs, Document{
						Name:    fmt.Sprintf("table_%d", tableCount),
						Content: current.String(),
					})
				}
				inTable = false
				current.Reset()
			case "p":
				if inPara && !inTable {
					paraCount++
					docs = append(docs, Document{
						Name:    fmt.Sprintf("paragraph_%d", paraCount),
						Content: current.String(),
					})
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
	return docs, nil
}

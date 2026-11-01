package analysis

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

// Document is a single document in the application package.
type Document struct {
	// Name is the document identifier, e.g. "form_1" or "certificate_tax".
	Name string
	// Title is the human-readable title.
	Title string
	// Section is the section within the document, if applicable.
	Section string
	// Content is the textual content of the document.
	Content string
	// IssuedAt is the issue date for certificates; zero if not applicable.
	IssuedAt time.Time
	// Fields holds structured form field values.
	Fields map[string]string
}

// Requirement describes a single requirement from the procurement documentation.
type Requirement struct {
	// DocumentName is the required document name.
	DocumentName string
	// Title is the human-readable title.
	Title string
	// RequiredFields lists form fields that must be filled.
	RequiredFields []string
	// MaxAgeDays is the maximum allowed age of a certificate in days; 0 means no limit.
	MaxAgeDays int
	// MustContain lists substrings that must appear in the document content.
	MustContain []string
	// Law is the applicable law, e.g. "44-ФЗ" or "223-ФЗ".
	Law string
}

// Package is the input to the analyzer.
type Package struct {
	// CustomerINN is the INN of the customer.
	CustomerINN string
	// Documents is the application package.
	Documents []Document
	// Requirements is the procurement documentation requirements.
	Requirements []Requirement
	// Now is the reference time for freshness checks; zero means time.Now().
	Now time.Time
}

// SupportedCustomers lists INNs of supported customers for stage 1.
var SupportedCustomers = map[string]string{
	"7804678913": "АО «Невский экологический оператор»",
	"7841312071": "ПАО «ТГК-1»",
	"7838129301": "АО «ТЭК СПб»",
}

// Analyze checks the application package against the procurement requirements.
func Analyze(ctx context.Context, pkg Package) (Report, error) {
	if err := ctx.Err(); err != nil {
		return Report{}, fmt.Errorf("analyze: %w", err)
	}
	if pkg.CustomerINN == "" {
		return Report{}, fmt.Errorf("analyze: customer INN is empty")
	}
	if _, ok := SupportedCustomers[pkg.CustomerINN]; !ok {
		slog.WarnContext(ctx, "customer is not in stage 1 scope", "inn", pkg.CustomerINN)
	}

	now := pkg.Now
	if now.IsZero() {
		now = time.Now()
	}

	docs := make(map[string]Document, len(pkg.Documents))
	for _, d := range pkg.Documents {
		docs[d.Name] = d
	}

	var findings []Finding
	for _, req := range pkg.Requirements {
		doc, ok := docs[req.DocumentName]
		if !ok {
			findings = append(findings, Finding{
				Severity: SeverityError,
				Code:     "missing_document",
				Document: req.DocumentName,
				Section:  req.Title,
				Message:  fmt.Sprintf("required document %q is missing", req.Title),
			})
			continue
		}
		findings = append(findings, checkDocument(now, req, doc)...)
	}

	return Report{Findings: findings}, nil
}

func checkDocument(now time.Time, req Requirement, doc Document) []Finding {
	var findings []Finding

	for _, field := range req.RequiredFields {
		v, ok := doc.Fields[field]
		if !ok || strings.TrimSpace(v) == "" {
			findings = append(findings, Finding{
				Severity: SeverityError,
				Code:     "empty_field",
				Document: doc.Name,
				Section:  doc.Section,
				Message:  fmt.Sprintf("field %q is not filled", field),
			})
		}
	}

	for _, sub := range req.MustContain {
		if !strings.Contains(doc.Content, sub) {
			findings = append(findings, Finding{
				Severity: SeverityWarning,
				Code:     "missing_content",
				Document: doc.Name,
				Section:  doc.Section,
				Message:  fmt.Sprintf("content does not mention %q", sub),
			})
		}
	}

	if req.MaxAgeDays > 0 {
		if doc.IssuedAt.IsZero() {
			findings = append(findings, Finding{
				Severity: SeverityWarning,
				Code:     "missing_issue_date",
				Document: doc.Name,
				Section:  doc.Section,
				Message:  "certificate issue date is missing",
			})
		} else {
			age := now.Sub(doc.IssuedAt)
			if age > time.Duration(req.MaxAgeDays)*24*time.Hour {
				findings = append(findings, Finding{
					Severity: SeverityError,
					Code:     "stale_certificate",
					Document: doc.Name,
					Section:  doc.Section,
					Message:  fmt.Sprintf("certificate is older than %d days", req.MaxAgeDays),
				})
			}
		}
	}

	if req.Law != "" {
		findings = append(findings, Finding{
			Severity: SeverityWarning,
			Code:     "law_compliance",
			Document: doc.Name,
			Section:  doc.Section,
			Message:  fmt.Sprintf("verify compliance with %s", req.Law),
		})
	}

	return findings
}

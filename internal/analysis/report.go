package analysis

import (
	"fmt"
	"sort"
)

// Severity describes how critical a finding is.
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
)

// Finding is a single detected issue in the application package.
type Finding struct {
	Severity Severity
	Code     string
	Document string
	Section  string
	Message  string
}

// Report is the result of analyzing an application package.
type Report struct {
	Findings []Finding
}

// Errors returns findings with error severity.
func (r Report) Errors() []Finding {
	return r.filter(SeverityError)
}

// Warnings returns findings with warning severity.
func (r Report) Warnings() []Finding {
	return r.filter(SeverityWarning)
}

func (r Report) filter(s Severity) []Finding {
	out := make([]Finding, 0, len(r.Findings))
	for _, f := range r.Findings {
		if f.Severity == s {
			out = append(out, f)
		}
	}
	return out
}

// String renders the report in a human-readable form.
func (r Report) String() string {
	if len(r.Findings) == 0 {
		return "no issues found"
	}
	findings := make([]Finding, len(r.Findings))
	copy(findings, r.Findings)
	sort.SliceStable(findings, func(i, j int) bool {
		if findings[i].Document != findings[j].Document {
			return findings[i].Document < findings[j].Document
		}
		return findings[i].Section < findings[j].Section
	})
	var b []byte
	for _, f := range findings {
		b = append(b, fmt.Sprintf("[%s] %s %s/%s: %s\n", f.Severity, f.Code, f.Document, f.Section, f.Message)...)
	}
	return string(b)
}

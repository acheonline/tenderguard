package analysis

import (
	"context"
	"testing"
	"time"
)

func TestAnalyze(t *testing.T) {
	now := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		pkg        Package
		wantCodes  []string
		wantErr    bool
	}{
		{
			name: "all documents present and valid",
			pkg: Package{
				CustomerINN: "7804678913",
				Now:         now,
				Documents: []Document{
					{Name: "form_1", Title: "Форма 1", Fields: map[string]string{"name": "ООО Ромашка"}, Content: "44-ФЗ"},
				},
				Requirements: []Requirement{
					{DocumentName: "form_1", Title: "Форма 1", RequiredFields: []string{"name"}, MustContain: []string{"44-ФЗ"}, Law: "44-ФЗ"},
				},
			},
			wantCodes: []string{"law_compliance"},
		},
		{
			name: "missing document",
			pkg: Package{
				CustomerINN: "7841312071",
				Now:         now,
				Requirements: []Requirement{
					{DocumentName: "form_2", Title: "Форма 2"},
				},
			},
			wantCodes: []string{"missing_document"},
		},
		{
			name: "empty field",
			pkg: Package{
				CustomerINN: "7838129301",
				Now:         now,
				Documents: []Document{
					{Name: "form_1", Fields: map[string]string{"name": "  "}},
				},
				Requirements: []Requirement{
					{DocumentName: "form_1", RequiredFields: []string{"name"}},
				},
			},
			wantCodes: []string{"empty_field"},
		},
		{
			name: "stale certificate",
			pkg: Package{
				CustomerINN: "7804678913",
				Now:         now,
				Documents: []Document{
					{Name: "cert", IssuedAt: now.AddDate(0, 0, -40)},
				},
				Requirements: []Requirement{
					{DocumentName: "cert", MaxAgeDays: 30},
				},
			},
			wantCodes: []string{"stale_certificate"},
		},
		{
			name: "missing issue date",
			pkg: Package{
				CustomerINN: "7804678913",
				Now:         now,
				Documents: []Document{
					{Name: "cert"},
				},
				Requirements: []Requirement{
					{DocumentName: "cert", MaxAgeDays: 30},
				},
			},
			wantCodes: []string{"missing_issue_date"},
		},
		{
			name: "missing content",
			pkg: Package{
				CustomerINN: "7804678913",
				Now:         now,
				Documents: []Document{
					{Name: "form_1", Content: "nothing"},
				},
				Requirements: []Requirement{
					{DocumentName: "form_1", MustContain: []string{"КПП"}},
				},
			},
			wantCodes: []string{"missing_content"},
		},
		{
			name: "empty customer inn",
			pkg: Package{
				Now: now,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report, err := Analyze(context.Background(), tt.pkg)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := make([]string, 0, len(report.Findings))
			for _, f := range report.Findings {
				got = append(got, f.Code)
			}
			if len(got) != len(tt.wantCodes) {
				t.Fatalf("got codes %v, want %v", got, tt.wantCodes)
			}
			for i := range got {
				if got[i] != tt.wantCodes[i] {
					t.Fatalf("got codes %v, want %v", got, tt.wantCodes)
				}
			}
		})
	}
}

func TestAnalyzeContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := Analyze(ctx, Package{CustomerINN: "7804678913"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

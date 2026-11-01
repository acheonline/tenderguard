package analysis

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"testing"
)

const complianceDocxXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
    <w:p><w:r><w:t>Участник не находится в процессе ликвидации</w:t></w:r></w:p>
    <w:p><w:r><w:t>Деятельность не приостановлена</w:t></w:r></w:p>
    <w:p><w:r><w:t>Недоимка по налогам отсутствует</w:t></w:r></w:p>
    <w:p><w:r><w:t>Сведения в реестре недобросовестных поставщиков отсутствуют</w:t></w:r></w:p>
    <w:p><w:r><w:t>Судимость отсутствует</w:t></w:r></w:p>
    <w:p><w:r><w:t>Конфликт интересов отсутствует</w:t></w:r></w:p>
    <w:p><w:r><w:t>Копия свидетельства о государственной регистрации</w:t></w:r></w:p>
    <w:p><w:r><w:t>Выписка из ЕГРЮЛ</w:t></w:r></w:p>
    <w:p><w:r><w:t>Копии учредительных документов</w:t></w:r></w:p>
    <w:p><w:r><w:t>Копия доверенности</w:t></w:r></w:p>
    <w:p><w:r><w:t>Свидетельство о постановке на учет в налоговом органе</w:t></w:r></w:p>
    <w:p><w:r><w:t>Справка об исполнении обязанностей</w:t></w:r></w:p>
    <w:p><w:r><w:t>Банковская гарантия</w:t></w:r></w:p>
    <w:p><w:r><w:t>Форма 1 — Письмо о подаче заявки</w:t></w:r></w:p>
    <w:p><w:r><w:t>Форма 2 — Анкета участника</w:t></w:r></w:p>
    <w:p><w:r><w:t>Форма 3 — Декларация соответствия</w:t></w:r></w:p>
    <w:p><w:r><w:t>Форма 12 — Ценовое предложение</w:t></w:r></w:p>
    <w:p><w:r><w:t>Форма 13 — Предложение об объекте закупки</w:t></w:r></w:p>
  </w:body>
</w:document>`

func writeComplianceDocx(t *testing.T, path, body string) {
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
	if _, err := w.Write([]byte(body)); err != nil {
		t.Fatalf("write entry: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
}

func TestParseDocxCompliance(t *testing.T) {
	path := filepath.Join(t.TempDir(), "requirements.docx")
	writeComplianceDocx(t, path, complianceDocxXML)

	pkg, err := ParseDocx(context.Background(), path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkg.Documents) == 0 {
		t.Fatal("expected documents, got none")
	}
	if pkg.Documents[0].Name != "paragraph_1" {
		t.Errorf("first doc name = %q, want paragraph_1", pkg.Documents[0].Name)
	}
}

func TestAnalyzeComplianceComplete(t *testing.T) {
	path := filepath.Join(t.TempDir(), "requirements.docx")
	writeComplianceDocx(t, path, complianceDocxXML)

	pkg, err := ParseDocx(context.Background(), path)
	if err != nil {
		t.Fatalf("parse docx: %v", err)
	}
	report, err := AnalyzeCompliance(context.Background(), pkg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("expected no findings, got %v", report.Findings)
	}
}

func TestAnalyzeComplianceIncomplete(t *testing.T) {
	const incompleteXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
    <w:p><w:r><w:t>Форма 1 — Письмо о подаче заявки</w:t></w:r></w:p>
  </w:body>
</w:document>`

	path := filepath.Join(t.TempDir(), "requirements.docx")
	writeComplianceDocx(t, path, incompleteXML)

	pkg, err := ParseDocx(context.Background(), path)
	if err != nil {
		t.Fatalf("parse docx: %v", err)
	}
	report, err := AnalyzeCompliance(context.Background(), pkg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	codes := make(map[string]int)
	for _, f := range report.Findings {
		codes[f.Code]++
	}
	if codes["missing_requirement_confirmation"] == 0 {
		t.Error("expected missing_requirement_confirmation findings")
	}
	if codes["missing_required_document"] == 0 {
		t.Error("expected missing_required_document findings")
	}
	if codes["missing_form"] == 0 {
		t.Error("expected missing_form findings")
	}
}

func TestAnalyzeComplianceContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := AnalyzeCompliance(ctx, Package{}); err == nil {
		t.Fatal("expected error, got nil")
	}
}

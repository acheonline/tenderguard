package analysis

import (
	"context"
	"fmt"
	"strings"
)

// requirementCheck describes a confirmation that must appear in the application.
// It corresponds to a clause in section 6 of the procurement documentation.
var requirementChecks = []struct {
	Clause   string
	Title    string
	Keywords []string
}{
	{"6.1.2", "отсутствие ликвидации/банкротства", []string{"не находится в процессе ликвидации", "не признан несостоятельным", "банкрот"}},
	{"6.1.3", "отсутствие приостановки деятельности", []string{"не приостановлена", "приостановление деятельности"}},
	{"6.1.4", "отсутствие недоимки по налогам", []string{"недоимк", "задолженност"}},
	{"6.1.5", "отсутствие в РНП по 44-ФЗ", []string{"реестре недобросовестных поставщиков", "рнп"}},
	{"6.1.6", "отсутствие в РНП по 223-ФЗ", []string{"реестре недобросовестных поставщиков", "рнп"}},
	{"6.1.7", "отсутствие судимости у руководителя/главбуха", []string{"судимост"}},
	{"6.1.10", "отсутствие конфликта интересов", []string{"конфликт интересов", "конфликта интересов"}},
}

// requiredDocumentChecks describes documents that must be present in the package.
// It corresponds to section 7 of the procurement documentation.
var requiredDocumentChecks = []struct {
	Clause   string
	Title    string
	Keywords []string
}{
	{"7.1.1", "копия свидетельства о госрегистрации", []string{"свидетельств", "государственной регистрации"}},
	{"7.1.2", "выписка из ЕГРЮЛ/ЕГРИП", []string{"выписк", "егрюл", "егрип"}},
	{"7.1.3", "копии учредительных документов", []string{"учредительн"}},
	{"7.1.5", "копия доверенности", []string{"доверенност"}},
	{"7.1.6", "копия свидетельства о постановке на учет в налоговом органе", []string{"постановке на учет", "налоговом органе"}},
	{"7.1.8", "справка об исполнении налоговых обязанностей", []string{"справк", "исполнении обязанностей"}},
	{"7.1.9", "банковская гарантия", []string{"банковск", "гаранти"}},
}

// requiredFormChecks describes forms that must be filled in the application.
// It corresponds to section 8 of the procurement documentation.
var requiredFormChecks = []struct {
	Clause string
	Title  string
	Forms  []string
}{
	{"8.1", "Форма 1 — Письмо о подаче заявки", []string{"форма 1", "письмо о подаче заявки"}},
	{"8.2", "Форма 2 — Анкета участника", []string{"форма 2", "анкета участника"}},
	{"8.3", "Форма 3 — Декларация соответствия", []string{"форма 3", "декларация соответствия"}},
	{"8.4", "Форма 12 — Ценовое предложение", []string{"форма 12", "ценовое предложение"}},
	{"8.5", "Форма 13 — Предложение об объекте закупки", []string{"форма 13", "предложение об объекте закупки"}},
}

// AnalyzeCompliance checks the application package against sections 6, 7 and 8
// of the procurement documentation.
func AnalyzeCompliance(ctx context.Context, pkg Package) (Report, error) {
	if err := ctx.Err(); err != nil {
		return Report{}, fmt.Errorf("analyze compliance: %w", err)
	}

	var b strings.Builder
	for _, d := range pkg.Documents {
		b.WriteString(d.Content)
		b.WriteString("\n")
	}
	text := strings.ToLower(b.String())

	var findings []Finding
	for _, c := range requirementChecks {
		if !containsAny(text, c.Keywords) {
			findings = append(findings, Finding{
				Severity: SeverityError,
				Code:     "missing_requirement_confirmation",
				Section:  c.Clause,
				Message:  fmt.Sprintf("п. %s: не найдено подтверждение (%s)", c.Clause, c.Title),
			})
		}
	}
	for _, c := range requiredDocumentChecks {
		if !containsAny(text, c.Keywords) {
			findings = append(findings, Finding{
				Severity: SeverityError,
				Code:     "missing_required_document",
				Section:  c.Clause,
				Message:  fmt.Sprintf("п. %s: не найден документ (%s)", c.Clause, c.Title),
			})
		}
	}
	for _, c := range requiredFormChecks {
		if !containsAny(text, c.Forms) {
			findings = append(findings, Finding{
				Severity: SeverityError,
				Code:     "missing_form",
				Section:  c.Clause,
				Message:  fmt.Sprintf("п. %s: не найдена форма (%s)", c.Clause, c.Title),
			})
		}
	}

	return Report{Findings: findings}, nil
}

func containsAny(text string, keywords []string) bool {
	for _, k := range keywords {
		if strings.Contains(text, strings.ToLower(k)) {
			return true
		}
	}
	return false
}

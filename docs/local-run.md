# Локальный запуск

## Требования

- Go 1.22+ (проверить командой `go version`).
- Для проверки DOCX используются пакеты `archive/zip` и `encoding/xml` из стандартной библиотеки — внешние зависимости не требуются.

## Сборка

```bash
go build -o tenderguard ./cmd/tenderguard
```

## Запуск

С JSON-файлом:

```bash
./tenderguard path/to/input.json
```

С DOCX-файлом:

```bash
./tenderguard path/to/input.docx
```

## Формат input.json

JSON-файл соответствует структуре `analysis.Package`:

- `customer_inn` (string, обязательное) — ИНН заказчика.
- `now` (string, RFC3339) — текущая дата для проверки сроков.
- `documents` (array) — список документов, каждый с полями `name`, `title`, `content`, `issued_at`, `fields`.
- `requirements` (array) — список требований, каждое с полями `document_name`, `title`, `required_fields`, `must_contain`, `max_age_days`, `law`.

Пример валидного JSON:

```json
{
  "customer_inn": "7804678913",
  "now": "2024-06-01T00:00:00Z",
  "documents": [
    {
      "name": "form_1",
      "title": "Форма 1 — Письмо о подаче заявки",
      "content": "Участник подтверждает соответствие требованиям 44-ФЗ",
      "issued_at": "2024-05-01T00:00:00Z",
      "fields": {
        "name": "ООО Ромашка"
      }
    }
  ],
  "requirements": [
    {
      "document_name": "form_1",
      "title": "Форма 1",
      "required_fields": ["name"],
      "must_contain": ["44-ФЗ"],
      "max_age_days": 30,
      "law": "44-ФЗ"
    }
  ]
}
```

## Формат input.docx

Документ должен содержать текст заявки (требования, документы, формы).

Парсер извлекает абзацы и таблицы. Каждый абзац становится отдельным `Document` с полями `Name` (например, `paragraph_N`) и `Content` (текст). Таблицы извлекаются аналогично с именами `table_N`.

Анализатор ищет в тексте ключевые слова, в том числе:

- «лицензия»;
- «СРО»;
- «44-ФЗ»;
- «223-ФЗ».

## Тестовые данные

Примеры находятся в папке `testdata/`:

- `testdata/input.json` — минимальный валидный пакет.
- `testdata/input.docx` — минимальный DOCX с текстом.

В тестах (`cmd/tenderguard/main_test.go`) есть пример программной генерации DOCX через `archive/zip` — функцию `writeDocx` можно использовать как образец для создания собственных тестовых файлов.

## Проверка

```bash
go test -race ./...
golangci-lint run
```

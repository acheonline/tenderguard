package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

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
		return fmt.Errorf("usage: tenderguard <input.json>")
	}
	data, err := os.ReadFile(args[0])
	if err != nil {
		return fmt.Errorf("read input: %w", err)
	}
	var pkg analysis.Package
	if err := json.Unmarshal(data, &pkg); err != nil {
		return fmt.Errorf("parse input: %w", err)
	}
	report, err := analysis.Analyze(ctx, pkg)
	if err != nil {
		return fmt.Errorf("analyze: %w", err)
	}
	fmt.Print(report.String())
	return nil
}

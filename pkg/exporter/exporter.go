package exporter

import (
	"context"
	"fmt"
	"strings"
)

type FrontmatterReplaceFieldEntry struct {
	Field     string `json:"field" yaml:"field"`
	FieldType string `json:"field_type" yaml:"field_type"`
	Replace   string `json:"replace" yaml:"replace"`
}

type Config struct {
	DendronNotesPath        string                         `json:"dendron_notes_path" yaml:"dendron_notes_path"`
	ExportPath              string                         `json:"export_path" yaml:"export_path"`
	FrontmatterReplaceField []FrontmatterReplaceFieldEntry `json:"frontmatter_replace_field" yaml:"frontmatter_replace_field"`
}

type Exporter struct {
	cfg *Config
}

func New(cfg *Config) *Exporter {
	cfg.DendronNotesPath = strings.TrimRight(cfg.DendronNotesPath, "/") + "/"
	cfg.ExportPath = strings.TrimRight(cfg.ExportPath, "/") + "/"

	return &Exporter{
		cfg: cfg,
	}
}

func (e *Exporter) Run(ctx context.Context) error {
	if err := e.processNotes(ctx); err != nil {
		return fmt.Errorf("exporter: %w", err)
	}

	if err := e.fixFileHierarchy(ctx); err != nil {
		return fmt.Errorf("exporter: %w", err)
	}

	return nil
}

func (e *Exporter) processNotes(ctx context.Context) error {
	noteFnames, err := crawlNotes(ctx, e.cfg.DendronNotesPath)
	if err != nil {
		return fmt.Errorf("process notes: %w", err)
	}

	for _, fname := range noteFnames {
		if err := e.processNote(fname); err != nil {
			return fmt.Errorf("process notes: %w", err)
		}
	}

	return nil
}

func (e *Exporter) processNote(fname string) error {
	_, content, err := readFile(fname)
	if err != nil {
		return fmt.Errorf("process note %s: %w", fname, err)
	}

	if err := processLinkedAssets(content, e.cfg.DendronNotesPath, e.cfg.ExportPath); err != nil {
		return fmt.Errorf("process note %s: %w", fname, err)
	}

	content, err = dendronFlavourToObsidianFlavour(content)
	if err != nil {
		return fmt.Errorf("process note %s: %w", fname, err)
	}

	content, err = replaceFrontmatter(content, e.cfg.FrontmatterReplaceField)
	if err != nil {
		return fmt.Errorf("process note %s: %w", fname, err)
	}

	pub, err := checkIsPublished(fname)
	if err != nil {
		return fmt.Errorf("process note %s: %w", fname, err)
	}

	if !pub {
		return nil
	}

	newFname := strings.ReplaceAll(
		strings.TrimPrefix(strings.TrimRight(fname, ".md"), e.cfg.DendronNotesPath),
		".", "/") + ".md"

	if err := writeFile(e.cfg.ExportPath+newFname, content, 0777); err != nil {
		return fmt.Errorf("process note %s: %w", fname, err)
	}

	return nil
}

func (e *Exporter) fixFileHierarchy(_ context.Context) error {
	if err := moveFile(e.cfg.ExportPath+"root.md", e.cfg.ExportPath+"index.md"); err != nil {
		return fmt.Errorf("fix file hierarchy : %w", err)
	}

	return nil
}

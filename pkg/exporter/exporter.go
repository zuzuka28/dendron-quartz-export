package exporter

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/adrg/frontmatter"
	"gopkg.in/yaml.v2"
)

type FrontmatterRule struct {
	Field     string `json:"field" yaml:"field"`
	FieldType string `json:"field_type" yaml:"field_type"`
	Replace   string `json:"replace" yaml:"replace"`
}

type Config struct {
	DendronNotesPath string            `json:"dendron_notes_path" yaml:"dendron_notes_path"`
	ExportPath       string            `json:"export_path" yaml:"export_path"`
	FrontmatterRule  []FrontmatterRule `json:"frontmatter_replace_field" yaml:"frontmatter_replace_field"`
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

func (e *Exporter) Run(_ context.Context) error {
	fnames, err := crawlMarkdownFiles(e.cfg.DendronNotesPath)
	if err != nil {
		return fmt.Errorf("crawl markdown files: %w", err)
	}

	rawnotes, err := parseNotesToExport(fnames, publishFilter())
	if err != nil {
		return fmt.Errorf("parse notes to export: %w", err)
	}

	notes, err := processNotes(rawnotes, combineProcessor(
		frontmatterProcess(e.cfg.FrontmatterRule),
		dendronToObsidianFlavour(),
		moveLinkedAssetsProcess(e.cfg.DendronNotesPath, e.cfg.ExportPath),
		changePathPrefixProcess(e.cfg.DendronNotesPath, e.cfg.ExportPath),
		renameRootToIndexProcess(),
	))
	if err != nil {
		return fmt.Errorf("process notes: %w", err)
	}

	if err := writeNotes(notes); err != nil {
		return fmt.Errorf("write notes: %w", err)
	}

	return nil
}

func crawlMarkdownFiles(sourcePath string) ([]string, error) {
	var filePaths []string

	walkFunc := func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() && filepath.Ext(path) == ".md" {
			filePaths = append(filePaths, path)
		}

		return err
	}

	if err := filepath.Walk(sourcePath, walkFunc); err != nil {
		return nil, fmt.Errorf("crawl markdown: %w", err)
	}

	return filePaths, nil
}

func parseNote(fname string) (Note, error) {
	info, content, err := readFile(fname)
	if err != nil {
		return Note{}, fmt.Errorf("read note %s: %w", fname, err)
	}

	meta := make(map[string]any)

	content, err = frontmatter.Parse(bytes.NewReader(content), &meta)
	if err != nil {
		return Note{}, fmt.Errorf("parse frontmatter: %w", err)
	}

	return Note{
		DisplayName:    info.Name(),
		SourceFileInfo: info,
		Frontmatter:    meta,
		Content:        content,
		AllNotes:       []Note{},
	}, nil
}

func parseNotesToExport(fnames []string, filter FilterFunc) ([]Note, error) {
	allNotes := make([]Note, 0, len(fnames))

	for _, fname := range fnames {
		note, err := parseNote(fname)
		if err != nil {
			return nil, fmt.Errorf("parse note: %w", err)
		}

		allNotes = append(allNotes, note)
	}

	var filteredNotes []Note

	for _, v := range allNotes {
		if ok := filter(v); ok {
			filteredNotes = append(filteredNotes, v)
		}
	}

	for _, v := range filteredNotes {
		v.AllNotes = filteredNotes
	}

	return filteredNotes, nil
}

func processNotes(notes []Note, process ProcessFunc) ([]Note, error) {
	res := make([]Note, 0, len(notes))

	for _, v := range notes {
		note, err := process(v)
		if err != nil {
			return nil, fmt.Errorf("process note: %w", err)
		}

		res = append(res, note)
	}

	return res, nil
}

func writeNotes(notes []Note) error {
	toFileContent := func(note Note) ([]byte, error) {
		fmatter, err := yaml.Marshal(note.Frontmatter)
		if err != nil {
			return nil, fmt.Errorf("marshal frontmatter: %w", err)
		}

		return append([]byte("---"+"\n"+string(fmatter)+"---"+"\n"), note.Content...), nil
	}

	for _, note := range notes {
		data, err := toFileContent(note)
		if err != nil {
			return fmt.Errorf("make file content: %w", err)
		}

		if err := writeFile(note.DisplayName, data, note.SourceFileInfo.Mode()); err != nil {
			return fmt.Errorf("write note: %w", err)
		}
	}

	return nil
}

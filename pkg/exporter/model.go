package exporter

import "io/fs"

type Note struct {
	DisplayName    string
	SourceFileInfo fs.FileInfo

	Frontmatter map[string]any
	Content     []byte

	AllNotes []Note
}

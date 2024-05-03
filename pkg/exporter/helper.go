package exporter

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func extractWikilinkTarget(wikilink string) string {
	link := strings.TrimLeft(wikilink, "[")
	link = strings.TrimRight(link, "]")

	// dendron has links like [[title|link]], capture right side
	if splitted := strings.Split(link, "|"); len(splitted) == 2 {
		l := splitted[1]
		link = l
	}

	return link
}

func getFilename(in string) string {
	parts := strings.Split(in, "/")
	return parts[len(parts)-1]
}

func crawlNotes(_ context.Context, sourcePath string) ([]string, error) {
	var filePaths []string

	walkFunc := func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() && filepath.Ext(path) == ".md" {
			filePaths = append(filePaths, path)
		}

		return err
	}

	if err := filepath.Walk(sourcePath, walkFunc); err != nil {
		return nil, fmt.Errorf("crawl notes: %w", err)
	}

	return filePaths, nil
}

func readFile(path string) (os.FileInfo, []byte, error) {
	s, err := os.Stat(path)
	if err != nil {
		return nil, nil, fmt.Errorf("read file %s: %w", path, err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("read file %s: %w", path, err)
	}

	return s, data, nil
}

func moveFile(source, target string) error {
	if err := copyFile(source, target); err != nil {
		return fmt.Errorf("move file %s: %w", source, err)
	}

	if err := os.Remove(source); err != nil {
		return fmt.Errorf("move file %s: %w", source, err)
	}

	return nil
}

func copyFile(source, target string) error {
	_, data, err := readFile(source)
	if err != nil {
		return fmt.Errorf("copy file %s: %w", source, err)
	}

	if err := writeFile(target, data, 0777); err != nil {
		return fmt.Errorf("copy file %s: %w", source, err)
	}

	return nil
}

func writeFile(path string, data []byte, perm fs.FileMode) error {
	file, err := createFile(path, perm)
	if err != nil {
		return fmt.Errorf("write file %s: %w", path, err)
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("write file %s: %w", path, err)
	}

	return nil
}

func createFile(path string, perm fs.FileMode) (*os.File, error) {
	newDir := ""
	dirs := strings.Split(path, "/")

	if len(dirs) > 1 {
		newDir = strings.Join(dirs[:len(dirs)-1], "/")
	}

	if err := os.MkdirAll(newDir, perm); err != nil {
		return nil, fmt.Errorf("create  file %s: %w", path, err)
	}

	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("create file %s: %w", path, err)
	}

	return file, nil
}

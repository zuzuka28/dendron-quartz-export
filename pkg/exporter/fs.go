package exporter

import (
	"fmt"
	"io/fs"
	"os"
	"strings"
)

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

func copyFile(source, target string) error {
	_, data, err := readFile(source)
	if err != nil {
		return fmt.Errorf("copy file %s: %w", source, err)
	}

	if err := writeFile(target, data, 0777); err != nil { //nolint:mnd
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

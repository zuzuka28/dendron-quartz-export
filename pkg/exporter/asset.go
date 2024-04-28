package exporter

import (
	"fmt"
	"strings"
)

func processLinkedAssets(content []byte, sourcePath string, exportPath string) error {
	for _, wl := range wikilinkRe.FindAllString(string(content), -1) {
		link := extractWikilinkTarget(wl)

		if !strings.HasPrefix(link, "assets/") {
			continue
		}

		if err := copyFile(sourcePath+link, exportPath+link); err != nil {
			return fmt.Errorf("process linked asset %s: %w", link, err)
		}
	}

	return nil
}

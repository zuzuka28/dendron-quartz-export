package exporter

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/adrg/frontmatter"
	"gopkg.in/yaml.v2"
)

func checkIsPublished(path string) (bool, error) {
	type published struct {
		Published bool `yaml:"publish"`
	}

	_, f, err := readFile(path)
	if err != nil {
		return false, fmt.Errorf("check published: %w", err)
	}

	fm := new(published)
	if _, err = frontmatter.Parse(bytes.NewReader(f), fm); err != nil {
		return false, fmt.Errorf("check published: %w", err)
	}

	return fm.Published, nil
}

func replaceFrontmatter(content []byte, replaceFm []FrontmatterReplaceFieldEntry) ([]byte, error) {
	fm := make(map[string]any)

	content, err := frontmatter.Parse(bytes.NewReader(content), &fm)
	if err != nil {
		return nil, fmt.Errorf("replace frontmatter: %w", err)
	}

	for _, replaceEntry := range replaceFm {
		val, ok := fm[replaceEntry.Field]

		if !ok {
			continue
		}

		if replaceEntry.FieldType == "tags" || replaceEntry.FieldType == "links" {
			ri, ok := val.([]any)
			if !ok {
				return nil, fmt.Errorf("replace frontmatter: can't extract list")
			}

			items := make([]string, 0, len(ri))
			for _, v := range ri {
				items = append(items, strings.ReplaceAll(v.(string), ".", "/"))
			}

			val = items
		}

		if replaceEntry.FieldType == "timestamp" {
			ri, ok := val.(int)
			if !ok {
				return nil, fmt.Errorf("replace frontmatter: can't extract timestamp")
			}

			val = time.UnixMilli(int64(ri)).Format(time.RFC3339)
		}

		delete(fm, replaceEntry.Field)

		fm[replaceEntry.Replace] = val
	}

	mfm, err := yaml.Marshal(fm)
	if err != nil {
		return nil, fmt.Errorf("replace frontmatter: %w", err)
	}

	content = append([]byte("---"+"\n"+string(mfm)+"---"), content...)

	return content, nil
}

package exporter

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

var (
	wikilinkRe = regexp.MustCompile(`\[\[([^|\]#]+)(?:#[^\]]*)?(?:\|([^\]]+))?\]\]`)
	hashtagRe  = regexp.MustCompile(`#[\d\w_\.]+`)
	mentionRe  = regexp.MustCompile(`@[\d\w_\.]+`)
)

type ProcessFunc func(note Note) (Note, error)

func combineProcessor(processors ...ProcessFunc) ProcessFunc {
	return func(note Note) (Note, error) {
		var err error
		for _, process := range processors {
			note, err = process(note)
			if err != nil {
				return Note{}, err
			}
		}

		return note, nil
	}
}

func frontmatterProcess(rules []FrontmatterRule) ProcessFunc {
	return func(note Note) (Note, error) {
		for _, rule := range rules {
			val, ok := note.Frontmatter[rule.Field]
			if !ok {
				continue
			}

			switch rule.FieldType {
			case "tags", "links":
				ri, ok := val.([]any)
				if !ok {
					continue
				}

				items := make([]string, 0, len(ri))
				for _, v := range ri {
					items = append(items, strings.ReplaceAll(v.(string), ".", "/"))
				}

				val = items

			case "timestamp":
				ri, ok := val.(int)
				if !ok {
					continue
				}

				val = time.UnixMilli(int64(ri)).Format(time.RFC3339)
			}

			delete(note.Frontmatter, rule.Field)

			note.Frontmatter[rule.Replace] = val
		}

		return note, nil
	}
}

func dendronToObsidianFlavour() ProcessFunc {
	return func(note Note) (Note, error) {
		content := note.Content

		content = wikilinkRe.ReplaceAllFunc(content, func(b []byte) []byte {
			s := string(b)
			s = strings.TrimLeft(s, "[")
			s = strings.TrimRight(s, "]")

			splitted := strings.Split(s, "|")

			// dendron has links like [[title|link]], we need [[link|title]]
			switch len(splitted) {
			case 1:
				r := splitted[0]

				// it's raw link like 'assets/...'
				if strings.Contains(r, "/") {
					break
				}

				hierarcy := strings.Split(r, ".")
				l := strings.ReplaceAll(hierarcy[len(hierarcy)-1], "-", " ")

				splitted = []string{r, l}

			case 2: //nolint:mnd
				r, l := splitted[0], splitted[1]
				splitted = []string{l, r}
			}

			s = strings.Join(splitted, "|")

			return []byte("[[" + s + "]]")
		})

		content = hashtagRe.ReplaceAllFunc(content, func(b []byte) []byte {
			s := string(b)
			s = strings.TrimLeft(s, "#")
			s = strings.ReplaceAll(s, ".", "/")

			return []byte("#" + s)
		})

		content = mentionRe.ReplaceAllFunc(content, func(b []byte) []byte {
			s := string(b)
			s = strings.TrimLeft(s, "@")
			s = strings.ReplaceAll(s, ".", "/")

			return []byte("@" + s)
		})

		note.Content = content

		return note, nil
	}
}

func extractWikilinkTarget(wikilink string) string {
	link := strings.TrimLeft(wikilink, "[")
	link = strings.TrimRight(link, "]")

	// dendron has links like [[title|link]], capture right side
	if splitted := strings.Split(link, "|"); len(splitted) == 2 { //nolint:mnd
		l := splitted[1]
		link = l
	}

	return link
}

func moveLinkedAssetsProcess(sourcePath string, targetPath string) ProcessFunc {
	return func(note Note) (Note, error) {
		for _, wl := range wikilinkRe.FindAllString(string(note.Content), -1) {
			link := extractWikilinkTarget(wl)

			if !strings.HasPrefix(link, "assets/") {
				continue
			}

			if err := copyFile(sourcePath+link, targetPath+link); err != nil {
				return Note{}, fmt.Errorf("process linked asset %s: %w", link, err)
			}
		}

		return note, nil
	}
}

func changePathPrefixProcess(sourcePath string, targetPath string) ProcessFunc {
	return func(note Note) (Note, error) {
		note.DisplayName = targetPath + strings.TrimPrefix(note.DisplayName, sourcePath)

		return note, nil
	}
}

func renameRootToIndexProcess() ProcessFunc {
	return func(note Note) (Note, error) {
		note.DisplayName = strings.ReplaceAll(note.DisplayName, "root.md", "index.md")

		return note, nil
	}
}

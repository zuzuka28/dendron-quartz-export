package exporter

import (
	"regexp"
	"strings"
)

var (
	wikilinkRe = regexp.MustCompile(`\[\[([^|\]#]+)(?:#[^\]]*)?(?:\|([^\]]+))?\]\]`)
	hashtagRe  = regexp.MustCompile(`#[\d\w_\.]+`)
	mentionRe  = regexp.MustCompile(`@[\d\w_\.]+`)
)

func dendronFlavourToObsidianFlavour(content []byte) ([]byte, error) {
	content = wikilinkRe.ReplaceAllFunc(content, func(b []byte) []byte {
		s := string(b)
		s = strings.TrimLeft(s, "[")
		s = strings.TrimRight(s, "]")

		// assets has raw links that should not be converted
		if strings.HasPrefix(s, "assets/") {
			return b
		}

		splitted := strings.Split(s, "|")

		// replace dot notation
		if len(splitted) == 1 {
			splitted = []string{strings.ReplaceAll(splitted[0], ".", "/")}
		}

		// dendron has links like [[title|link]], we need [[link|title]]
		if len(splitted) == 2 {
			r, l := splitted[0], splitted[1]
			l = strings.ReplaceAll(l, ".", "/")
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

	return content, nil
}

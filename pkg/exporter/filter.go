package exporter

type FilterFunc func(note Note) bool

func publishFilter() FilterFunc {
	return func(note Note) bool {
		val, ok := note.Frontmatter["publish"]
		if !ok {
			return false
		}

		flag, ok := val.(bool)
		if !ok {
			return false
		}

		return flag
	}
}

package misc

// check if v has duplicated elements
func HasDupElems[T comparable](v []T) bool {
	m := make(map[T]struct{})
	for _, e := range v {
		if _, ok := m[e]; ok {
			return true
		}
		m[e] = struct{}{}
	}
	return false
}

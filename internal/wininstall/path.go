package wininstall

import "strings"

func normalizeDir(dir string) string {
	return strings.TrimRight(strings.TrimSpace(dir), `/\`)
}

func splitPath(pathList string) []string {
	var out []string
	for _, p := range strings.Split(pathList, ";") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}

// PathContains reports whether dir is already an entry in a Windows PATH list.
func PathContains(pathList, dir string) bool {
	want := normalizeDir(dir)
	if want == "" {
		return false
	}
	for _, p := range splitPath(pathList) {
		if strings.EqualFold(normalizeDir(p), want) {
			return true
		}
	}
	return false
}

// AppendPath adds dir to a Windows PATH list if it is not already present.
func AppendPath(pathList, dir string) string {
	dir = strings.TrimSpace(dir)
	if dir == "" || PathContains(pathList, dir) {
		return pathList
	}
	parts := splitPath(pathList)
	parts = append(parts, dir)
	return strings.Join(parts, ";")
}

// RemovePath drops dir from a Windows PATH list (case-insensitive).
func RemovePath(pathList, dir string) string {
	want := normalizeDir(dir)
	if want == "" {
		return pathList
	}
	var kept []string
	for _, p := range splitPath(pathList) {
		if strings.EqualFold(normalizeDir(p), want) {
			continue
		}
		kept = append(kept, p)
	}
	return strings.Join(kept, ";")
}

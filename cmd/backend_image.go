package cmd

import "strings"

// resolveBackendImageRef picks the backend image URL/tag from either a combined ref or separate URL/tag.
// If ref contains a ":", the portion after the last colon is treated as the tag.
func resolveBackendImageRef(ref, url, tag string) (string, string) {
	ref = strings.TrimSpace(ref)
	if ref != "" {
		if idx := strings.LastIndex(ref, ":"); idx != -1 && idx != len(ref)-1 {
			return strings.TrimSpace(ref[:idx]), strings.TrimSpace(ref[idx+1:])
		}
		return ref, ""
	}
	return strings.TrimSpace(url), strings.TrimSpace(tag)
}

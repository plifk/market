package webpages

import (
	"path/filepath"
	"testing"
)

// TestTemplateFilesDuplication enforces that we don't add a file with the same name on the web directory by accident,
// as template.ParseFiles discards files with duplicated names in different directories.
func TestTemplateFilesDuplication(t *testing.T) {
	dedup := map[string]struct{}{}
	for _, tf := range templateFiles {
		d := filepath.Base(tf)
		if _, ok := dedup[d]; ok {
			t.Errorf("found files with same name in two locations: %q and %q", tf, d)
		}
	}
}

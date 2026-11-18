package web

import (
	"io/fs"
	"regexp"
	"strings"
	"testing"
)

// assetRefPattern matches absolute /assets/... references emitted by the
// Vite build (HTML markup and string literals inside the bundled JS).
var assetRefPattern = regexp.MustCompile(`/assets/[A-Za-z0-9._/-]+`)

// TestEmbeddedAssetsResolve guards against shipping a dashboard whose
// stylesheet or images 404 inside the packaged binary. A stale asset
// reference silently degrades the whole app window (unstyled markup), so
// every referenced file must exist in the embedded filesystem.
func TestEmbeddedAssetsResolve(t *testing.T) {
	refs := map[string]struct{}{}

	index, err := EmbeddedFS.ReadFile("index.html")
	if err != nil {
		t.Fatalf("read embedded index.html: %v", err)
	}
	for _, ref := range assetRefPattern.FindAllString(string(index), -1) {
		refs[strings.TrimPrefix(ref, "/")] = struct{}{}
	}

	entries, err := fs.ReadDir(EmbeddedFS, "assets")
	if err != nil {
		t.Fatalf("read embedded assets dir: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".js") {
			continue
		}
		bundle, err := EmbeddedFS.ReadFile("assets/" + entry.Name())
		if err != nil {
			t.Fatalf("read embedded bundle %s: %v", entry.Name(), err)
		}
		for _, ref := range assetRefPattern.FindAllString(string(bundle), -1) {
			refs[strings.TrimPrefix(ref, "/")] = struct{}{}
		}
	}

	for ref := range refs {
		if _, err := fs.Stat(EmbeddedFS, ref); err != nil {
			t.Errorf("referenced asset %q is missing from the embedded filesystem: %v", ref, err)
		}
	}
}

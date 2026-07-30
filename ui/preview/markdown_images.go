package preview

import (
	"os"
	"path/filepath"
	"strings"
)

const previewPlaceholderImage = "screenshot000.png"

// previewImagePath maps remote URLs to a bundled local screenshot until attachments exist.
func previewImagePath(url string) string {
	url = strings.TrimSpace(url)
	if isHTTPURL(url) {
		return previewPlaceholderImage
	}
	return url
}

// previewAssetBases are working-directory candidates (same idea as the markdown fixture loader).
func previewAssetBases() []string {
	wd, err := os.Getwd()
	if err != nil {
		return []string{"."}
	}
	return []string{
		wd,
		filepath.Join(wd, ".."),
		filepath.Join(wd, "..", ".."),
	}
}

// resolvePreviewImagePath returns the first path that exists on disk.
func resolvePreviewImagePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return path
	}
	for _, base := range previewAssetBases() {
		for _, p := range []string{
			path,
			filepath.Join(base, path),
		} {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	}
	return path
}

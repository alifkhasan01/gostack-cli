// Package updater checks and applies CLI self-updates from GitHub Releases.
package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
)

const (
	githubRepo  = "alifkhasan01/gostack-cli"
	releasesAPI = "https://api.github.com/repos/" + githubRepo + "/releases/latest"
)

// Release represents a GitHub release.
type Release struct {
	TagName string  `json:"tag_name"` // e.g. "v0.2.0"
	Body    string  `json:"body"`
	Assets  []Asset `json:"assets"`
}

// Asset is a single file attached to a GitHub release.
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// CheckLatest fetches the latest release info from GitHub.
func CheckLatest() (*Release, error) {
	req, err := http.NewRequest(http.MethodGet, releasesAPI, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("check update: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("no releases found for %s", githubRepo)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: HTTP %d", resp.StatusCode)
	}

	var r Release
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("parse release: %w", err)
	}
	return &r, nil
}

// IsNewer returns true if latestTag is a higher version than currentVersion.
// Both are expected in "v0.1.0" or "0.1.0" format.
func IsNewer(currentVersion, latestTag string) bool {
	cur := normalizeVersion(currentVersion)
	lat := normalizeVersion(latestTag)
	if cur == "dev" || cur == "" {
		return false // dev build — never auto-update
	}
	return compareSemver(lat, cur) > 0
}

// compareSemver returns 1 if a > b, -1 if a < b, 0 if equal.
func compareSemver(a, b string) int {
	ap := parseSemver(a)
	bp := parseSemver(b)
	for i := range ap {
		if i >= len(bp) {
			break
		}
		if ap[i] > bp[i] {
			return 1
		}
		if ap[i] < bp[i] {
			return -1
		}
	}
	return 0
}

func parseSemver(v string) [3]int {
	var major, minor, patch int
	fmt.Sscanf(v, "%d.%d.%d", &major, &minor, &patch)
	return [3]int{major, minor, patch}
}

// PlatformAssetName returns the expected asset name for the current OS/arch.
// Matches GoReleaser's default naming: gostack_linux_amd64, gostack_darwin_arm64, etc.
func PlatformAssetName() string {
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	name := fmt.Sprintf("gostack_%s_%s", goos, goarch)
	if goos == "windows" {
		name += ".exe"
	}
	return name
}

// DownloadAsset downloads the given URL into a temp file and returns its path.
func DownloadAsset(url string) (string, error) {
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return "", fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	tmp, err := os.CreateTemp("", "gostack-update-*")
	if err != nil {
		return "", err
	}
	defer tmp.Close()

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		os.Remove(tmp.Name())
		return "", err
	}
	return tmp.Name(), nil
}

// ReplaceCurrentBinary replaces the running binary with the file at newBinPath.
func ReplaceCurrentBinary(newBinPath string) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("find current binary: %w", err)
	}

	// Make the new binary executable
	if err := os.Chmod(newBinPath, 0755); err != nil {
		return err
	}

	// On Linux/macOS we can rename over the running binary
	if err := os.Rename(newBinPath, exe); err != nil {
		// Fallback: copy
		return copyFile(newBinPath, exe)
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return os.Chmod(dst, 0755)
}

func normalizeVersion(v string) string {
	v = strings.TrimPrefix(v, "v")
	return v
}

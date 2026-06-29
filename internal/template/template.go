package template

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	// GitHub org where templates live.
	githubOrg = "gostack-templates"
	// Base URL for downloading a repo as zip from GitHub.
	githubZipURL = "https://codeload.github.com/%s/%s/zip/refs/heads/main"
)

// Manifest represents manifest.json inside a template repo.
type Manifest struct {
	Name         string `json:"name"`
	Framework    string `json:"framework"`
	Architecture string `json:"architecture"`
	Version      string `json:"version"`
	GoVersion    string `json:"go"`
}

// RepoName builds the template repository name from framework and architecture.
// e.g. gin + clean → "template-gin-clean"
func RepoName(framework, architecture string) string {
	return fmt.Sprintf("template-%s-%s", framework, architecture)
}

// Download fetches the template repository zip and extracts the inner
// "template/" directory into destDir.
func Download(framework, architecture, destDir string) error {
	repo := RepoName(framework, architecture)
	url := fmt.Sprintf(githubZipURL, githubOrg, repo)

	fmt.Printf("  Downloading template from github.com/%s/%s ...\n", githubOrg, repo)

	// Download zip into a temp file
	tmpFile, err := os.CreateTemp("", "gostack-*.zip")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return fmt.Errorf("download template: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("template not found: %s (HTTP %d). "+
			"Available templates: gin-clean, gin-standard, fiber-clean, fiber-standard, echo-clean, chi-clean",
			repo, resp.StatusCode)
	}

	if _, err = io.Copy(tmpFile, resp.Body); err != nil {
		return fmt.Errorf("write zip: %w", err)
	}
	tmpFile.Close()

	// Extract
	return extractTemplate(tmpFile.Name(), repo, destDir)
}

// extractTemplate unzips the downloaded archive and copies the "template/"
// subdirectory to destDir.
func extractTemplate(zipPath, repo, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}
	defer r.Close()

	// GitHub zips have a top-level folder like "template-gin-clean-main/"
	topPrefix := repo + "-main/"
	templatePrefix := topPrefix + "template/"

	for _, f := range r.File {
		// Only files inside "template/" subfolder
		if !strings.HasPrefix(f.Name, templatePrefix) {
			continue
		}

		// Relative path inside destDir
		rel := strings.TrimPrefix(f.Name, templatePrefix)
		if rel == "" {
			continue
		}

		targetPath := filepath.Join(destDir, rel)

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return err
			}
			continue
		}

		if err := extractFile(f, targetPath); err != nil {
			return err
		}
	}

	return nil
}

func extractFile(f *zip.File, destPath string) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}

	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, rc) //nolint:gosec
	return err
}

// FetchManifest downloads and parses manifest.json from a template repo.
func FetchManifest(framework, architecture string) (*Manifest, error) {
	repo := RepoName(framework, architecture)
	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/main/manifest.json",
		githubOrg, repo)

	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("manifest not found for %s", repo)
	}

	var m Manifest
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, err
	}
	return &m, nil
}

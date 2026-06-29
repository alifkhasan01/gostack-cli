package replacer

import (
	"os"
	"path/filepath"
	"strings"
)

// Placeholder keys used inside templates.
const (
	PlaceholderModule  = "{{MODULE_NAME}}"
	PlaceholderProject = "{{PROJECT_NAME}}"
	PlaceholderApp     = "{{APP_NAME}}"
)

// Config holds the values that will replace the placeholders.
type Config struct {
	ModuleName  string
	ProjectName string
}

// ReplaceAll walks destDir and replaces all placeholders in every file.
func ReplaceAll(destDir string, cfg Config) error {
	return filepath.Walk(destDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		return replaceInFile(path, cfg)
	})
}

func replaceInFile(path string, cfg Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	content := string(data)
	content = strings.ReplaceAll(content, PlaceholderModule, cfg.ModuleName)
	content = strings.ReplaceAll(content, PlaceholderProject, cfg.ProjectName)
	content = strings.ReplaceAll(content, PlaceholderApp, cfg.ProjectName)

	return os.WriteFile(path, []byte(content), info(path).Mode())
}

func info(path string) os.FileInfo {
	fi, _ := os.Stat(path)
	return fi
}

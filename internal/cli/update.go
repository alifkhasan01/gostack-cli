package cli

import (
	"fmt"

	"github.com/alifkhasan01/gostack-cli/internal/printer"
	"github.com/alifkhasan01/gostack-cli/internal/updater"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Check for a newer version of GoStack CLI and self-update",
	RunE:  runUpdate,
}

var checkOnly bool

func init() {
	updateCmd.Flags().BoolVar(&checkOnly, "check", false, "Only check for updates, do not install")
	rootCmd.AddCommand(updateCmd)
}

func runUpdate(_ *cobra.Command, _ []string) error {
	sp := printer.NewSpinner("Checking for updates ...")
	sp.Start()

	release, err := updater.CheckLatest()
	if err != nil {
		sp.Fail("Could not reach GitHub")
		return fmt.Errorf("%w\nCheck your internet connection or visit: https://github.com/alifkhasan01/gostack-cli/releases", err)
	}
	sp.Stop()

	latest := release.TagName
	printer.Dim(fmt.Sprintf("Current  : %s", Version))
	printer.Dim(fmt.Sprintf("Latest   : %s", latest))

	if !updater.IsNewer(Version, latest) {
		printer.Success(fmt.Sprintf("Already up to date (%s)", Version))
		return nil
	}

	fmt.Printf("\n")
	printer.Step("🆕", fmt.Sprintf("New version available: %s", latest))
	if release.Body != "" {
		printer.Dim("─── Release Notes ───────────────────────")
		for _, line := range splitLines(release.Body) {
			printer.Dim(line)
		}
		printer.Dim("─────────────────────────────────────────")
	}

	if checkOnly {
		fmt.Printf("\n  Run %s to install.\n", "`gostack update`")
		return nil
	}

	// Find asset for this platform
	assetName := updater.PlatformAssetName()
	var downloadURL string
	for _, a := range release.Assets {
		if a.Name == assetName {
			downloadURL = a.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf(
			"no binary for %s in release %s\nDownload manually: https://github.com/alifkhasan01/gostack-cli/releases",
			assetName, latest,
		)
	}

	sp2 := printer.NewSpinner(fmt.Sprintf("Downloading %s ...", assetName))
	sp2.Start()
	tmpPath, err := updater.DownloadAsset(downloadURL)
	if err != nil {
		sp2.Fail("Download failed")
		return fmt.Errorf("download: %w", err)
	}
	sp2.Done("Downloaded")

	sp3 := printer.NewSpinner("Installing update ...")
	sp3.Start()
	if err := updater.ReplaceCurrentBinary(tmpPath); err != nil {
		sp3.Fail("Install failed — try: sudo gostack update")
		return fmt.Errorf("install: %w", err)
	}
	sp3.Done(fmt.Sprintf("Updated to %s", latest))

	printer.Dim("Restart your shell if the version doesn't change immediately.")
	return nil
}

func splitLines(s string) []string {
	var lines []string
	current := ""
	for _, c := range s {
		if c == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

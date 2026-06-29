package cli

import (
	"fmt"

	"github.com/gostack/cli/internal/printer"
	"github.com/gostack/cli/internal/updater"
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
	printer.Step("🔍", "Checking for updates ...")

	release, err := updater.CheckLatest()
	if err != nil {
		return fmt.Errorf("could not reach GitHub: %w\nCheck your internet connection and try again.", err)
	}

	latest := release.TagName
	printer.Dim(fmt.Sprintf("Current version : %s", Version))
	printer.Dim(fmt.Sprintf("Latest version  : %s", latest))

	if !updater.IsNewer(Version, latest) {
		printer.Success(fmt.Sprintf("You are already on the latest version (%s)", Version))
		return nil
	}

	fmt.Printf("\n  🆕 New version available: %s\n", latest)
	if release.Body != "" {
		printer.Dim("Release notes:")
		printer.Dim(release.Body)
	}

	if checkOnly {
		fmt.Printf("\n  Run %s to install the update.\n", "`gostack update`")
		return nil
	}

	// Find the right asset for this platform
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
			"no binary found for %s in release %s\n"+
				"Download manually from: https://github.com/gostack/cli/releases",
			assetName, latest,
		)
	}

	printer.Step("⬇️ ", fmt.Sprintf("Downloading %s ...", assetName))
	tmpPath, err := updater.DownloadAsset(downloadURL)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	printer.Step("🔄", "Installing update ...")
	if err := updater.ReplaceCurrentBinary(tmpPath); err != nil {
		return fmt.Errorf("install failed: %w\nTry: sudo gostack update", err)
	}

	printer.Success(fmt.Sprintf("Updated to %s — restart your shell if needed.", latest))
	return nil
}

package cli

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"repoctr/internal/version"
)

// httpClient is a shared HTTP client with reasonable timeouts.
var httpClient = &http.Client{
	Timeout: 60 * time.Second,
}

// allowedDownloadHosts contains the valid hosts for binary downloads.
var allowedDownloadHosts = []string{
	"https://github.com/",
	"https://objects.githubusercontent.com/",
}

// githubRelease represents a GitHub release from the API.
type githubRelease struct {
	TagName     string        `json:"tag_name"`
	Name        string        `json:"name"`
	Body        string        `json:"body"`
	Draft       bool          `json:"draft"`
	Prerelease  bool          `json:"prerelease"`
	PublishedAt string        `json:"published_at"`
	Assets      []githubAsset `json:"assets"`
}

// githubAsset represents a release asset (binary file).
type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// NewUpdateCmd creates the update command.
func NewUpdateCmd() *cobra.Command {
	var forceUpdate bool
	var checkOnly bool
	var skipChecksum bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Check for updates and upgrade repo-ctr to the latest version",
		Long: `Checks GitHub for new releases and displays release notes for all
versions since your current version. If updates are available, prompts
to download and install the latest version.

Use --check to only check for updates without installing.
Use --force to update even if already on the latest version.
Use --skip-checksum to skip SHA256 verification (not recommended).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(forceUpdate, checkOnly, skipChecksum)
		},
	}

	cmd.Flags().BoolVarP(&forceUpdate, "force", "f", false, "Force update even if already on latest version")
	cmd.Flags().BoolVarP(&checkOnly, "check", "c", false, "Only check for updates, don't install")
	cmd.Flags().BoolVar(&skipChecksum, "skip-checksum", false, "Skip SHA256 checksum verification (not recommended)")

	return cmd
}

func runUpdate(forceUpdate, checkOnly, skipChecksum bool) error {
	currentVersion := version.Version

	fmt.Printf("Current version: %s\n", currentVersion)
	fmt.Println("Checking for updates...")

	// Fetch releases from GitHub
	releases, err := fetchReleases()
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if len(releases) == 0 {
		fmt.Println("No releases found.")
		return nil
	}

	// Filter to stable releases only (no drafts or prereleases)
	var stableReleases []githubRelease
	for _, r := range releases {
		if !r.Draft && !r.Prerelease {
			stableReleases = append(stableReleases, r)
		}
	}

	if len(stableReleases) == 0 {
		fmt.Println("No stable releases found.")
		return nil
	}

	// Sort releases by version (newest first for finding latest, but we'll reverse for display)
	sortReleasesByVersion(stableReleases)

	latestRelease := stableReleases[0]
	latestVersion := latestRelease.TagName

	// Find releases newer than current version
	newerReleases := findNewerReleases(stableReleases, currentVersion)

	if len(newerReleases) == 0 && !forceUpdate {
		fmt.Printf("\nYou are already on the latest version (%s).\n", latestVersion)
		return nil
	}

	if len(newerReleases) > 0 {
		fmt.Printf("\nNew version available: %s\n", latestVersion)
		fmt.Println("\n" + strings.Repeat("=", 60))
		fmt.Println("RELEASE NOTES")
		fmt.Println(strings.Repeat("=", 60))

		// Display release notes from oldest to newest
		for i := len(newerReleases) - 1; i >= 0; i-- {
			r := newerReleases[i]
			displayReleaseNotes(r)
		}
		fmt.Println(strings.Repeat("=", 60))
	} else if forceUpdate {
		fmt.Printf("\nForce updating to %s...\n", latestVersion)
	}

	if checkOnly {
		if len(newerReleases) > 0 {
			fmt.Printf("\nRun 'repo-ctr update' to install version %s.\n", latestVersion)
		}
		return nil
	}

	// Find the appropriate asset for this OS/arch
	asset := findAssetForPlatform(latestRelease.Assets)
	if asset == nil {
		return fmt.Errorf("no binary available for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	// Find the checksum file
	checksumAsset := findChecksumAsset(latestRelease.Assets)

	// Prompt for confirmation
	if !promptConfirm(fmt.Sprintf("Update to %s?", latestVersion)) {
		fmt.Println("Update cancelled.")
		return nil
	}

	// Download and install
	fmt.Printf("\nDownloading %s...\n", asset.Name)
	if err := downloadAndInstall(asset, checksumAsset, skipChecksum); err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}

	fmt.Printf("\nSuccessfully updated to %s!\n", latestVersion)
	return nil
}

func fetchReleases() ([]githubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", version.GitHubOwner, version.GitHubRepo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "repo-ctr/"+version.Version)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var releases []githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, err
	}

	return releases, nil
}

// sortReleasesByVersion sorts releases by semantic version (newest first).
func sortReleasesByVersion(releases []githubRelease) {
	sort.Slice(releases, func(i, j int) bool {
		return compareVersions(releases[i].TagName, releases[j].TagName) > 0
	})
}

// findNewerReleases returns releases newer than the current version (newest first).
func findNewerReleases(releases []githubRelease, currentVersion string) []githubRelease {
	var newer []githubRelease
	for _, r := range releases {
		if compareVersions(r.TagName, currentVersion) > 0 {
			newer = append(newer, r)
		}
	}
	return newer
}

// compareVersions compares two version strings.
// Returns: 1 if v1 > v2, -1 if v1 < v2, 0 if equal.
func compareVersions(v1, v2 string) int {
	// Strip 'v' prefix if present
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")

	// Split by '.' and also handle prerelease suffixes like "-beta"
	v1 = strings.Split(v1, "-")[0] // Remove any prerelease suffix
	v2 = strings.Split(v2, "-")[0]

	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	// Compare each part
	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var n1, n2 int
		if i < len(parts1) {
			n1, _ = strconv.Atoi(parts1[i])
		}
		if i < len(parts2) {
			n2, _ = strconv.Atoi(parts2[i])
		}

		if n1 > n2 {
			return 1
		}
		if n1 < n2 {
			return -1
		}
	}

	return 0
}

func displayReleaseNotes(r githubRelease) {
	fmt.Printf("\n## %s", r.TagName)
	if r.Name != "" && r.Name != r.TagName {
		fmt.Printf(" - %s", r.Name)
	}
	fmt.Println()

	if r.PublishedAt != "" {
		// Parse and format date nicely
		date := r.PublishedAt
		if len(date) >= 10 {
			date = date[:10] // Just the date part
		}
		fmt.Printf("Released: %s\n", date)
	}

	if r.Body != "" {
		fmt.Println()
		fmt.Println(strings.TrimSpace(r.Body))
	} else {
		fmt.Println("\n(No release notes)")
	}
	fmt.Println()
}

func findAssetForPlatform(assets []githubAsset) *githubAsset {
	// Build expected suffix based on current OS/arch
	suffix := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		suffix += ".exe"
	}

	for _, a := range assets {
		if strings.Contains(a.Name, suffix) && !strings.HasSuffix(a.Name, ".sha256") {
			return &a
		}
	}

	return nil
}

func findChecksumAsset(assets []githubAsset) *githubAsset {
	for _, a := range assets {
		if a.Name == "checksums.sha256" {
			return &a
		}
	}
	return nil
}

func promptConfirm(message string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [y/N]: ", message)

	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// isAllowedDownloadURL validates that the URL is from an allowed host.
func isAllowedDownloadURL(url string) bool {
	for _, host := range allowedDownloadHosts {
		if strings.HasPrefix(url, host) {
			return true
		}
	}
	return false
}

func downloadAndInstall(asset, checksumAsset *githubAsset, skipChecksum bool) error {
	// Validate download URL
	if !isAllowedDownloadURL(asset.BrowserDownloadURL) {
		return fmt.Errorf("invalid download URL: must be from github.com or objects.githubusercontent.com")
	}

	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot determine executable path: %w", err)
	}

	// Resolve symlinks to get the real path
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("cannot resolve executable path: %w", err)
	}

	// Download to a temporary file
	tmpFile, err := os.CreateTemp(filepath.Dir(execPath), "repo-ctr-update-*")
	if err != nil {
		return fmt.Errorf("cannot create temporary file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Clean up temp file on error
	defer func() {
		if tmpFile != nil {
			tmpFile.Close()
			os.Remove(tmpPath)
		}
	}()

	// Download the new binary
	resp, err := httpClient.Get(asset.BrowserDownloadURL)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Write to temp file and calculate checksum simultaneously
	hash := sha256.New()
	writer := io.MultiWriter(tmpFile, hash)
	_, err = io.Copy(writer, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Close the temp file before verification
	tmpFile.Close()
	tmpFile = nil

	// Calculate the checksum
	downloadedChecksum := hex.EncodeToString(hash.Sum(nil))

	// Verify checksum if available
	if !skipChecksum {
		if checksumAsset == nil {
			fmt.Println("Warning: No checksum file available for this release. Use --skip-checksum to proceed anyway.")
			return fmt.Errorf("checksum verification failed: no checksum file available")
		}

		if !isAllowedDownloadURL(checksumAsset.BrowserDownloadURL) {
			return fmt.Errorf("invalid checksum URL: must be from github.com or objects.githubusercontent.com")
		}

		expectedChecksum, err := fetchExpectedChecksum(checksumAsset.BrowserDownloadURL, asset.Name)
		if err != nil {
			return fmt.Errorf("checksum verification failed: %w", err)
		}

		if downloadedChecksum != expectedChecksum {
			return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, downloadedChecksum)
		}
		fmt.Println("Checksum verified.")
	} else {
		fmt.Println("Warning: Skipping checksum verification.")
	}

	// Make executable
	if err := os.Chmod(tmpPath, 0755); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Atomic replace: rename temp file to actual executable
	// On Windows, we need to rename the old file first
	if runtime.GOOS == "windows" {
		oldPath := execPath + ".old"
		os.Remove(oldPath) // Remove any previous .old file
		if err := os.Rename(execPath, oldPath); err != nil {
			return fmt.Errorf("failed to backup old binary: %w", err)
		}
		if err := os.Rename(tmpPath, execPath); err != nil {
			// Try to restore old binary
			if restoreErr := os.Rename(oldPath, execPath); restoreErr != nil {
				return fmt.Errorf("failed to install new binary: %w (rollback also failed: %v)", err, restoreErr)
			}
			return fmt.Errorf("failed to install new binary: %w", err)
		}
		// Clean up old file (may fail if still in use, that's OK)
		os.Remove(oldPath)
	} else {
		// On Unix, rename is atomic
		if err := os.Rename(tmpPath, execPath); err != nil {
			return fmt.Errorf("failed to install new binary: %w", err)
		}
	}

	return nil
}

// fetchExpectedChecksum downloads the checksum file and extracts the checksum for the given asset.
func fetchExpectedChecksum(checksumURL, assetName string) (string, error) {
	resp, err := httpClient.Get(checksumURL)
	if err != nil {
		return "", fmt.Errorf("failed to download checksum file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download checksum file: status %d", resp.StatusCode)
	}

	// Read the checksum file
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read checksum file: %w", err)
	}

	// Parse checksum file (format: "checksum  filename" per line)
	lines := strings.Split(string(body), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Split by whitespace (checksum files use two spaces or tab)
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			checksum := parts[0]
			filename := parts[len(parts)-1]
			if filename == assetName {
				return strings.ToLower(checksum), nil
			}
		}
	}

	return "", fmt.Errorf("checksum not found for %s", assetName)
}

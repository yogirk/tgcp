package core

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	// GitHubRepo is the repository for version checking
	GitHubRepo = "yogirk/tgcp"
	// ReleasesURL is the GitHub API endpoint for latest release
	ReleasesURL = "https://api.github.com/repos/" + GitHubRepo + "/releases/latest"
)

// VersionInfo holds the current application version information
type VersionInfo struct {
	Version string
	Commit  string
	Date    string
}

// UpdateInfo holds information about available updates
type UpdateInfo struct {
	Available      bool
	LatestVersion  string
	CurrentVersion string
	ReleaseURL     string
	ReleaseNotes   string
	CheckedAt      time.Time
	Error          error
}

// GitHubRelease represents the GitHub API response for a release
type GitHubRelease struct {
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	HTMLURL     string `json:"html_url"`
	Body        string `json:"body"`
	Prerelease  bool   `json:"prerelease"`
	Draft       bool   `json:"draft"`
	PublishedAt string `json:"published_at"`
}

// UpdateCheckedMsg is sent when version check completes
type UpdateCheckedMsg struct {
	UpdateInfo UpdateInfo
}

// CheckForUpdates fetches the latest release from GitHub and compares versions
func CheckForUpdates(currentVersion string) tea.Cmd {
	return func() tea.Msg {
		info := UpdateInfo{
			CurrentVersion: currentVersion,
			CheckedAt:      time.Now(),
		}

		// Skip check for dev builds
		if currentVersion == "dev" || currentVersion == "" {
			info.Available = false
			return UpdateCheckedMsg{UpdateInfo: info}
		}

		// Create HTTP client with timeout
		client := &http.Client{Timeout: 5 * time.Second}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, "GET", ReleasesURL, nil)
		if err != nil {
			info.Error = err
			return UpdateCheckedMsg{UpdateInfo: info}
		}

		req.Header.Set("Accept", "application/vnd.github.v3+json")
		req.Header.Set("User-Agent", "tgcp/"+currentVersion)

		resp, err := client.Do(req)
		if err != nil {
			info.Error = err
			return UpdateCheckedMsg{UpdateInfo: info}
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			info.Error = fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
			return UpdateCheckedMsg{UpdateInfo: info}
		}

		var release GitHubRelease
		if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
			info.Error = err
			return UpdateCheckedMsg{UpdateInfo: info}
		}

		// Skip draft and prerelease
		if release.Draft || release.Prerelease {
			info.Available = false
			return UpdateCheckedMsg{UpdateInfo: info}
		}

		latestVersion := strings.TrimPrefix(release.TagName, "v")
		info.LatestVersion = latestVersion
		info.ReleaseURL = release.HTMLURL
		info.ReleaseNotes = truncateReleaseNotes(release.Body, 200)

		// Compare versions
		info.Available = isNewerVersion(latestVersion, currentVersion)

		return UpdateCheckedMsg{UpdateInfo: info}
	}
}

// isNewerVersion compares two semantic versions (e.g., "1.2.3" vs "1.2.0")
// Returns true if latest is newer than current
func isNewerVersion(latest, current string) bool {
	// Clean up version strings
	latest = strings.TrimPrefix(latest, "v")
	current = strings.TrimPrefix(current, "v")

	// Handle dev/unknown versions
	if current == "dev" || current == "" || current == "unknown" {
		return false
	}

	latestParts := parseVersion(latest)
	currentParts := parseVersion(current)

	for i := 0; i < 3; i++ {
		if latestParts[i] > currentParts[i] {
			return true
		}
		if latestParts[i] < currentParts[i] {
			return false
		}
	}
	return false
}

// parseVersion parses a version string into [major, minor, patch]
func parseVersion(v string) [3]int {
	var parts [3]int
	// Remove any suffix after hyphen (e.g., "1.2.3-beta" -> "1.2.3")
	if idx := strings.Index(v, "-"); idx != -1 {
		v = v[:idx]
	}

	segments := strings.Split(v, ".")
	for i := 0; i < 3 && i < len(segments); i++ {
		fmt.Sscanf(segments[i], "%d", &parts[i])
	}
	return parts
}

// truncateReleaseNotes shortens release notes for display
func truncateReleaseNotes(notes string, maxLen int) string {
	// Remove markdown headers and clean up
	lines := strings.Split(notes, "\n")
	var cleaned []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		cleaned = append(cleaned, line)
	}
	result := strings.Join(cleaned, " ")

	if len(result) > maxLen {
		return result[:maxLen-3] + "..."
	}
	return result
}

// FormatVersion returns a formatted version string
func (v VersionInfo) FormatVersion() string {
	if v.Version == "dev" || v.Version == "" {
		return "dev"
	}
	return "v" + strings.TrimPrefix(v.Version, "v")
}

// String returns a full version string with commit info
func (v VersionInfo) String() string {
	ver := v.FormatVersion()
	if v.Commit != "" && v.Commit != "none" {
		ver += " (" + v.Commit + ")"
	}
	return ver
}

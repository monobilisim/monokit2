package lib

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

const (
	githubOwnerUpdater = "monobilisim"
	githubRepoUpdater  = "monokit2"
)

type UpdateResult struct {
	Name       string
	OldVersion string
	NewVersion string
	Updated    bool
	Error      error
}

// parseSemver extracts major, minor, patch from a version string
// Returns major version or -1 if not a valid semver
func parseSemver(version string) (major, minor, patch int, isDevel bool) {
	version = strings.TrimSpace(version)

	if strings.Contains(version, "devel") {
		return 0, 0, 0, true
	}

	// Remove 'v' prefix if present
	version = strings.TrimPrefix(version, "v")

	// Match semver pattern
	re := regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)`)
	matches := re.FindStringSubmatch(version)
	if len(matches) < 4 {
		return -1, -1, -1, false
	}

	fmt.Sscanf(matches[1], "%d", &major)
	fmt.Sscanf(matches[2], "%d", &minor)
	fmt.Sscanf(matches[3], "%d", &patch)

	return major, minor, patch, false
}

func isMajorVersionChange(oldVersion, newVersion string) bool {
	oldMajor, _, _, oldDevel := parseSemver(oldVersion)
	newMajor, _, _, newDevel := parseSemver(newVersion)

	// Devel versions can always be updated
	if oldDevel || newDevel {
		return false
	}

	// If we couldn't parse either version, don't consider it a major change
	if oldMajor == -1 || newMajor == -1 {
		return false
	}

	return newMajor > oldMajor
}

// Returns false for devel versions since they need checksum comparison instead
func isVersionMatch(currentVersion, targetVersion string) bool {
	current := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(currentVersion), "v"))
	target := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(targetVersion), "v"))

	// Devel versions need checksum comparison, not version comparison
	if strings.Contains(strings.ToLower(current), "devel") || strings.Contains(strings.ToLower(target), "devel") {
		return false
	}

	return current == target
}

// isChecksumMatch checks if a local file's hash matches the expected remote hash
// Returns true if they match (no update needed), false otherwise
func isChecksumMatch(filePath, expectedHash string) bool {
	if expectedHash == "" {
		return false // Can't verify, assume update needed
	}

	actualHash := calculateFileHash(filePath)
	if actualHash == "" {
		return false // Can't calculate, assume update needed
	}

	return strings.EqualFold(actualHash, expectedHash)
}

// verifyChecksum verifies a file's SHA256 hash against an expected hash
// Uses calculateFileHash from tui.go
func verifyChecksum(filePath, expectedHash string) error {
	if expectedHash == "" {
		return fmt.Errorf("no checksum available for verification")
	}

	actualHash := calculateFileHash(filePath)
	if actualHash == "" {
		return fmt.Errorf("failed to calculate file hash")
	}

	if !strings.EqualFold(actualHash, expectedHash) {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedHash, actualHash)
	}

	return nil
}

// ensureTmpDir creates the tmp directory if it doesn't exist
func ensureTmpDir(baseDir string) (string, error) {
	tmpDir := filepath.Join(baseDir, "tmp")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create tmp directory: %v", err)
	}
	return tmpDir, nil
}

// backupFile copies a file to the tmp directory
func backupFile(srcPath, tmpDir string) (string, error) {
	fileName := filepath.Base(srcPath)
	backupPath := filepath.Join(tmpDir, fileName+".backup")

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return "", fmt.Errorf("failed to open source file: %v", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(backupPath)
	if err != nil {
		return "", fmt.Errorf("failed to create backup file: %v", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return "", fmt.Errorf("failed to copy file: %v", err)
	}

	// Preserve permissions
	srcInfo, err := os.Stat(srcPath)
	if err == nil {
		os.Chmod(backupPath, srcInfo.Mode())
	}

	return backupPath, nil
}

// restoreFile restores a file from backup
func restoreFile(backupPath, originalPath string) error {
	srcFile, err := os.Open(backupPath)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %v", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(originalPath)
	if err != nil {
		return fmt.Errorf("failed to create original file: %v", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to restore file: %v", err)
	}

	// Preserve permissions
	srcInfo, err := os.Stat(backupPath)
	if err == nil {
		os.Chmod(originalPath, srcInfo.Mode())
	}

	return nil
}

// cleanupBackup removes a backup file
func cleanupBackup(backupPath string) error {
	return os.Remove(backupPath)
}

// downloadFile downloads a file from URL and saves it to the specified path
func downloadFile(url, destPath string) error {
	client := &http.Client{
		Timeout: 5 * time.Minute,
	}

	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	outFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, resp.Body); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	// Make executable
	if err := os.Chmod(destPath, 0755); err != nil {
		return fmt.Errorf("failed to set permissions: %v", err)
	}

	return nil
}

// getVersionFromBinary runs the binary with "version" argument to get its version
func getVersionFromBinary(binaryPath string) string {
	cmd := exec.Command(binaryPath, "-v")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

func verifyBinary(binaryPath string) error {
	cmd := exec.Command(binaryPath, "-v")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("binary verification failed: %v", err)
	}
	return nil
}

func getLatestReleaseTag(useDevel bool) (string, error) {
	if useDevel {
		return "devel", nil
	}

	url := fmt.Sprintf("https://github.com/%s/%s/releases/latest", githubOwnerUpdater, githubRepoUpdater)

	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch latest release: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusMovedPermanently {
		location := resp.Header.Get("Location")
		parts := strings.Split(location, "/tag/")
		if len(parts) == 2 {
			return parts[1], nil
		}
	}

	return "", fmt.Errorf("could not determine latest release tag")
}

// UpdateMonokit2 updates the monokit2 binary to the latest version
// It backs up the current binary, downloads the new one, and restores on failure
func UpdateMonokit2(currentVersion string, forceUpdate bool) (*UpdateResult, error) {
	if !GlobalConfig.AutoUpdate.Enabled {
		Logger.Info().Msg("Auto-update is disabled in configuration, skipping monokit2 update")
		return &UpdateResult{}, nil
	}

	result := &UpdateResult{
		Name:       "monokit2",
		OldVersion: currentVersion,
		Updated:    false,
	}

	useDevel := strings.Contains(strings.ToLower(currentVersion), "devel")

	tag, err := getLatestReleaseTag(useDevel)
	if err != nil {
		result.Error = err
		return result, err
	}
	result.NewVersion = tag

	// For stable releases: check if version matches (skip update if same)
	if !useDevel && !forceUpdate && isVersionMatch(currentVersion, tag) {
		Logger.Info().Msgf("monokit2 is already at version %s, skipping update", tag)
		return result, nil
	}

	// Check for major version change (skip for devel)
	if !useDevel && !forceUpdate && isMajorVersionChange(currentVersion, tag) {
		result.Error = fmt.Errorf("major version change detected (%s -> %s), use force update", currentVersion, tag)
		return result, result.Error
	}

	// Get the path to the current binary
	execPath, err := os.Executable()
	if err != nil {
		result.Error = fmt.Errorf("failed to get executable path: %v", err)
		return result, result.Error
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		result.Error = fmt.Errorf("failed to resolve symlinks: %v", err)
		return result, result.Error
	}

	// Build filename for checksum lookup
	currentOS := runtime.GOOS
	currentArch := runtime.GOARCH
	fileName := fmt.Sprintf("monokit2_%s_%s_%s", tag, currentOS, currentArch)
	if currentOS == "windows" {
		fileName += ".exe"
	}

	checksums, err := fetchChecksums(tag)
	if err != nil {
		result.Error = fmt.Errorf("failed to fetch checksums: %v", err)
		return result, result.Error
	}
	expectedHash := checksums[fileName]

	// For devel releases: check if checksum matches (skip update if same binary)
	if useDevel && !forceUpdate && isChecksumMatch(execPath, expectedHash) {
		Logger.Info().Msgf("monokit2 is already up-to-date (checksum match), skipping update")
		return result, nil
	}

	tmpDir, err := ensureTmpDir(DbDir)
	if err != nil {
		result.Error = err
		return result, err
	}

	backupPath, err := backupFile(execPath, tmpDir)
	if err != nil {
		result.Error = fmt.Errorf("failed to backup current binary: %v", err)
		return result, result.Error
	}

	downloadURL := fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s",
		githubOwnerUpdater, githubRepoUpdater, tag, fileName)

	tmpDownloadPath := filepath.Join(tmpDir, "monokit2.new")
	if err := downloadFile(downloadURL, tmpDownloadPath); err != nil {
		cleanupBackup(backupPath)
		result.Error = fmt.Errorf("failed to download new version: %v", err)
		return result, result.Error
	}

	if err := verifyChecksum(tmpDownloadPath, expectedHash); err != nil {
		os.Remove(tmpDownloadPath)
		cleanupBackup(backupPath)
		result.Error = fmt.Errorf("checksum verification failed: %v", err)
		return result, result.Error
	}

	if err := verifyBinary(tmpDownloadPath); err != nil {
		os.Remove(tmpDownloadPath)
		cleanupBackup(backupPath)
		result.Error = fmt.Errorf("new binary verification failed: %v", err)
		return result, result.Error
	}

	// Replace the old binary with the new one
	if err := os.Rename(tmpDownloadPath, execPath); err != nil {
		// Try copy instead
		srcFile, err2 := os.Open(tmpDownloadPath)
		if err2 != nil {
			restoreFile(backupPath, execPath)
			cleanupBackup(backupPath)
			result.Error = fmt.Errorf("failed to replace binary: %v", err)
			return result, result.Error
		}
		defer srcFile.Close()

		dstFile, err2 := os.Create(execPath)
		if err2 != nil {
			restoreFile(backupPath, execPath)
			cleanupBackup(backupPath)
			result.Error = fmt.Errorf("failed to replace binary: %v", err2)
			return result, result.Error
		}
		defer dstFile.Close()

		if _, err2 := io.Copy(dstFile, srcFile); err2 != nil {
			restoreFile(backupPath, execPath)
			cleanupBackup(backupPath)
			result.Error = fmt.Errorf("failed to copy new binary: %v", err2)
			return result, result.Error
		}
		os.Chmod(execPath, 0755)
		os.Remove(tmpDownloadPath)
	}

	// Verify the replaced binary works
	if err := verifyBinary(execPath); err != nil {
		// Restore from backup
		if restoreErr := restoreFile(backupPath, execPath); restoreErr != nil {
			result.Error = fmt.Errorf("update failed and restore failed: %v (restore error: %v)", err, restoreErr)
			return result, result.Error
		}
		cleanupBackup(backupPath)
		result.Error = fmt.Errorf("new binary failed verification, restored old version: %v", err)
		return result, result.Error
	}

	cleanupBackup(backupPath)
	result.Updated = true
	return result, nil
}

// UpdatePlugins updates all plugins to the latest version
// It backs up current plugins, downloads new ones, and restores on failure
func UpdatePlugins(currentMonokitVersion string, forceUpdate bool) ([]UpdateResult, error) {
	if !GlobalConfig.AutoUpdate.Enabled {
		Logger.Info().Msg("Auto-update is disabled in configuration, skipping plugins update")
		return []UpdateResult{}, nil
	}

	var results []UpdateResult

	useDevel := strings.Contains(strings.ToLower(currentMonokitVersion), "devel")

	tag, err := getLatestReleaseTag(useDevel)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest release: %v", err)
	}

	tmpDir, err := ensureTmpDir(PluginsDir)
	if err != nil {
		return nil, err
	}

	checksums, err := fetchChecksums(tag)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch checksums: %v", err)
	}

	entries, err := os.ReadDir(PluginsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugins directory: %v", err)
	}

	currentOS := runtime.GOOS
	currentArch := runtime.GOARCH

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		pluginName := entry.Name()

		// Skip tmp directory and non-executable files
		info, err := entry.Info()
		if err != nil || info.Mode()&0111 == 0 {
			continue
		}

		// Skip monokit2 itself if present
		if pluginName == "monokit2" {
			continue
		}

		pluginPath := filepath.Join(PluginsDir, pluginName)
		result := UpdateResult{
			Name:       pluginName,
			OldVersion: getVersionFromBinary(pluginPath),
			NewVersion: tag,
			Updated:    false,
		}

		// For stable releases: check if version matches (skip update if same)
		if !useDevel && !forceUpdate && isVersionMatch(result.OldVersion, tag) {
			Logger.Info().Msgf("Plugin %s is already at version %s, skipping update", pluginName, tag)
			results = append(results, result)
			continue
		}

		// Build filename for checksum lookup
		fileName := fmt.Sprintf("%s_%s_%s_%s", pluginName, tag, currentOS, currentArch)
		if currentOS == "windows" {
			fileName += ".exe"
		}
		expectedHash := checksums[fileName]

		// For devel releases: check if checksum matches (skip update if same binary)
		if useDevel && !forceUpdate && isChecksumMatch(pluginPath, expectedHash) {
			Logger.Info().Msgf("Plugin %s is already up-to-date (checksum match), skipping update", pluginName)
			results = append(results, result)
			continue
		}

		// Check for major version change
		if !useDevel && !forceUpdate && isMajorVersionChange(result.OldVersion, tag) {
			result.Error = fmt.Errorf("major version change detected, skipping")
			results = append(results, result)
			continue
		}

		backupPath, err := backupFile(pluginPath, tmpDir)
		if err != nil {
			result.Error = fmt.Errorf("failed to backup: %v", err)
			results = append(results, result)
			continue
		}

		downloadURL := fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s",
			githubOwnerUpdater, githubRepoUpdater, tag, fileName)

		tmpDownloadPath := filepath.Join(tmpDir, pluginName+".new")
		if err := downloadFile(downloadURL, tmpDownloadPath); err != nil {
			cleanupBackup(backupPath)
			result.Error = fmt.Errorf("failed to download: %v", err)
			results = append(results, result)
			continue
		}

		if err := verifyChecksum(tmpDownloadPath, expectedHash); err != nil {
			os.Remove(tmpDownloadPath)
			cleanupBackup(backupPath)
			result.Error = fmt.Errorf("checksum verification failed: %v", err)
			results = append(results, result)
			continue
		}

		if err := verifyBinary(tmpDownloadPath); err != nil {
			os.Remove(tmpDownloadPath)
			cleanupBackup(backupPath)
			result.Error = fmt.Errorf("verification failed: %v", err)
			results = append(results, result)
			continue
		}

		// Replace the old plugin with the new one
		if err := os.Rename(tmpDownloadPath, pluginPath); err != nil {
			// Try copy instead
			if err := copyFile(tmpDownloadPath, pluginPath); err != nil {
				restoreFile(backupPath, pluginPath)
				cleanupBackup(backupPath)
				result.Error = fmt.Errorf("failed to replace: %v", err)
				results = append(results, result)
				continue
			}
			os.Remove(tmpDownloadPath)
		}

		// Verify the replaced plugin works
		if err := verifyBinary(pluginPath); err != nil {
			restoreFile(backupPath, pluginPath)
			cleanupBackup(backupPath)
			result.Error = fmt.Errorf("new plugin failed, restored old: %v", err)
			results = append(results, result)
			continue
		}

		cleanupBackup(backupPath)
		result.Updated = true
		results = append(results, result)
	}

	return results, nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	return os.Chmod(dst, 0755)
}

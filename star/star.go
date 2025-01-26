package star

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Package represents a GitHub package to be managed.
type Package struct {
	User    string `json:"user"`
	Repo    string `json:"repo"`
	Version string `json:"version"`
	File    string `json:"file"`
}

const starFile = "./stars/.stars"
const installDir = "./stars"

// Install installs the latest release of the specified package.
func Install(pkg Package) error {
	platform := getPlatform() // Get both OS and architecture.
	if platform == "unknown-unknown" {
		return errors.New("unsupported platform")
	}

	releaseURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", pkg.User, pkg.Repo)
	resp, err := http.Get(releaseURL)
	if err != nil {
		return fmt.Errorf("failed to fetch release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch release: status code %d", resp.StatusCode)
	}

	var releaseData struct {
		Assets []struct {
			Name string `json:"name"`
			URL  string `json:"browser_download_url"`
		} `json:"assets"`
		TagName string `json:"tag_name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&releaseData); err != nil {
		return fmt.Errorf("failed to decode release data: %w", err)
	}

	// Match the correct asset based on platform (OS + architecture)
	var downloadURL string
	for _, asset := range releaseData.Assets {
		if strings.Contains(strings.ToLower(asset.Name), platform) {
			downloadURL = asset.URL
			pkg.File = asset.Name
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("no compatible release found for platform: %s", platform)
	}

	pkg.Version = releaseData.TagName
	destPath := filepath.Join(installDir, pkg.File)
	if err := os.MkdirAll(installDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create install directory: %w", err)
	}

	if err := downloadExecutable(downloadURL, destPath); err != nil {
		return fmt.Errorf("failed to download executable: %w", err)
	}

	if err := updateStarsFile(pkg); err != nil {
		return fmt.Errorf("failed to update stars file: %w", err)
	}

	return nil
}

// getPlatform determines the platform based on runtime environment.
func getPlatform() string {
	os := runtime.GOOS
	arch := runtime.GOARCH

	// Combine OS and architecture (e.g., "linux-amd64", "windows-arm64")
	return fmt.Sprintf("%s-%s", os, arch)
}

// downloadExecutable downloads the executable file to the specified path.
func downloadExecutable(url, destPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download file: status code %d", resp.StatusCode)
	}

	file, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// updateStarsFile updates the .stars file with the installed package.
func updateStarsFile(pkg Package) error {
	packages, err := listInstalledPackages()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to list installed packages: %w", err)
	}

	packages = append(packages, pkg)
	return saveInstalledPackages(packages)
}

// listInstalledPackages lists all packages from the .stars file.
func listInstalledPackages() ([]Package, error) {
	file, err := os.Open(starFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var packages []Package
	if err := json.NewDecoder(file).Decode(&packages); err != nil {
		return nil, fmt.Errorf("failed to decode .stars file: %w", err)
	}

	return packages, nil
}

// saveInstalledPackages saves the package list to the .stars file.
func saveInstalledPackages(packages []Package) error {
	file, err := os.Create(starFile)
	if err != nil {
		return fmt.Errorf("failed to create .stars file: %w", err)
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(packages); err != nil {
		return fmt.Errorf("failed to write to .stars file: %w", err)
	}

	return nil
}

// ListInstalledStars lists all installed packages.
func ListInstalledStars() ([]Package, error) {
	return listInstalledPackages()
}

// Uninstall removes a package and updates the .stars file.
// Uninstall removes a package and updates the .stars file.
func Uninstall(pkg Package) error {
	// Find the installed package by its file name
	packages, err := listInstalledPackages()
	if err != nil {
		return fmt.Errorf("failed to list installed packages: %w", err)
	}

	// Locate the file for the package
	var filePath string
	for _, p := range packages {
		if p.User == pkg.User && p.Repo == pkg.Repo {
			filePath = filepath.Join(installDir, p.File)
			break
		}
	}

	if filePath == "" {
		return fmt.Errorf("package %s/%s not found", pkg.User, pkg.Repo)
	}

	// Remove the executable file
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to remove executable: %w", err)
	}

	// Update the package list in .stars file
	var updatedPackages []Package
	for _, p := range packages {
		if p.User != pkg.User || p.Repo != pkg.Repo {
			updatedPackages = append(updatedPackages, p)
		}
	}

	if err := saveInstalledPackages(updatedPackages); err != nil {
		return fmt.Errorf("failed to update stars file: %w", err)
	}

	return nil
}

// Update updates an installed package.
func Update(pkg Package) error {
	if err := Uninstall(pkg); err != nil {
		return fmt.Errorf("failed to uninstall package: %w", err)
	}
	return Install(pkg)
}

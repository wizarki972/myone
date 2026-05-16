package release

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/wizarki972/myone/internal/utils/fldir"
)

const BASE_URL = "https://api.github.com/wizarki972/"

type Release struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	PublishedAt time.Time `json:"published_at"`
	HTMLURL     string    `json:"html_url"`
	Assets      []struct {
		Url  string `json:"browser_download_url"`
		Name string `json:"name"`
	} `json:"assets"`
}

func getRepoURL(repoName string) (string, error) {
	if len(strings.TrimSpace(repoName)) == 0 {
		return "", errors.New("enter a valid github repository name")
	}
	return fmt.Sprintf("https://api.github.com/repos/wizarki972/%s/releases/latest", repoName), nil
}

func GetLatestRelease(repoName string) (*Release, error) {
	// get github release api URL
	url, err := getRepoURL(repoName)
	if err != nil {
		return nil, err
	}

	// http request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("user-agent", repoName+"-release-checker")

	// http client
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// http response
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("github API returned status: " + resp.Status)
	}

	// release
	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}
	release.Name = repoName

	return &release, nil
}

func DownloadLatestRelease(release *Release) (string, error) {
	destinationDir := filepath.Join(fldir.GetHomeDir(), ".cache/myone/releases")
	if err := fldir.CreateDirectory(destinationDir); err != nil {
		return "", err
	}

	destination := filepath.Join(destinationDir, release.Name+"-"+release.TagName+".zip")

	for _, asset := range release.Assets {
		if asset.Name == "flat-archive.zip" {
			fmt.Println(asset.Url)
			if err := fldir.DownloadURL(asset.Url, destination, true); err != nil {
				return "", err
			}
			break
		}
	}

	return destination, nil
}

func IsNewer(release *Release, cmajor, cminor, cpatch int) (bool, error) {
	major, minor, patch, err := VersionParser(release.TagName)
	if err != nil {
		return false, err
	}

	if major > cmajor ||
		(major == cmajor && minor > cminor) ||
		(major == cmajor && minor == cminor && patch > cpatch) {
		return true, nil
	}
	return false, nil
}

func VersionParser(version string) (int, int, int, error) {
	var patchVer int
	var err error

	parts := strings.Split(version, "-")
	if len(parts) < 2 {
		patchVer = 0
	} else {
		patchVer, err = strconv.Atoi(parts[1])
		if err != nil {
			return -1, -1, -1, err
		}
	}

	parts = strings.Split(parts[0], ".")
	if len(parts) < 3 {
		return -1, -1, -1, errors.New("Improper version string. failed to parse - " + version)

	}
	majorVer, err := strconv.Atoi(parts[1])
	if err != nil {
		return -1, -1, -1, err
	}
	minorVer, err := strconv.Atoi(parts[2])
	if err != nil {
		return -1, -1, -1, err
	}
	return majorVer, minorVer, patchVer, nil
}

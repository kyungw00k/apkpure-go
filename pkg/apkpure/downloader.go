package apkpure

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ListVersions retrieves available versions for the given apps
func (c *Client) ListVersions(apps []AppInfo) error {
	for _, app := range apps {
		if c.options.OutputFormat == "plaintext" {
			fmt.Printf("Versions available for %s on APKPure:\n", app.PackageID)
		}

		versions, err := c.fetchVersions(app.PackageID)
		if err != nil {
			if c.options.OutputFormat == "plaintext" {
				fmt.Printf("| Error: %v\n", err)
			}
			continue
		}

		if c.options.OutputFormat == "plaintext" {
			versionNames := make([]string, 0, len(versions))
			for _, v := range versions {
				versionNames = append(versionNames, v.VersionName)
			}
			if len(versionNames) > 0 {
				fmt.Printf("| ")
				for i, v := range versionNames {
					if i > 0 {
						fmt.Printf(", ")
					}
					fmt.Printf("%s", v)
				}
				fmt.Printf("\n")
			}
		}
	}

	return nil
}

// fetchVersions fetches version information from APKPure API
func (c *Client) fetchVersions(packageID string) ([]VersionInfo, error) {
	url := c.getVersionsURL(packageID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header = c.buildHeaders()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return c.parseVersionResponse(body)
}

// parseVersionResponse parses the JSON response from APKPure API
func (c *Client) parseVersionResponse(body []byte) ([]VersionInfo, error) {
	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	versions := make([]VersionInfo, 0, len(apiResp.VersionList))
	for _, v := range apiResp.VersionList {
		if v.Asset.URL != "" {
			versions = append(versions, VersionInfo{
				VersionName: v.VersionName,
				VersionCode: v.VersionCode,
				APKType:     v.Asset.Type,
				DownloadURL: v.Asset.URL,
			})
		}
	}

	return versions, nil
}

// Download downloads a single APK
func (c *Client) Download(app AppInfo, outPath string) error {
	fmt.Printf("Downloading %s...\n", app.PackageID)

	versions, err := c.fetchVersions(app.PackageID)
	if err != nil {
		return fmt.Errorf("failed to fetch versions: %w", err)
	}

	if len(versions) == 0 {
		return fmt.Errorf("no versions available for %s", app.PackageID)
	}

	// Find the matching version or use the latest
	var targetVersion *VersionInfo
	if app.Version != "" {
		for i := range versions {
			if versions[i].VersionName == app.Version {
				targetVersion = &versions[i]
				break
			}
		}
		if targetVersion == nil {
			return fmt.Errorf("version %s not found for %s", app.Version, app.PackageID)
		}
	} else {
		// Use the first (latest) version
		targetVersion = &versions[0]
	}

	// Build filename
	appString := app.PackageID
	if app.Version != "" {
		appString = fmt.Sprintf("%s@%s", app.PackageID, app.Version)
	}

	ext := ".apk"
	if targetVersion.APKType == "XAPK" {
		ext = ".xapk"
	}
	filename := appString + ext

	// Download with retry
	err = c.downloadWithRetry(targetVersion.DownloadURL, outPath, filename)
	if err != nil {
		return err
	}

	fmt.Printf("%s downloaded successfully!\n", appString)
	return nil
}

// downloadWithRetry downloads a file with retry logic (up to 3 attempts)
func (c *Client) downloadWithRetry(url, outPath, filename string) error {
	maxRetries := 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if attempt > 1 {
			fmt.Printf("Retry #%d...\n", attempt-1)
		}

		err := c.downloadFile(url, outPath, filename)
		if err == nil {
			return nil
		}

		lastErr = err
		if attempt < maxRetries {
			time.Sleep(time.Second)
		}
	}

	return fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

// downloadFile downloads a file from the given URL
func (c *Client) downloadFile(url, outPath, filename string) (err error) {
	fullPath := filepath.Join(outPath, filename)

	// Check if file already exists
	if _, err := os.Stat(fullPath); err == nil {
		return fmt.Errorf("file already exists: %s", filename)
	}

	// Create output file
	outFile, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		if closeErr := outFile.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close output file: %w", closeErr)
		}
	}()

	// Download
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid status code: %d", resp.StatusCode)
	}

	// Copy with progress
	total := resp.ContentLength
	downloaded := int64(0)

	buffer := make([]byte, 32*1024) // 32KB buffer
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			_, writeErr := outFile.Write(buffer[:n])
			if writeErr != nil {
				return writeErr
			}
			downloaded += int64(n)

			// Call progress callback if set
			if c.options.ProgressCallback != nil {
				c.options.ProgressCallback(filename, downloaded, total)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// DownloadMultiple downloads multiple APKs in parallel
func (c *Client) DownloadMultiple(apps []AppInfo, outPath string) []DownloadResult {
	results := make([]DownloadResult, len(apps))
	var wg sync.WaitGroup

	// Create a semaphore to limit concurrent downloads
	sem := make(chan struct{}, c.options.Parallel)

	for i, app := range apps {
		wg.Add(1)
		go func(idx int, appInfo AppInfo) {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			// Sleep if configured
			if c.options.SleepDuration > 0 {
				time.Sleep(c.options.SleepDuration)
			}

			// Download
			err := c.Download(appInfo, outPath)

			appString := appInfo.PackageID
			if appInfo.Version != "" {
				appString = fmt.Sprintf("%s@%s", appInfo.PackageID, appInfo.Version)
			}

			results[idx] = DownloadResult{
				AppInfo:  appInfo,
				Filename: appString,
				Success:  err == nil,
				Error:    err,
			}
		}(i, app)
	}

	wg.Wait()
	return results
}

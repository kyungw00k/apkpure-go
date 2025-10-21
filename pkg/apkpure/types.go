package apkpure

import "time"

// DownloadOptions represents options for downloading APKs
type DownloadOptions struct {
	// Architecture (e.g., "arm64-v8a", "armeabi-v7a", "x86", "x86_64")
	Arch string
	// Language (e.g., "en-US", "ko-KR")
	Language string
	// OS Version (e.g., "35" for Android 15)
	OSVersion string
	// Parallel download count
	Parallel int
	// Sleep duration between downloads
	SleepDuration time.Duration
	// Output format (plaintext or json)
	OutputFormat string
	// Progress callback
	ProgressCallback func(filename string, downloaded, total int64)
}

// AppInfo represents an app to download
type AppInfo struct {
	// Package name (e.g., "com.instagram.android")
	PackageID string
	// Version (optional, e.g., "1.2.3")
	Version string
}

// VersionInfo represents a version of an app
type VersionInfo struct {
	VersionName string
	VersionCode string
	APKType     string // "APK" or "XAPK"
	DownloadURL string
}

// DownloadResult represents the result of a download operation
type DownloadResult struct {
	AppInfo  AppInfo
	Filename string
	Success  bool
	Error    error
}

// APIResponse represents the API response from APKPure
type APIResponse struct {
	VersionList []struct {
		VersionName string `json:"version_name"`
		VersionCode string `json:"version_code"`
		Asset       struct {
			URL  string `json:"url"`
			Type string `json:"type"`
		} `json:"asset"`
	} `json:"version_list"`
}

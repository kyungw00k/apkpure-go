package apkpure

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	versionsURLFormat = "https://tapi.pureapk.com/v3/get_app_his_version?hl=en&package_name=%s"
	defaultUserAgent  = "Dalvik/2.1.0 (Linux; U; Android 15; Pixel 4a (5G) Build/BP1A.250505.005); APKPure/3.20.53 (Aegon)"
)

// Client represents an APKPure client
type Client struct {
	httpClient *http.Client
	options    DownloadOptions
}

// NewClient creates a new APKPure client with the given options
func NewClient(opts DownloadOptions) *Client {
	// Set defaults
	if opts.Arch == "" {
		opts.Arch = "arm64-v8a;armeabi-v7a;armeabi;x86;x86_64"
	}
	if opts.Language == "" {
		opts.Language = "en-US"
	}
	if opts.OSVersion == "" {
		opts.OSVersion = "35"
	}
	if opts.Parallel <= 0 {
		opts.Parallel = 4
	}
	if opts.OutputFormat == "" {
		opts.OutputFormat = "plaintext"
	}

	return &Client{
		httpClient: &http.Client{},
		options:    opts,
	}
}

// buildHeaders creates HTTP headers for APKPure API requests
func (c *Client) buildHeaders() http.Header {
	headers := http.Header{}
	headers.Set("User-Agent", defaultUserAgent)
	headers.Set("ual-access-businessid", "projecta")

	// Build device info JSON
	deviceInfo := c.buildDeviceInfo()
	headers.Set("ual-access-projecta", deviceInfo)

	return headers
}

// buildDeviceInfo creates the device info JSON string
func (c *Client) buildDeviceInfo() string {
	// Parse architecture list
	var abis []string
	if c.options.Arch != "" {
		// Split by semicolon
		archList := c.options.Arch
		current := ""
		for i := 0; i < len(archList); i++ {
			if archList[i] == ';' {
				if current != "" {
					abis = append(abis, current)
					current = ""
				}
			} else {
				current += string(archList[i])
			}
		}
		if current != "" {
			abis = append(abis, current)
		}
	}

	if len(abis) == 0 {
		abis = []string{"arm64-v8a", "armeabi-v7a", "armeabi", "x86", "x86_64"}
	}

	deviceInfoMap := map[string]interface{}{
		"device_info": map[string]interface{}{
			"abis":     abis,
			"language": c.options.Language,
			"os_ver":   c.options.OSVersion,
		},
	}

	jsonBytes, _ := json.Marshal(deviceInfoMap)
	return string(jsonBytes)
}

// getVersionsURL returns the URL for fetching app versions
func (c *Client) getVersionsURL(packageID string) string {
	return fmt.Sprintf(versionsURLFormat, packageID)
}

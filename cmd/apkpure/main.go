package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kyungw00k/apkpure-go/pkg/apkpure"
)

var (
	appID         string
	csvFile       string
	fieldNum      int
	versionField  int
	listVersions  bool
	options       string
	parallel      int
	sleepDuration int64
	outPath       string
)

func init() {
	flag.StringVar(&appID, "a", "", "App ID (e.g., com.instagram.android or com.instagram.android@1.2.3)")
	flag.StringVar(&appID, "app", "", "App ID (alias for -a)")
	flag.StringVar(&csvFile, "c", "", "CSV file containing app IDs")
	flag.StringVar(&csvFile, "csv", "", "CSV file containing app IDs (alias for -c)")
	flag.IntVar(&fieldNum, "f", 1, "CSV field number containing app IDs")
	flag.IntVar(&fieldNum, "field", 1, "CSV field number (alias for -f)")
	flag.IntVar(&versionField, "v", 0, "CSV field number containing versions")
	flag.IntVar(&versionField, "version-field", 0, "CSV field number for versions (alias for -v)")
	flag.BoolVar(&listVersions, "l", false, "List available versions")
	flag.BoolVar(&listVersions, "list-versions", false, "List available versions (alias for -l)")
	flag.StringVar(&options, "o", "", "Additional options (e.g., arch=arm64-v8a,language=en-US)")
	flag.StringVar(&options, "options", "", "Additional options (alias for -o)")
	flag.IntVar(&parallel, "r", 4, "Number of parallel downloads")
	flag.IntVar(&parallel, "parallel", 4, "Number of parallel downloads (alias for -r)")
	flag.Int64Var(&sleepDuration, "s", 0, "Sleep duration between downloads in milliseconds")
	flag.Int64Var(&sleepDuration, "sleep-duration", 0, "Sleep duration (alias for -s)")
}

func main() {
	flag.Parse()

	// Get output path from remaining args
	args := flag.Args()
	if !listVersions && len(args) == 0 {
		fmt.Println("Error: OUTPATH is required when downloading files")
		flag.Usage()
		os.Exit(1)
	}

	if len(args) > 0 {
		outPath = args[0]
	}

	// Parse app list
	var apps []apkpure.AppInfo
	var err error

	if appID != "" {
		apps, err = parseAppID(appID)
	} else if csvFile != "" {
		apps, err = parseCSVFile(csvFile, fieldNum, versionField)
	} else {
		fmt.Println("Error: Either -a/--app or -c/--csv must be specified")
		flag.Usage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("Error parsing apps: %v\n", err)
		os.Exit(1)
	}

	if len(apps) == 0 {
		fmt.Println("Error: No apps to process")
		os.Exit(1)
	}

	// Parse options
	opts := parseOptions(options)
	opts.Parallel = parallel
	opts.SleepDuration = time.Duration(sleepDuration) * time.Millisecond

	// Add progress callback if not listing versions
	if !listVersions {
		opts.ProgressCallback = apkpure.SimpleProgressCallback()
	}

	// Create client
	client := apkpure.NewClient(opts)

	// Execute
	if listVersions {
		err = client.ListVersions(apps)
		if err != nil {
			fmt.Printf("Error listing versions: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Validate output path
		if err := validateOutPath(outPath); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if len(apps) == 1 {
			// Single download
			err = client.Download(apps[0], outPath)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		} else {
			// Multiple downloads
			results := client.DownloadMultiple(apps, outPath)

			// Print summary
			successCount := 0
			for _, result := range results {
				if result.Success {
					successCount++
				} else {
					fmt.Printf("Failed to download %s: %v\n", result.AppInfo.PackageID, result.Error)
				}
			}

			fmt.Printf("\nDownload complete: %d/%d succeeded\n", successCount, len(results))
		}
	}
}

// parseAppID parses a single app ID (with optional version)
func parseAppID(appID string) ([]apkpure.AppInfo, error) {
	parts := strings.SplitN(appID, "@", 2)
	app := apkpure.AppInfo{
		PackageID: parts[0],
	}
	if len(parts) > 1 {
		app.Version = parts[1]
	}
	return []apkpure.AppInfo{app}, nil
}

// parseCSVFile parses a CSV file for app IDs
func parseCSVFile(filename string, fieldNum, versionField int) ([]apkpure.AppInfo, error) {
	if fieldNum < 1 {
		return nil, fmt.Errorf("field number must be 1 or greater")
	}
	if versionField != 0 && versionField < 1 {
		return nil, fmt.Errorf("version field number must be 1 or greater")
	}
	if versionField != 0 && fieldNum == versionField {
		return nil, fmt.Errorf("app ID and version fields must be different")
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var apps []apkpure.AppInfo
	fieldIdx := fieldNum - 1
	versionIdx := versionField - 1

	for _, record := range records {
		if len(record) <= fieldIdx {
			continue
		}

		appID := strings.TrimSpace(record[fieldIdx])
		if appID == "" {
			continue
		}

		app := apkpure.AppInfo{
			PackageID: appID,
		}

		if versionField > 0 && len(record) > versionIdx {
			version := strings.TrimSpace(record[versionIdx])
			if version != "" {
				app.Version = version
			}
		}

		apps = append(apps, app)
	}

	return apps, nil
}

// parseOptions parses the options string
func parseOptions(optStr string) apkpure.DownloadOptions {
	opts := apkpure.DownloadOptions{}

	if optStr == "" {
		return opts
	}

	pairs := strings.Split(optStr, ",")
	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			continue
		}

		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		switch key {
		case "arch":
			opts.Arch = value
		case "language":
			opts.Language = value
		case "os_ver":
			opts.OSVersion = value
		case "output_format":
			opts.OutputFormat = value
		}
	}

	return opts
}

// validateOutPath validates the output path
func validateOutPath(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("path does not exist: %s", path)
	}

	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}

	return nil
}

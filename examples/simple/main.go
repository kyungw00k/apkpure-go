package main

import (
	"fmt"
	"log"

	"github.com/kyungw00k/apkpure-go/pkg/apkpure"
)

func main() {
	// Example 1: List versions for an app
	fmt.Println("Example 1: Listing versions")
	fmt.Println("================================")

	client := apkpure.NewClient(apkpure.DownloadOptions{})

	apps := []apkpure.AppInfo{
		{PackageID: "com.instagram.android"},
	}

	err := client.ListVersions(apps)
	if err != nil {
		log.Fatalf("Error listing versions: %v", err)
	}

	fmt.Println()

	// Example 2: Download latest version
	fmt.Println("Example 2: Downloading latest version")
	fmt.Println("================================")

	opts := apkpure.DownloadOptions{
		Arch:             "arm64-v8a",
		ProgressCallback: apkpure.SimpleProgressCallback(),
	}

	client2 := apkpure.NewClient(opts)

	app := apkpure.AppInfo{
		PackageID: "com.instagram.android",
	}

	err = client2.Download(app, ".")
	if err != nil {
		log.Fatalf("Error downloading: %v", err)
	}

	fmt.Println()

	// Example 3: Download specific version
	fmt.Println("Example 3: Downloading specific version")
	fmt.Println("================================")

	appWithVersion := apkpure.AppInfo{
		PackageID: "com.instagram.android",
		Version:   "150.0.0.0",
	}

	err = client2.Download(appWithVersion, ".")
	if err != nil {
		log.Fatalf("Error downloading: %v", err)
	}

	fmt.Println()

	// Example 4: Download multiple apps in parallel
	fmt.Println("Example 4: Downloading multiple apps")
	fmt.Println("================================")

	multipleApps := []apkpure.AppInfo{
		{PackageID: "com.instagram.android"},
		{PackageID: "com.facebook.katana"},
		{PackageID: "com.twitter.android"},
	}

	results := client2.DownloadMultiple(multipleApps, ".")

	fmt.Println("\nDownload Results:")
	for _, result := range results {
		if result.Success {
			fmt.Printf("✓ %s downloaded successfully\n", result.AppInfo.PackageID)
		} else {
			fmt.Printf("✗ %s failed: %v\n", result.AppInfo.PackageID, result.Error)
		}
	}
}

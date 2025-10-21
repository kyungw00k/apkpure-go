# apkpure-go

A Go library and CLI tool for downloading APK files from APKPure.

This project is a Go port of the APKPure functionality from [apkeep](https://github.com/EFForg/apkeep).

## Features

- Download APK files from APKPure
- List available versions for apps
- Download specific versions
- Parallel downloads
- Progress tracking
- CSV batch processing
- Support for different architectures

## Installation

### As a CLI tool

```bash
go install github.com/kyungw00k/apkpure-go/cmd/apkpure@latest
```

### As a library

```bash
go get github.com/kyungw00k/apkpure-go
```

## Usage

### CLI

#### Download the latest version of an app

```bash
apkpure -a com.instagram.android /path/to/output
```

#### Download a specific version

```bash
apkpure -a com.instagram.android@150.0.0.0 /path/to/output
```

#### List available versions

```bash
apkpure -l -a com.instagram.android
```

#### Download from a CSV file

```bash
apkpure -c apps.csv /path/to/output
```

CSV file format (one app ID per line):
```
com.instagram.android
com.facebook.katana
com.twitter.android
```

Example CSV file is available in [`examples/test_apps.csv`](examples/test_apps.csv).

#### Advanced options

```bash
# Specify architecture
apkpure -a com.instagram.android -o arch=arm64-v8a /path/to/output

# Parallel downloads (default: 4)
apkpure -c apps.csv -r 8 /path/to/output

# Add delay between downloads (in milliseconds)
apkpure -c apps.csv -s 1000 /path/to/output
```

### Library

```go
package main

import (
    "log"
    "github.com/kyungw00k/apkpure-go/pkg/apkpure"
)

func main() {
    // Create a client
    opts := apkpure.DownloadOptions{
        Arch:             "arm64-v8a",
        ProgressCallback: apkpure.SimpleProgressCallback(),
    }
    client := apkpure.NewClient(opts)

    // Download an app
    app := apkpure.AppInfo{
        PackageID: "com.instagram.android",
        Version:   "150.0.0.0", // Optional
    }

    err := client.Download(app, "/path/to/output")
    if err != nil {
        log.Fatal(err)
    }

    // List versions
    apps := []apkpure.AppInfo{
        {PackageID: "com.instagram.android"},
    }
    err = client.ListVersions(apps)
    if err != nil {
        log.Fatal(err)
    }

    // Download multiple apps in parallel
    multipleApps := []apkpure.AppInfo{
        {PackageID: "com.instagram.android"},
        {PackageID: "com.facebook.katana"},
    }
    results := client.DownloadMultiple(multipleApps, "/path/to/output")
    for _, result := range results {
        if !result.Success {
            log.Printf("Failed: %v", result.Error)
        }
    }
}
```

## CLI Options

- `-a, --app`: App ID (e.g., `com.instagram.android` or `com.instagram.android@1.2.3`)
- `-c, --csv`: CSV file containing app IDs
- `-f, --field`: CSV field number containing app IDs (default: 1)
- `-v, --version-field`: CSV field number containing versions
- `-l, --list-versions`: List available versions
- `-o, --options`: Additional options (e.g., `arch=arm64-v8a,language=en-US`)
- `-r, --parallel`: Number of parallel downloads (default: 4)
- `-s, --sleep-duration`: Sleep duration between downloads in milliseconds

## Download Options

When using `-o` or `--options`, you can specify:

- `arch`: Architecture (e.g., `arm64-v8a`, `armeabi-v7a`, `x86`, `x86_64`)
- `language`: Language code (e.g., `en-US`, `ko-KR`)
- `os_ver`: Android OS version (e.g., `35` for Android 15)
- `output_format`: Output format for list commands (`plaintext` or `json`)

Multiple options can be combined with commas:
```bash
apkpure -a com.instagram.android -o arch=arm64-v8a,language=ko-KR /output
```

## Examples

See the [examples](examples/) directory for more usage examples.

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Credits

This project is a Go port of the APKPure functionality from [apkeep](https://github.com/EFForg/apkeep) by the Electronic Frontier Foundation (EFF).


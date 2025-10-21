package apkpure

import (
	"fmt"
	"time"
)

// ProgressTracker tracks download progress
type ProgressTracker struct {
	filename   string
	total      int64
	downloaded int64
	startTime  time.Time
	lastUpdate time.Time
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker(filename string, total int64) *ProgressTracker {
	return &ProgressTracker{
		filename:   filename,
		total:      total,
		downloaded: 0,
		startTime:  time.Now(),
		lastUpdate: time.Now(),
	}
}

// Update updates the progress and prints if needed
func (p *ProgressTracker) Update(downloaded int64) {
	p.downloaded = downloaded

	// Update every 500ms to avoid too frequent updates
	if time.Since(p.lastUpdate) < 500*time.Millisecond && downloaded < p.total {
		return
	}

	p.lastUpdate = time.Now()
	p.printProgress()
}

// printProgress prints the current progress
func (p *ProgressTracker) printProgress() {
	if p.total <= 0 {
		return
	}

	percentage := float64(p.downloaded) / float64(p.total) * 100
	elapsed := time.Since(p.startTime)

	// Calculate speed
	speed := float64(p.downloaded) / elapsed.Seconds() / 1024 / 1024 // MB/s

	fmt.Printf("\r[%s] %.1f%% (%.2f MB/s) - %s",
		p.formatElapsed(elapsed),
		percentage,
		speed,
		p.filename,
	)

	if p.downloaded >= p.total {
		fmt.Println() // New line when complete
	}
}

// formatElapsed formats elapsed time
func (p *ProgressTracker) formatElapsed(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}

// SimpleProgressCallback creates a simple progress callback function
func SimpleProgressCallback() func(filename string, downloaded, total int64) {
	trackers := make(map[string]*ProgressTracker)

	return func(filename string, downloaded, total int64) {
		tracker, exists := trackers[filename]
		if !exists {
			tracker = NewProgressTracker(filename, total)
			trackers[filename] = tracker
		}
		tracker.Update(downloaded)
	}
}

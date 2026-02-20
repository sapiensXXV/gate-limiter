package logging

import (
	"gate-limiter/config/settings"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

// RotatingWriter combines daily file rotation with size-based rotation via lumberjack.
type RotatingWriter struct {
	mu      sync.Mutex
	cfg     settings.LogFileConfig
	current *lumberjack.Logger
	today   string
}

// NewRotatingWriter creates a writer that rotates log files daily and by size.
func NewRotatingWriter(cfg settings.LogFileConfig) (*RotatingWriter, error) {
	if err := os.MkdirAll(cfg.Directory, 0755); err != nil {
		return nil, err
	}

	w := &RotatingWriter{cfg: cfg}
	today := time.Now().Format("2006-01-02")
	w.openDay(today)

	go w.cleanOld()

	return w, nil
}

func (w *RotatingWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	today := time.Now().Format("2006-01-02")
	if today != w.today {
		w.openDay(today)
		go w.cleanOld()
	}

	return w.current.Write(p)
}

// Close closes the underlying lumberjack logger.
func (w *RotatingWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.current != nil {
		return w.current.Close()
	}
	return nil
}

func (w *RotatingWriter) openDay(today string) {
	if w.current != nil {
		_ = w.current.Close()
	}

	filename := filepath.Join(w.cfg.Directory, "gl-"+today+".log")
	w.current = &lumberjack.Logger{
		Filename: filename,
		MaxSize:  w.cfg.MaxSizeMB,
		Compress: w.cfg.Compress,
	}
	w.today = today
}

func (w *RotatingWriter) cleanOld() {
	if w.cfg.MaxAgeDays <= 0 {
		return
	}

	cutoff := time.Now().AddDate(0, 0, -w.cfg.MaxAgeDays)
	entries, err := os.ReadDir(w.cfg.Directory)
	if err != nil {
		return
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !strings.HasPrefix(e.Name(), "gl-") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			_ = os.Remove(filepath.Join(w.cfg.Directory, e.Name()))
		}
	}
}

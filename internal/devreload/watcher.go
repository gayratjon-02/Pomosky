// Package devreload refreshes the browser during local development.
//
// A Watcher polls files (built frontend assets, templates) and, when something
// changes, tells every connected browser tab to reload through a
// Server-Sent Events stream. Only the standard library is used. The package is
// wired in internal/app when DEV_RELOAD is on and never in production.
package devreload

import (
	"context"
	"io/fs"
	"path/filepath"
	"sync"
	"time"
)

// Watcher polls paths for changes and notifies subscribers.
type Watcher struct {
	paths    []string
	interval time.Duration

	mu   sync.Mutex
	subs map[chan struct{}]struct{}

	done chan struct{}
	once sync.Once
}

// NewWatcher creates a Watcher that polls the given files or directories.
func NewWatcher(interval time.Duration, paths ...string) *Watcher {
	return &Watcher{
		paths:    paths,
		interval: interval,
		subs:     map[chan struct{}]struct{}{},
		done:     make(chan struct{}),
	}
}

// Run polls until ctx is cancelled or Close is called. The state at the moment
// Run starts is the baseline; only later changes trigger a notification.
func (w *Watcher) Run(ctx context.Context) {
	prev := snapshot(w.paths)
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-w.done:
			return
		case <-ticker.C:
			cur := snapshot(w.paths)
			if !sameSnapshot(prev, cur) {
				prev = cur
				w.Notify()
			}
		}
	}
}

// Notify tells every subscriber to reload now.
func (w *Watcher) Notify() {
	w.mu.Lock()
	defer w.mu.Unlock()
	for ch := range w.subs {
		select {
		case ch <- struct{}{}:
		default: // a reload is already pending for this subscriber
		}
	}
}

// Close stops Run and ends all event streams.
func (w *Watcher) Close() {
	w.once.Do(func() { close(w.done) })
}

func (w *Watcher) subscribe() (<-chan struct{}, func()) {
	ch := make(chan struct{}, 1)
	w.mu.Lock()
	w.subs[ch] = struct{}{}
	w.mu.Unlock()
	return ch, func() {
		w.mu.Lock()
		delete(w.subs, ch)
		w.mu.Unlock()
	}
}

type fileState struct {
	mod  time.Time
	size int64
}

// snapshot records modification time and size of every file under roots.
// Missing roots are ignored (for example web/dist before the first build).
func snapshot(roots []string) map[string]fileState {
	out := map[string]fileState{}
	for _, root := range roots {
		_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			if info, err := d.Info(); err == nil {
				out[p] = fileState{mod: info.ModTime(), size: info.Size()}
			}
			return nil
		})
	}
	return out
}

func sameSnapshot(a, b map[string]fileState) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if w, ok := b[k]; !ok || w != v {
			return false
		}
	}
	return true
}

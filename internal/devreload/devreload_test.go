package devreload

import (
	"bufio"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSnapshotDetectsChanges(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(file, []byte("one"), 0o644); err != nil {
		t.Fatal(err)
	}

	before := snapshot([]string{dir, filepath.Join(dir, "missing")})
	if !sameSnapshot(before, snapshot([]string{dir})) {
		t.Fatal("an unchanged tree must produce an equal snapshot")
	}

	if err := os.WriteFile(file, []byte("longer content"), 0o644); err != nil {
		t.Fatal(err)
	}
	if sameSnapshot(before, snapshot([]string{dir})) {
		t.Fatal("a modified file was not detected")
	}

	before = snapshot([]string{dir})
	if err := os.WriteFile(filepath.Join(dir, "b.txt"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if sameSnapshot(before, snapshot([]string{dir})) {
		t.Fatal("a new file was not detected")
	}
}

func TestWatcherNotifiesOnChange(t *testing.T) {
	dir := t.TempDir()
	w := NewWatcher(10*time.Millisecond, dir)
	ch, unsubscribe := w.subscribe()
	defer unsubscribe()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go w.Run(ctx)
	time.Sleep(50 * time.Millisecond) // let Run take its baseline snapshot

	if err := os.WriteFile(filepath.Join(dir, "page.html"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatal("no notification after a file change")
	}
}

func TestEventsHandlerStreamsReloadAndEndsOnClose(t *testing.T) {
	w := NewWatcher(time.Hour)
	srv := httptest.NewServer(w.EventsHandler())
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if ct := resp.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Fatalf("Content-Type = %q", ct)
	}

	lines := make(chan string, 16)
	go func() {
		sc := bufio.NewScanner(resp.Body)
		for sc.Scan() {
			lines <- sc.Text()
		}
		close(lines)
	}()

	// The handler writes "retry:" only after it subscribed, so Notify cannot be missed.
	if first := <-lines; !strings.HasPrefix(first, "retry:") {
		t.Fatalf("first line = %q", first)
	}
	w.Notify()

	got := false
	for line := range lines {
		if line == "event: reload" {
			got = true
			break
		}
	}
	if !got {
		t.Fatal("no reload event received")
	}

	w.Close()
	for range lines { // the stream must end after Close
	}
}

func TestScriptHandler(t *testing.T) {
	rec := httptest.NewRecorder()
	ScriptHandler().ServeHTTP(rec, httptest.NewRequest("GET", "/__dev/reload.js", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "EventSource") {
		t.Fatalf("unexpected response: %d %q", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/javascript") {
		t.Fatalf("Content-Type = %q", ct)
	}
}

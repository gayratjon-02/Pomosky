package devreload

import (
	"fmt"
	"net/http"
)

// EventsHandler streams "reload" events (Server-Sent Events) to the browser.
func (w *Watcher) EventsHandler() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		flusher, ok := rw.(http.Flusher)
		if !ok {
			http.Error(rw, "streaming is not supported", http.StatusInternalServerError)
			return
		}
		ch, unsubscribe := w.subscribe()
		defer unsubscribe()

		h := rw.Header()
		h.Set("Content-Type", "text/event-stream")
		h.Set("Cache-Control", "no-store")
		fmt.Fprint(rw, "retry: 500\n\n") // reconnect quickly after the server restarts
		flusher.Flush()

		for {
			select {
			case <-r.Context().Done():
				return
			case <-w.done:
				return
			case <-ch:
				fmt.Fprint(rw, "event: reload\ndata: now\n\n")
				flusher.Flush()
			}
		}
	})
}

// reloadScript reloads the page on a "reload" event and also when the
// connection comes back after having been lost (the Go server was restarted).
const reloadScript = `(() => {
  let opened = false;
  let lost = false;
  const source = new EventSource("/__dev/events");
  source.addEventListener("reload", () => location.reload());
  source.onopen = () => {
    if (lost) location.reload();
    opened = true;
  };
  source.onerror = () => {
    if (opened) lost = true;
  };
})();
`

// ScriptHandler serves the browser-side script.
func ScriptHandler() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		rw.Header().Set("Cache-Control", "no-store")
		fmt.Fprint(rw, reloadScript)
	})
}

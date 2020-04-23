package ipn

import (
	"net"
	"net/http"
	"time"
)

// HTTPAuther is used to authenticate incoming net/http requests
// by ensuring they are over the Tailscale network. One note, make
// sure to actually be listening over your tailscale interface
// otherwise this isn't effective.
func HTTPAuther(h http.Handler, cacheTTL time.Duration) http.Handler {
	return &httpAuther{q: CachedQueryer(cacheTTL), s: h}
}

type httpAuther struct {
	q Queryer
	s http.Handler
}

func (h *httpAuther) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// since we are using a cached queryer, this is an extremely fast operation.
	m, err := h.q.Query()
	if err != nil {
		http.Error(w, "failed to lookup auth list", http.StatusInternalServerError)
		return
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		http.Error(w, "invalid remote address", http.StatusInternalServerError)
		return
	}
	if _, ok := m[host]; !ok {
		http.Error(w, "not authorized", http.StatusUnauthorized)
		return
	}
	h.s.ServeHTTP(w, r)
}

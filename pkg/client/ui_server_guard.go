package client

import (
	"net/http"
	"net/url"
	"strings"
)

// dashboardRequestHeader is required on every state-changing dashboard API
// request. A cross-site browser request cannot set a custom header without a
// CORS preflight, and the dashboard never grants CORS, so this blocks CSRF
// against the loopback API. Non-browser callers (curl, scripts, the HUD) must
// send it too; GET/HEAD/OPTIONS stay open because they are read-only here.
const dashboardRequestHeader = "X-Lagvex-Request"

// isStateChangingMethod reports whether an HTTP method can modify state.
func isStateChangingMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

// isLocalOrigin accepts only loopback dashboards (including the Vite dev
// server on another port). Any internet origin is rejected even before the
// custom header check.
func isLocalOrigin(origin *url.URL) bool {
	if origin == nil {
		return false
	}
	if origin.Scheme != "http" && origin.Scheme != "https" {
		return false
	}
	host := strings.ToLower(origin.Hostname())
	return host == "127.0.0.1" || host == "localhost" || host == "::1"
}

// withLocalRequestGuard protects state-changing dashboard endpoints from
// cross-site requests and from browsers that cannot prove intent with the
// custom header. It deliberately does not pretend to authenticate local
// processes: any local process can read the dashboard and send the header.
func withLocalRequestGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isStateChangingMethod(r.Method) {
			next.ServeHTTP(w, r)
			return
		}

		if origin := r.Header.Get("Origin"); origin != "" {
			parsed, err := url.Parse(origin)
			if err != nil || !isLocalOrigin(parsed) {
				writeGuardError(w, "cross-origin request rejected")
				return
			}
		}

		if r.Header.Get(dashboardRequestHeader) != "1" {
			writeGuardError(w, "missing "+dashboardRequestHeader+" header")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func writeGuardError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	_, _ = w.Write([]byte(`{"error":"` + message + `"}`))
}

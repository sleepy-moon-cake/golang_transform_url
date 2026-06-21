package subnet

import (
	"net"
	"net/http"
)

func NewTrustedSubnetMiddleware(mask string) func(http.Handler) http.Handler {
	_, ipNet, err := net.ParseCIDR(mask)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if mask == "" || err != nil {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			clientIPStr := r.Header.Get("X-Real-IP")
			clientIP := net.ParseIP(clientIPStr)

			if clientIP == nil || !ipNet.Contains(clientIP) {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

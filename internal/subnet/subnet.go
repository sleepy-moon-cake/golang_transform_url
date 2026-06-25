package subnet

import (
	"fmt"
	"net"
	"net/http"
)

func NewTrustedSubnetMiddleware(mask string) (func(http.Handler) http.Handler, error) {
	if mask == "" {
		return nil, nil
	}

	_, ipNet, err := net.ParseCIDR(mask)

	if err != nil {
		return nil, fmt.Errorf("parseCIDR:%w", err)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			clientIPStr := r.Header.Get("X-Real-IP")
			clientIP := net.ParseIP(clientIPStr)

			if clientIP == nil || !ipNet.Contains(clientIP) {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}, nil
}

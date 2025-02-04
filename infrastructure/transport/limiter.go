package transport

import (
	"net/http"

	"golang.org/x/time/rate"
)

type RateLimitTransport struct {
	limiter *rate.Limiter
	rTriper http.RoundTripper
}

func NewRateLimitTransport(r float64, rTriper http.RoundTripper) http.RoundTripper {
	return &RateLimitTransport{
		limiter: rate.NewLimiter(rate.Limit(r), 1),
		rTriper: rTriper,
	}
}

func (t *RateLimitTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	t.limiter.Wait(r.Context())
	return t.rTriper.RoundTrip(r)
}

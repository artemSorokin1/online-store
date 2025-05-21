package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type AuthMetrics struct {
	requestDuration *prometheus.HistogramVec
	requestCount    prometheus.Counter
}

func NewAuthMetrics() *AuthMetrics {
	var am = &AuthMetrics{
		requestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "auth_service_request_duration_seconds",
				Help: "Duration of HTTP requests in seconds",
			},
			[]string{"method", "path", "status"},
		),
		requestCount: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "auth_service_request_total",
				Help: "Total number of HTTP requests",
			},
		),
	}
	prometheus.MustRegister(am.requestDuration)
	prometheus.MustRegister(am.requestCount)

	return am
}

func (am *AuthMetrics) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		duration := time.Since(start).Seconds()
		statusCode := http.StatusOK

		am.requestDuration.WithLabelValues(
			r.Method,
			r.URL.Path,
			strconv.Itoa(statusCode),
		).Observe(duration)
		am.requestCount.Inc()
	})
}

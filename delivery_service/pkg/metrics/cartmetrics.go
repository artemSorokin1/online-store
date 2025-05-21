package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type CartMetrics struct {
	requestDuration *prometheus.HistogramVec
	requestCount    prometheus.Counter
}

func NewCartMetrics() *CartMetrics {
	var pm = &CartMetrics{
		requestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "cart_service_request_duration_seconds",
				Help: "Duration of HTTP requests in seconds",
			},
			[]string{"method", "path", "status"},
		),
		requestCount: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "cart_service_request_total",
				Help: "Total number of HTTP requests",
			},
		),
	}
	prometheus.MustRegister(pm.requestDuration)
	prometheus.MustRegister(pm.requestCount)
	return pm
}

func (m *CartMetrics) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		duration := time.Since(start).Seconds()
		statusCode := http.StatusOK

		m.requestDuration.WithLabelValues(
			r.Method,
			r.URL.Path,
			strconv.Itoa(statusCode),
		).Observe(duration)
		m.requestCount.Inc()
	})
}

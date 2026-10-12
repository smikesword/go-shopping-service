package main

import (
	"github.com/prometheus/client_golang/prometheus"
	// "github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"time"
)

type Metrics struct {
	RequestCounter *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec
}

func NewMetrics() *Metrics {
  metrics := Metrics{}
  metrics.RequestCounter = prometheus.NewCounterVec(
    prometheus.CounterOpts{
      Name: "http_requests_total",
      Help: "Total number of HTTP requests received",
    },
    []string{"method", "path"},
  )
  metrics.RequestDuration = prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
      Name: "http_request_duration_seconds",
      Help: "Histogram of response durations",
      Buckets: prometheus.DefBuckets,
    },
    []string{"method", "path"},
  )
  prometheus.MustRegister(metrics.RequestCounter)
  prometheus.MustRegister(metrics.RequestDuration)

  return &metrics
}

func MetricsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start).Seconds()
		metricsService.RequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
		metricsService.RequestCounter.WithLabelValues(r.Method, r.URL.Path).Inc()
	})
}
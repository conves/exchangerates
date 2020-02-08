package main

import (
	"flag"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

var (
	// Flags
	addr     = flag.String("addr", ":8080", "HTTP server address")
	timeout  = flag.Int("timeout", 500, "HTTP client timeout")
	cacheTtl = flag.Int("cache-ttl", 3600, "cache ttl value")

	// Api url template
	apiurltpl = "https://api.exchangeratesapi.io/%s"

	// Prometheus metrics
	errorCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "exchangerate_http_client_errors_total",
			Help: "HTTP client errors",
		},
		[]string{"err_type"},
	)
	cacheHits = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "exchangerate_cache_hits_total",
			Help: "Cache hits",
		},
	)
	cacheMisses = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "exchangerate_cache_misses_total",
			Help: "Cache misses",
		},
	)

	// Logger
	logger = logrus.New()
)

func main() {
	flag.Parse()

	// Build an in-mem Cache; in a future iteration, we could write a Redis or Memcached implementation of Cache interface
	cache := NewImMemCache()
	defer cache.Close()

	http.Handle("/metrics", promhttp.Handler())

	http.HandleFunc("/recommend", recommendHandler(cache, http.DefaultClient))

	// Liveness probe
	http.HandleFunc("/healthz", func (w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Mocked readiness probe
	http.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	logger.Fatalf("HTTP server crashed: %s", http.ListenAndServe(*addr, nil))
}
package services

import (
	"net/http"
	"time"

	"github.com/deezer/groroti/internal/model"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	total_rotis = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "groroti_total_rotis",
		Help: "All the ROTIs that have been created since the beginning (including deleted ones)",
	})
	active_rotis = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "groroti_active_rotis",
		Help: "All the ROTIs that have been created and not deleted",
	})
)

// NewMetricsHandler creates the handler allowing to dump Prometheus metrics
func NewMetricsHandler() http.Handler {
	prometheus.MustRegister(total_rotis)
	prometheus.MustRegister(active_rotis)

	return promhttp.Handler()
}

// create a goroutine that will periodically query the DB
// this allows to avoid querying too much the db
func recordMetrics() {
	go func() {
		for {
			total_rotis.Set(float64(model.GetMaxROTIID()))
			active_rotis.Set(float64(model.CountROTIs()))
			time.Sleep(15 * time.Second)
		}
	}()
}

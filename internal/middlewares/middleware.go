package middlewares

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	responseTimeHistogram = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "promhttp_metric_handler_response_time",
			Help:    "Global response time",
			Buckets: []float64{.001, .002, .005, .01, .025, .05, .1, .25, .5, 1, 5, 10},
		},
		[]string{"code", "method", "name"},
	)
	requestSizeHistogram = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "promhttp_metric_handler_request_size",
			Help:    "Size of the request",
			Buckets: []float64{10, 100, 1000, 10000, 100000, 1000000, 10000000, 100000000},
		},
		[]string{"code", "method", "name"},
	)
	responseSizeHistogram = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "promhttp_metric_handler_response_size",
			Help:    "size of the response sent",
			Buckets: []float64{10, 100, 1000, 10000, 100000, 1000000, 10000000, 100000000},
		},
		[]string{"code", "method", "name"},
	)
	timeToWriteHeaderHistogram = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "promhttp_metric_handler_time_to_write_header",
			Help:    "time between start of request processing and the first header written and responded",
			Buckets: []float64{.001, .002, .005, .01, .025, .05, .1, .25, .5, 1, 5, 10},
		},
		[]string{"code", "method", "name"},
	)
)

// PrometheusInstrumentation This middle ware is sensitive to order in the middleware chain. For a close to real
// timing measurement. It must be as high in the chain as possible, be the first one declared.
func PrometheusInstrumentation(name string, next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		promhttp.InstrumentHandlerDuration(responseTimeHistogram.MustCurryWith(prometheus.Labels{"name": name}),
			promhttp.InstrumentHandlerRequestSize(requestSizeHistogram.MustCurryWith(prometheus.Labels{"name": name}),
				promhttp.InstrumentHandlerResponseSize(responseSizeHistogram.MustCurryWith(prometheus.Labels{"name": name}),
					promhttp.InstrumentHandlerTimeToWriteHeader(timeToWriteHeaderHistogram.MustCurryWith(prometheus.Labels{"name": name}), next),
				),
			),
		).ServeHTTP(w, r)
	})
}

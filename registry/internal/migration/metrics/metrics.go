package metrics

import (
	"github.com/docker/distribution/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	tagCountHist *prometheus.HistogramVec
)

const (
	subsystem       = "migration"
	importTypeLabel = "import_type"

	tagCountName = "tag_counts"
	tagCountDesc = "A histogram of tag counts per repository (pre)import."
)

func init() {
	tagCountHist = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metrics.NamespacePrefix,
			Subsystem: subsystem,
			Name:      tagCountName,
			Help:      tagCountDesc,
			Buckets:   []float64{0, 1, 2, 5, 10, 25, 50, 100, 250, 500, 1000, 2000, 5000, 10000, 15000, 20000, 50000, 100000},
		},
		[]string{importTypeLabel},
	)

	prometheus.MustRegister(tagCountHist)
}

type importType string

const (
	ImportTypePre   importType = "pre"
	ImportTypeFinal importType = "final"
)

func (t importType) String() string {
	return string(t)
}
func TagCount(t importType, count int) {
	tagCountHist.WithLabelValues(t.String()).Observe(float64(count))
}

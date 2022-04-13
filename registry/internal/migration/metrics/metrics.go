package metrics

import (
	"github.com/docker/distribution/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	tagCountHist             *prometheus.HistogramVec
	layerCountHist           prometheus.Histogram
	blobTransferDurationHist *prometheus.HistogramVec
	blobTransferSizeHist     *prometheus.HistogramVec
)

const (
	subsystem       = "migration"
	importTypeLabel = "import_type"
	blobTypeLabel   = "blob_type"

	tagCountName = "tag_counts"
	tagCountDesc = "A histogram of tag counts per repository (pre)import."

	layerCountName = "layer_counts"
	layerCountDesc = "A histogram of layer counts per (pre)imported manifest."

	blobTransferDurationName = "blob_transfer_duration_seconds"
	blobTransferDurationDesc = "A histogram of latencies for blob transfers."

	blobTransferSizeName = "blob_transfer_size_bytes"
	blobTransferSizeDesc = "A histogram of byte sizes for blob transfers."
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

	layerCountHist = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: metrics.NamespacePrefix,
			Subsystem: subsystem,
			Name:      layerCountName,
			Help:      layerCountDesc,
			Buckets:   []float64{1, 2, 5, 10, 25, 50, 100, 200},
		},
	)

	blobTransferDurationHist = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metrics.NamespacePrefix,
			Subsystem: subsystem,
			Name:      blobTransferDurationName,
			Help:      blobTransferDurationDesc,
			Buckets:   prometheus.DefBuckets,
		},
		[]string{blobTypeLabel},
	)

	blobTransferSizeHist = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metrics.NamespacePrefix,
			Subsystem: subsystem,
			Name:      blobTransferSizeName,
			Help:      blobTransferSizeDesc,
			Buckets: []float64{
				512 * 1024,              // 512KiB
				1024 * 1024,             // 1MiB
				1024 * 1024 * 64,        // 64MiB
				1024 * 1024 * 128,       // 128MiB
				1024 * 1024 * 256,       // 256MiB
				1024 * 1024 * 512,       // 512MiB
				1024 * 1024 * 1024,      // 1GiB
				1024 * 1024 * 1024 * 2,  // 2GiB
				1024 * 1024 * 1024 * 3,  // 3GiB
				1024 * 1024 * 1024 * 4,  // 4GiB
				1024 * 1024 * 1024 * 5,  // 5GiB
				1024 * 1024 * 1024 * 6,  // 6GiB
				1024 * 1024 * 1024 * 7,  // 7GiB
				1024 * 1024 * 1024 * 8,  // 8GiB
				1024 * 1024 * 1024 * 9,  // 9GiB
				1024 * 1024 * 1024 * 10, // 10GiB
				1024 * 1024 * 1024 * 20, // 20GiB
				1024 * 1024 * 1024 * 30, // 30GiB
				1024 * 1024 * 1024 * 40, // 40GiB
				1024 * 1024 * 1024 * 50, // 50GiB
			},
		},
		[]string{blobTypeLabel},
	)

	prometheus.MustRegister(tagCountHist)
	prometheus.MustRegister(layerCountHist)
	prometheus.MustRegister(blobTransferDurationHist)
	prometheus.MustRegister(blobTransferSizeHist)
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

func LayerCount(count int) {
	layerCountHist.Observe(float64(count))
}

type BlobType string

const (
	BlobTypeConfig  BlobType = "config"
	BlobTypeLayer   BlobType = "layer"
	BlobTypeUnknown BlobType = "unknown"
)

func (t BlobType) String() string {
	return string(t)
}

func BlobTransfer(duration float64, size float64, t BlobType) {
	blobTransferDurationHist.WithLabelValues(t.String()).Observe(duration)
	blobTransferSizeHist.WithLabelValues(t.String()).Observe(size)
}

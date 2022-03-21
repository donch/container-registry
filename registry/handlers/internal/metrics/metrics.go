package metrics

import (
	"strconv"
	"time"

	"github.com/docker/distribution/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	migrationRoutingCounter *prometheus.CounterVec
	importDurationHist      *prometheus.HistogramVec
	importCounter           *prometheus.CounterVec
	inFlightImports         *prometheus.GaugeVec
	concurrentImports       prometheus.Gauge

	timeSince = time.Since // for test purposes only
)

const (
	subsystem = "http"

	codePathLabel    = "path"
	oldCodePathValue = "old"
	newCodePathValue = "new"

	errorLabel      = "error"
	importTypeLabel = "type"
	importAttempted = "attempted"

	importTypeValue    = "import"
	preImportTypeValue = "pre_import"

	migrationRouteTotalName = "request_migration_route_total"
	migrationRouteTotalDesc = "A counter for code path routing of requests during migration."

	importDurationName    = "import_duration_seconds"
	importDurationDesc    = "A histogram of durations for imports."
	importTotalName       = "imports_total"
	importTotalDesc       = "A counter for API triggered imports."
	inFlightImportsName   = "in_flight_imports"
	inFlightImportsDesc   = "A gauge of imports currently being undertaken by the registry."
	concurrentImportsName = "import_worker_saturation"
	concurrentImportsDesc = "A gauge of saturation of workers per instance."
)

func init() {
	migrationRoutingCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metrics.NamespacePrefix,
			Subsystem: subsystem,
			Name:      migrationRouteTotalName,
			Help:      migrationRouteTotalDesc,
		},
		[]string{codePathLabel},
	)

	importDurationHist = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metrics.NamespacePrefix,
			Subsystem: subsystem,
			Name:      importDurationName,
			Help:      importDurationDesc,
			// 5ms to 6h
			Buckets: []float64{.005, .01, .025, .5, 1, 5, 15, 30, 60, 300, 600, 900, 1800, 3600, 7200, 10800, 21600},
		},
		[]string{importTypeLabel, errorLabel, importAttempted},
	)

	importCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metrics.NamespacePrefix,
			Subsystem: subsystem,
			Name:      importTotalName,
			Help:      importTotalDesc,
		},
		[]string{importTypeLabel, errorLabel, importAttempted},
	)

	inFlightImports = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metrics.NamespacePrefix,
		Subsystem: subsystem,
		Name:      inFlightImportsName,
		Help:      inFlightImportsDesc,
	},
		[]string{importTypeLabel},
	)

	concurrentImports = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: metrics.NamespacePrefix,
			Subsystem: subsystem,
			Name:      concurrentImportsName,
			Help:      concurrentImportsDesc,
		},
	)

	prometheus.MustRegister(migrationRoutingCounter)
	prometheus.MustRegister(importDurationHist)
	prometheus.MustRegister(importCounter)
	prometheus.MustRegister(inFlightImports)
	prometheus.MustRegister(concurrentImports)
}

func MigrationRoute(newCodePath bool) {
	var codePath string
	if newCodePath {
		codePath = newCodePathValue
	} else {
		codePath = oldCodePathValue
	}
	migrationRoutingCounter.WithLabelValues(codePath).Inc()
}

type ImportReportFunc func(importAttempted bool, err error)

func Import() ImportReportFunc {
	return doImport(importTypeValue)
}

func PreImport() ImportReportFunc {
	return doImport(preImportTypeValue)
}

func doImport(importType string) func(importAttempted bool, err error) {
	g := inFlightImports.WithLabelValues(importType)

	g.Inc()
	start := time.Now()
	return func(importAttempted bool, err error) {
		g.Dec()

		failed := strconv.FormatBool(err != nil)
		attempted := strconv.FormatBool(importAttempted)

		importDurationHist.WithLabelValues(importType, failed, attempted).Observe(timeSince(start).Seconds())
		importCounter.WithLabelValues(importType, failed, attempted).Inc()
	}
}

func ImportWorkerSaturation(percentage float64) {
	concurrentImports.Set(percentage)
}

package metrics

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/docker/distribution/metrics"
	"github.com/prometheus/client_golang/prometheus"
	testutil "github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func mockTimeSince(d time.Duration) func() {
	bkp := timeSince
	timeSince = func(_ time.Time) time.Duration { return d }
	return func() { timeSince = bkp }
}

type metricsSuite struct{ suite.Suite }

// Reset counters so that tests do not interact with each other.
func (suite *metricsSuite) SetupTest() {
	migrationRoutingCounter.Reset()
	importDurationHist.Reset()
	importCounter.Reset()
	inFlightImports.Reset()
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestMetrics(t *testing.T) {
	suite.Run(t, new(metricsSuite))
}

func (suite *metricsSuite) TestMigrationRoute() {
	restore := mockTimeSince(10 * time.Millisecond)
	defer restore()

	MigrationRoute(true)
	MigrationRoute(true)

	mockTimeSince(20 * time.Millisecond)
	MigrationRoute(false)

	var expected bytes.Buffer
	expected.WriteString(`
# HELP registry_http_request_migration_route_total A counter for code path routing of requests during migration.
# TYPE registry_http_request_migration_route_total counter
registry_http_request_migration_route_total{path="new"} 2
registry_http_request_migration_route_total{path="old"} 1
`)
	fullName := fmt.Sprintf("%s_%s_%s", metrics.NamespacePrefix, subsystem, migrationRouteTotalName)

	err := testutil.GatherAndCompare(prometheus.DefaultGatherer, &expected, fullName)
	require.NoError(suite.T(), err)
}

func (suite *metricsSuite) TestImport() {
	restore := mockTimeSince(10 * time.Millisecond)
	defer restore()

	report := Import()
	report(true, errors.New("foo"))
	report(true, errors.New("foo"))
	report(true, nil)
	report(false, errors.New("foo"))
	report(false, nil)

	var expected bytes.Buffer
	expected.WriteString(`
# HELP registry_http_import_duration_seconds A histogram of durations for imports.
# TYPE registry_http_import_duration_seconds histogram
registry_http_import_duration_seconds_bucket{attempted="false",error="false",type="import",le="0.005"} 0
registry_http_import_duration_seconds_bucket{attempted="false",error="false",type="import",le="0.01"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="false",type="import",le="0.025"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="false",type="import",le="0.5"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="false",type="import",le="1"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="false",type="import",le="5"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="false",type="import",le="15"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="false",type="import",le="30"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="false",type="import",le="60"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="false",type="import",le="300"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="false",type="import",le="600"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="false",type="import",le="900"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="false",type="import",le="1800"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="false",type="import",le="3600"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="false",type="import",le="7200"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="false",type="import",le="10800"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="false",type="import",le="21600"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="false",type="import",le="+Inf"} 1
registry_http_import_duration_seconds_sum{attempted="false",error="false",type="import"} 0.01
registry_http_import_duration_seconds_count{attempted="false",error="false",type="import"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="true",type="import",le="0.005"} 0
registry_http_import_duration_seconds_bucket{attempted="false",error="true",type="import",le="0.01"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="true",type="import",le="0.025"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="true",type="import",le="0.5"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="true",type="import",le="1"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="true",type="import",le="5"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="true",type="import",le="15"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="true",type="import",le="30"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="true",type="import",le="60"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="true",type="import",le="300"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="true",type="import",le="600"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="true",type="import",le="900"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="true",type="import",le="1800"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="true",type="import",le="3600"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="true",type="import",le="7200"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="true",type="import",le="10800"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="true",type="import",le="21600"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="true",type="import",le="+Inf"} 1
registry_http_import_duration_seconds_sum{attempted="false",error="true",type="import"} 0.01
registry_http_import_duration_seconds_count{attempted="false",error="true",type="import"} 1
registry_http_import_duration_seconds_bucket{attempted="true",error="false",type="import",le="0.005"} 0
registry_http_import_duration_seconds_bucket{attempted="true",error="false",type="import",le="0.01"} 1
registry_http_import_duration_seconds_bucket{attempted="true",error="false",type="import",le="0.025"} 1
registry_http_import_duration_seconds_bucket{attempted="true",error="false",type="import",le="0.5"} 1
registry_http_import_duration_seconds_bucket{attempted="true",error="false",type="import",le="1"} 1
registry_http_import_duration_seconds_bucket{attempted="true",error="false",type="import",le="5"} 1
registry_http_import_duration_seconds_bucket{attempted="true",error="false",type="import",le="15"} 1
registry_http_import_duration_seconds_bucket{attempted="true",error="false",type="import",le="30"} 1
registry_http_import_duration_seconds_bucket{attempted="true",error="false",type="import",le="60"} 1
registry_http_import_duration_seconds_bucket{attempted="true",error="false",type="import",le="300"} 1
registry_http_import_duration_seconds_bucket{attempted="true",error="false",type="import",le="600"} 1
registry_http_import_duration_seconds_bucket{attempted="true",error="false",type="import",le="900"} 1
registry_http_import_duration_seconds_bucket{attempted="true",error="false",type="import",le="1800"} 1
registry_http_import_duration_seconds_bucket{attempted="true",error="false",type="import",le="3600"} 1
registry_http_import_duration_seconds_bucket{attempted="true",error="false",type="import",le="7200"} 1
registry_http_import_duration_seconds_bucket{attempted="true",error="false",type="import",le="10800"} 1
registry_http_import_duration_seconds_bucket{attempted="true",error="false",type="import",le="21600"} 1
registry_http_import_duration_seconds_bucket{attempted="true",error="false",type="import",le="+Inf"} 1
registry_http_import_duration_seconds_sum{attempted="true",error="false",type="import"} 0.01
registry_http_import_duration_seconds_count{attempted="true",error="false",type="import"} 1
registry_http_import_duration_seconds_bucket{attempted="true",error="true",type="import",le="0.005"} 0
registry_http_import_duration_seconds_bucket{attempted="true",error="true",type="import",le="0.01"} 2
registry_http_import_duration_seconds_bucket{attempted="true",error="true",type="import",le="0.025"} 2
registry_http_import_duration_seconds_bucket{attempted="true",error="true",type="import",le="0.5"} 2
registry_http_import_duration_seconds_bucket{attempted="true",error="true",type="import",le="1"} 2
registry_http_import_duration_seconds_bucket{attempted="true",error="true",type="import",le="5"} 2
registry_http_import_duration_seconds_bucket{attempted="true",error="true",type="import",le="15"} 2
registry_http_import_duration_seconds_bucket{attempted="true",error="true",type="import",le="30"} 2
registry_http_import_duration_seconds_bucket{attempted="true",error="true",type="import",le="60"} 2
registry_http_import_duration_seconds_bucket{attempted="true",error="true",type="import",le="300"} 2
registry_http_import_duration_seconds_bucket{attempted="true",error="true",type="import",le="600"} 2
registry_http_import_duration_seconds_bucket{attempted="true",error="true",type="import",le="900"} 2
registry_http_import_duration_seconds_bucket{attempted="true",error="true",type="import",le="1800"} 2
registry_http_import_duration_seconds_bucket{attempted="true",error="true",type="import",le="3600"} 2
registry_http_import_duration_seconds_bucket{attempted="true",error="true",type="import",le="7200"} 2
registry_http_import_duration_seconds_bucket{attempted="true",error="true",type="import",le="10800"} 2
registry_http_import_duration_seconds_bucket{attempted="true",error="true",type="import",le="21600"} 2
registry_http_import_duration_seconds_bucket{attempted="true",error="true",type="import",le="+Inf"} 2
registry_http_import_duration_seconds_sum{attempted="true",error="true",type="import"} 0.02
registry_http_import_duration_seconds_count{attempted="true",error="true",type="import"} 2
# HELP registry_http_imports_total A counter for API triggered imports.
# TYPE registry_http_imports_total counter
registry_http_imports_total{attempted="false",error="false",type="import"} 1
registry_http_imports_total{attempted="false",error="true",type="import"} 1
registry_http_imports_total{attempted="true",error="false",type="import"} 1
registry_http_imports_total{attempted="true",error="true",type="import"} 2
`)
	durationFullName := fmt.Sprintf("%s_%s_%s", metrics.NamespacePrefix, subsystem, importDurationName)
	totalFullName := fmt.Sprintf("%s_%s_%s", metrics.NamespacePrefix, subsystem, importTotalName)

	err := testutil.GatherAndCompare(prometheus.DefaultGatherer, &expected, durationFullName, totalFullName)
	require.NoError(suite.T(), err)
}

func (suite *metricsSuite) TestPreImport() {
	restore := mockTimeSince(10 * time.Millisecond)
	defer restore()

	report := PreImport()
	report(true, errors.New("foo"))
	report(true, errors.New("foo"))
	report(true, nil)
	report(false, errors.New("foo"))
	report(false, nil)

	var expected bytes.Buffer
	expected.WriteString(`
# HELP registry_http_import_duration_seconds A histogram of durations for imports.
# TYPE registry_http_import_duration_seconds histogram
registry_http_import_duration_seconds_bucket{attempted="false",error="false",type="pre_import",le="0.005"} 0
registry_http_import_duration_seconds_bucket{attempted="false",error="false",type="pre_import",le="0.01"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="false",type="pre_import",le="0.025"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="false",type="pre_import",le="0.5"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="false",type="pre_import",le="1"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="false",type="pre_import",le="5"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="false",type="pre_import",le="15"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="false",type="pre_import",le="30"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="false",type="pre_import",le="60"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="false",type="pre_import",le="300"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="false",type="pre_import",le="600"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="false",type="pre_import",le="900"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="false",type="pre_import",le="1800"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="false",type="pre_import",le="3600"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="false",type="pre_import",le="7200"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="false",type="pre_import",le="10800"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="false",type="pre_import",le="21600"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="false",type="pre_import",le="+Inf"} 1
registry_http_import_duration_seconds_sum{attempted="false",error="false",type="pre_import"} 0.01
registry_http_import_duration_seconds_count{attempted="false",error="false",type="pre_import"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="true",type="pre_import",le="0.005"} 0
registry_http_import_duration_seconds_bucket{attempted="false",error="true",type="pre_import",le="0.01"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="true",type="pre_import",le="0.025"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="true",type="pre_import",le="0.5"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="true",type="pre_import",le="1"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="true",type="pre_import",le="5"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="true",type="pre_import",le="15"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="true",type="pre_import",le="30"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="true",type="pre_import",le="60"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="true",type="pre_import",le="300"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="true",type="pre_import",le="600"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="true",type="pre_import",le="900"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="true",type="pre_import",le="1800"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="true",type="pre_import",le="3600"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="true",type="pre_import",le="7200"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="true",type="pre_import",le="10800"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="true",type="pre_import",le="21600"} 1
registry_http_import_duration_seconds_bucket{attempted="false",error="true",type="pre_import",le="+Inf"} 1
registry_http_import_duration_seconds_sum{attempted="false",error="true",type="pre_import"} 0.01
registry_http_import_duration_seconds_count{attempted="false",error="true",type="pre_import"} 1
registry_http_import_duration_seconds_bucket{attempted="true",error="false",type="pre_import",le="0.005"} 0
registry_http_import_duration_seconds_bucket{attempted="true",error="false",type="pre_import",le="0.01"} 1
registry_http_import_duration_seconds_bucket{attempted="true",error="false",type="pre_import",le="0.025"} 1
registry_http_import_duration_seconds_bucket{attempted="true",error="false",type="pre_import",le="0.5"} 1
registry_http_import_duration_seconds_bucket{attempted="true",error="false",type="pre_import",le="1"} 1
registry_http_import_duration_seconds_bucket{attempted="true",error="false",type="pre_import",le="5"} 1
registry_http_import_duration_seconds_bucket{attempted="true",error="false",type="pre_import",le="15"} 1
registry_http_import_duration_seconds_bucket{attempted="true",error="false",type="pre_import",le="30"} 1
registry_http_import_duration_seconds_bucket{attempted="true",error="false",type="pre_import",le="60"} 1
registry_http_import_duration_seconds_bucket{attempted="true",error="false",type="pre_import",le="300"} 1
registry_http_import_duration_seconds_bucket{attempted="true",error="false",type="pre_import",le="600"} 1
registry_http_import_duration_seconds_bucket{attempted="true",error="false",type="pre_import",le="900"} 1
registry_http_import_duration_seconds_bucket{attempted="true",error="false",type="pre_import",le="1800"} 1
registry_http_import_duration_seconds_bucket{attempted="true",error="false",type="pre_import",le="3600"} 1
registry_http_import_duration_seconds_bucket{attempted="true",error="false",type="pre_import",le="7200"} 1
registry_http_import_duration_seconds_bucket{attempted="true",error="false",type="pre_import",le="10800"} 1
registry_http_import_duration_seconds_bucket{attempted="true",error="false",type="pre_import",le="21600"} 1
registry_http_import_duration_seconds_bucket{attempted="true",error="false",type="pre_import",le="+Inf"} 1
registry_http_import_duration_seconds_sum{attempted="true",error="false",type="pre_import"} 0.01
registry_http_import_duration_seconds_count{attempted="true",error="false",type="pre_import"} 1
registry_http_import_duration_seconds_bucket{attempted="true",error="true",type="pre_import",le="0.005"} 0
registry_http_import_duration_seconds_bucket{attempted="true",error="true",type="pre_import",le="0.01"} 2
registry_http_import_duration_seconds_bucket{attempted="true",error="true",type="pre_import",le="0.025"} 2
registry_http_import_duration_seconds_bucket{attempted="true",error="true",type="pre_import",le="0.5"} 2
registry_http_import_duration_seconds_bucket{attempted="true",error="true",type="pre_import",le="1"} 2
registry_http_import_duration_seconds_bucket{attempted="true",error="true",type="pre_import",le="5"} 2
registry_http_import_duration_seconds_bucket{attempted="true",error="true",type="pre_import",le="15"} 2
registry_http_import_duration_seconds_bucket{attempted="true",error="true",type="pre_import",le="30"} 2
registry_http_import_duration_seconds_bucket{attempted="true",error="true",type="pre_import",le="60"} 2
registry_http_import_duration_seconds_bucket{attempted="true",error="true",type="pre_import",le="300"} 2
registry_http_import_duration_seconds_bucket{attempted="true",error="true",type="pre_import",le="600"} 2
registry_http_import_duration_seconds_bucket{attempted="true",error="true",type="pre_import",le="900"} 2
registry_http_import_duration_seconds_bucket{attempted="true",error="true",type="pre_import",le="1800"} 2
registry_http_import_duration_seconds_bucket{attempted="true",error="true",type="pre_import",le="3600"} 2
registry_http_import_duration_seconds_bucket{attempted="true",error="true",type="pre_import",le="7200"} 2
registry_http_import_duration_seconds_bucket{attempted="true",error="true",type="pre_import",le="10800"} 2
registry_http_import_duration_seconds_bucket{attempted="true",error="true",type="pre_import",le="21600"} 2
registry_http_import_duration_seconds_bucket{attempted="true",error="true",type="pre_import",le="+Inf"} 2
registry_http_import_duration_seconds_sum{attempted="true",error="true",type="pre_import"} 0.02
registry_http_import_duration_seconds_count{attempted="true",error="true",type="pre_import"} 2
# HELP registry_http_imports_total A counter for API triggered imports.
# TYPE registry_http_imports_total counter
registry_http_imports_total{attempted="false",error="false",type="pre_import"} 1
registry_http_imports_total{attempted="false",error="true",type="pre_import"} 1
registry_http_imports_total{attempted="true",error="false",type="pre_import"} 1
registry_http_imports_total{attempted="true",error="true",type="pre_import"} 2
`)
	durationFullName := fmt.Sprintf("%s_%s_%s", metrics.NamespacePrefix, subsystem, importDurationName)
	totalFullName := fmt.Sprintf("%s_%s_%s", metrics.NamespacePrefix, subsystem, importTotalName)

	err := testutil.GatherAndCompare(prometheus.DefaultGatherer, &expected, durationFullName, totalFullName)
	require.NoError(suite.T(), err)
}

func (suite *metricsSuite) TestInFlightImports() {
	t := suite.T()

	restore := mockTimeSince(10 * time.Millisecond)
	defer restore()

	type rptFunc func(importAttempted bool, err error)
	reports := []rptFunc{}

	// Open five imports.
	reports = append(reports, PreImport())
	reports = append(reports, Import())
	reports = append(reports, PreImport())
	reports = append(reports, Import())
	reports = append(reports, PreImport())

	var expected bytes.Buffer
	expected.WriteString(`
# HELP registry_http_in_flight_imports A gauge of imports currently being undertaken by the registry.
# TYPE registry_http_in_flight_imports gauge
registry_http_in_flight_imports{type="import"} 2
registry_http_in_flight_imports{type="pre_import"} 3
`)

	inFlightFullName := fmt.Sprintf("%s_%s_%s", metrics.NamespacePrefix, subsystem, inFlightImportsName)

	err := testutil.GatherAndCompare(prometheus.DefaultGatherer, &expected, inFlightFullName)
	require.NoError(t, err)

	// Simulate two imports finishing.
	reports[0](true, nil)
	reports[1](true, nil)

	expected.WriteString(`
# HELP registry_http_in_flight_imports A gauge of imports currently being undertaken by the registry.
# TYPE registry_http_in_flight_imports gauge
registry_http_in_flight_imports{type="import"} 1
registry_http_in_flight_imports{type="pre_import"} 2
`)

	err = testutil.GatherAndCompare(prometheus.DefaultGatherer, &expected, inFlightFullName)
	require.NoError(t, err)

	// Simulate remaining imports finishing.
	reports[2](true, nil)
	reports[3](true, nil)
	reports[4](true, nil)

	expected.WriteString(`
# HELP registry_http_in_flight_imports A gauge of imports currently being undertaken by the registry.
# TYPE registry_http_in_flight_imports gauge
registry_http_in_flight_imports{type="import"} 0
registry_http_in_flight_imports{type="pre_import"} 0
`)

	err = testutil.GatherAndCompare(prometheus.DefaultGatherer, &expected, inFlightFullName)
	require.NoError(t, err)
}

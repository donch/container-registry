package metrics

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/docker/distribution/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
)

func TestTagCount(t *testing.T) {
	TagCount(ImportTypePre, 10)
	TagCount(ImportTypeFinal, 9)

	var expected bytes.Buffer
	expected.WriteString(`
# HELP registry_migration_tag_counts A histogram of tag counts per repository (pre)import.
# TYPE registry_migration_tag_counts histogram
registry_migration_tag_counts_bucket{import_type="final",le="0"} 0
registry_migration_tag_counts_bucket{import_type="final",le="1"} 0
registry_migration_tag_counts_bucket{import_type="final",le="2"} 0
registry_migration_tag_counts_bucket{import_type="final",le="5"} 0
registry_migration_tag_counts_bucket{import_type="final",le="10"} 1
registry_migration_tag_counts_bucket{import_type="final",le="25"} 1
registry_migration_tag_counts_bucket{import_type="final",le="50"} 1
registry_migration_tag_counts_bucket{import_type="final",le="100"} 1
registry_migration_tag_counts_bucket{import_type="final",le="250"} 1
registry_migration_tag_counts_bucket{import_type="final",le="500"} 1
registry_migration_tag_counts_bucket{import_type="final",le="1000"} 1
registry_migration_tag_counts_bucket{import_type="final",le="2000"} 1
registry_migration_tag_counts_bucket{import_type="final",le="5000"} 1
registry_migration_tag_counts_bucket{import_type="final",le="10000"} 1
registry_migration_tag_counts_bucket{import_type="final",le="15000"} 1
registry_migration_tag_counts_bucket{import_type="final",le="20000"} 1
registry_migration_tag_counts_bucket{import_type="final",le="50000"} 1
registry_migration_tag_counts_bucket{import_type="final",le="100000"} 1
registry_migration_tag_counts_bucket{import_type="final",le="+Inf"} 1
registry_migration_tag_counts_sum{import_type="final"} 9
registry_migration_tag_counts_count{import_type="final"} 1
registry_migration_tag_counts_bucket{import_type="pre",le="0"} 0
registry_migration_tag_counts_bucket{import_type="pre",le="1"} 0
registry_migration_tag_counts_bucket{import_type="pre",le="2"} 0
registry_migration_tag_counts_bucket{import_type="pre",le="5"} 0
registry_migration_tag_counts_bucket{import_type="pre",le="10"} 1
registry_migration_tag_counts_bucket{import_type="pre",le="25"} 1
registry_migration_tag_counts_bucket{import_type="pre",le="50"} 1
registry_migration_tag_counts_bucket{import_type="pre",le="100"} 1
registry_migration_tag_counts_bucket{import_type="pre",le="250"} 1
registry_migration_tag_counts_bucket{import_type="pre",le="500"} 1
registry_migration_tag_counts_bucket{import_type="pre",le="1000"} 1
registry_migration_tag_counts_bucket{import_type="pre",le="2000"} 1
registry_migration_tag_counts_bucket{import_type="pre",le="5000"} 1
registry_migration_tag_counts_bucket{import_type="pre",le="10000"} 1
registry_migration_tag_counts_bucket{import_type="pre",le="15000"} 1
registry_migration_tag_counts_bucket{import_type="pre",le="20000"} 1
registry_migration_tag_counts_bucket{import_type="pre",le="50000"} 1
registry_migration_tag_counts_bucket{import_type="pre",le="100000"} 1
registry_migration_tag_counts_bucket{import_type="pre",le="+Inf"} 1
registry_migration_tag_counts_sum{import_type="pre"} 10
registry_migration_tag_counts_count{import_type="pre"} 1
`)
	CountsFullName := fmt.Sprintf("%s_%s_%s", metrics.NamespacePrefix, subsystem, tagCountName)

	err := testutil.GatherAndCompare(prometheus.DefaultGatherer, &expected, CountsFullName)
	require.NoError(t, err)
}

func TestLayerCount(t *testing.T) {
	LayerCount(1)
	LayerCount(100)

	var expected bytes.Buffer
	expected.WriteString(`
# HELP registry_migration_layer_counts A histogram of layer counts per (pre)imported manifest.
# TYPE registry_migration_layer_counts histogram
registry_migration_layer_counts_bucket{le="1"} 1
registry_migration_layer_counts_bucket{le="2"} 1
registry_migration_layer_counts_bucket{le="5"} 1
registry_migration_layer_counts_bucket{le="10"} 1
registry_migration_layer_counts_bucket{le="25"} 1
registry_migration_layer_counts_bucket{le="50"} 1
registry_migration_layer_counts_bucket{le="100"} 2
registry_migration_layer_counts_bucket{le="200"} 2
registry_migration_layer_counts_bucket{le="+Inf"} 2
registry_migration_layer_counts_sum 101
registry_migration_layer_counts_count 2
`)
	CountFullName := fmt.Sprintf("%s_%s_%s", metrics.NamespacePrefix, subsystem, layerCountName)

	err := testutil.GatherAndCompare(prometheus.DefaultGatherer, &expected, CountFullName)
	require.NoError(t, err)
}

func TestBlobTransfer(t *testing.T) {
	BlobTransfer(.1, 34234123, BlobTypeLayer)
	BlobTransfer(0.25, 7754, BlobTypeConfig)
	BlobTransfer(0.4, 5351258, BlobTypeUnknown)

	var expected bytes.Buffer
	expected.WriteString(`
# HELP registry_migration_blob_transfer_duration_seconds A histogram of latencies for blob transfers.
# TYPE registry_migration_blob_transfer_duration_seconds histogram
registry_migration_blob_transfer_duration_seconds_bucket{blob_type="config",le="0.005"} 0
registry_migration_blob_transfer_duration_seconds_bucket{blob_type="config",le="0.01"} 0
registry_migration_blob_transfer_duration_seconds_bucket{blob_type="config",le="0.025"} 0
registry_migration_blob_transfer_duration_seconds_bucket{blob_type="config",le="0.05"} 0
registry_migration_blob_transfer_duration_seconds_bucket{blob_type="config",le="0.1"} 0
registry_migration_blob_transfer_duration_seconds_bucket{blob_type="config",le="0.25"} 1
registry_migration_blob_transfer_duration_seconds_bucket{blob_type="config",le="0.5"} 1
registry_migration_blob_transfer_duration_seconds_bucket{blob_type="config",le="1"} 1
registry_migration_blob_transfer_duration_seconds_bucket{blob_type="config",le="2.5"} 1
registry_migration_blob_transfer_duration_seconds_bucket{blob_type="config",le="5"} 1
registry_migration_blob_transfer_duration_seconds_bucket{blob_type="config",le="10"} 1
registry_migration_blob_transfer_duration_seconds_bucket{blob_type="config",le="+Inf"} 1
registry_migration_blob_transfer_duration_seconds_sum{blob_type="config"} 0.25
registry_migration_blob_transfer_duration_seconds_count{blob_type="config"} 1
registry_migration_blob_transfer_duration_seconds_bucket{blob_type="layer",le="0.005"} 0
registry_migration_blob_transfer_duration_seconds_bucket{blob_type="layer",le="0.01"} 0
registry_migration_blob_transfer_duration_seconds_bucket{blob_type="layer",le="0.025"} 0
registry_migration_blob_transfer_duration_seconds_bucket{blob_type="layer",le="0.05"} 0
registry_migration_blob_transfer_duration_seconds_bucket{blob_type="layer",le="0.1"} 1
registry_migration_blob_transfer_duration_seconds_bucket{blob_type="layer",le="0.25"} 1
registry_migration_blob_transfer_duration_seconds_bucket{blob_type="layer",le="0.5"} 1
registry_migration_blob_transfer_duration_seconds_bucket{blob_type="layer",le="1"} 1
registry_migration_blob_transfer_duration_seconds_bucket{blob_type="layer",le="2.5"} 1
registry_migration_blob_transfer_duration_seconds_bucket{blob_type="layer",le="5"} 1
registry_migration_blob_transfer_duration_seconds_bucket{blob_type="layer",le="10"} 1
registry_migration_blob_transfer_duration_seconds_bucket{blob_type="layer",le="+Inf"} 1
registry_migration_blob_transfer_duration_seconds_sum{blob_type="layer"} 0.1
registry_migration_blob_transfer_duration_seconds_count{blob_type="layer"} 1
registry_migration_blob_transfer_duration_seconds_bucket{blob_type="unknown",le="0.005"} 0
registry_migration_blob_transfer_duration_seconds_bucket{blob_type="unknown",le="0.01"} 0
registry_migration_blob_transfer_duration_seconds_bucket{blob_type="unknown",le="0.025"} 0
registry_migration_blob_transfer_duration_seconds_bucket{blob_type="unknown",le="0.05"} 0
registry_migration_blob_transfer_duration_seconds_bucket{blob_type="unknown",le="0.1"} 0
registry_migration_blob_transfer_duration_seconds_bucket{blob_type="unknown",le="0.25"} 0
registry_migration_blob_transfer_duration_seconds_bucket{blob_type="unknown",le="0.5"} 1
registry_migration_blob_transfer_duration_seconds_bucket{blob_type="unknown",le="1"} 1
registry_migration_blob_transfer_duration_seconds_bucket{blob_type="unknown",le="2.5"} 1
registry_migration_blob_transfer_duration_seconds_bucket{blob_type="unknown",le="5"} 1
registry_migration_blob_transfer_duration_seconds_bucket{blob_type="unknown",le="10"} 1
registry_migration_blob_transfer_duration_seconds_bucket{blob_type="unknown",le="+Inf"} 1
registry_migration_blob_transfer_duration_seconds_sum{blob_type="unknown"} 0.4
registry_migration_blob_transfer_duration_seconds_count{blob_type="unknown"} 1
# HELP registry_migration_blob_transfer_size_bytes A histogram of byte sizes for blob transfers.
# TYPE registry_migration_blob_transfer_size_bytes histogram
registry_migration_blob_transfer_size_bytes_bucket{blob_type="config",le="524288"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="config",le="1.048576e+06"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="config",le="6.7108864e+07"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="config",le="1.34217728e+08"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="config",le="2.68435456e+08"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="config",le="5.36870912e+08"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="config",le="1.073741824e+09"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="config",le="2.147483648e+09"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="config",le="3.221225472e+09"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="config",le="4.294967296e+09"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="config",le="5.36870912e+09"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="config",le="6.442450944e+09"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="config",le="7.516192768e+09"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="config",le="8.589934592e+09"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="config",le="9.663676416e+09"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="config",le="1.073741824e+10"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="config",le="2.147483648e+10"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="config",le="3.221225472e+10"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="config",le="4.294967296e+10"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="config",le="5.36870912e+10"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="config",le="+Inf"} 1
registry_migration_blob_transfer_size_bytes_sum{blob_type="config"} 7754
registry_migration_blob_transfer_size_bytes_count{blob_type="config"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="layer",le="524288"} 0
registry_migration_blob_transfer_size_bytes_bucket{blob_type="layer",le="1.048576e+06"} 0
registry_migration_blob_transfer_size_bytes_bucket{blob_type="layer",le="6.7108864e+07"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="layer",le="1.34217728e+08"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="layer",le="2.68435456e+08"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="layer",le="5.36870912e+08"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="layer",le="1.073741824e+09"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="layer",le="2.147483648e+09"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="layer",le="3.221225472e+09"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="layer",le="4.294967296e+09"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="layer",le="5.36870912e+09"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="layer",le="6.442450944e+09"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="layer",le="7.516192768e+09"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="layer",le="8.589934592e+09"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="layer",le="9.663676416e+09"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="layer",le="1.073741824e+10"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="layer",le="2.147483648e+10"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="layer",le="3.221225472e+10"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="layer",le="4.294967296e+10"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="layer",le="5.36870912e+10"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="layer",le="+Inf"} 1
registry_migration_blob_transfer_size_bytes_sum{blob_type="layer"} 3.4234123e+07
registry_migration_blob_transfer_size_bytes_count{blob_type="layer"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="unknown",le="524288"} 0
registry_migration_blob_transfer_size_bytes_bucket{blob_type="unknown",le="1.048576e+06"} 0
registry_migration_blob_transfer_size_bytes_bucket{blob_type="unknown",le="6.7108864e+07"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="unknown",le="1.34217728e+08"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="unknown",le="2.68435456e+08"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="unknown",le="5.36870912e+08"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="unknown",le="1.073741824e+09"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="unknown",le="2.147483648e+09"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="unknown",le="3.221225472e+09"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="unknown",le="4.294967296e+09"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="unknown",le="5.36870912e+09"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="unknown",le="6.442450944e+09"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="unknown",le="7.516192768e+09"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="unknown",le="8.589934592e+09"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="unknown",le="9.663676416e+09"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="unknown",le="1.073741824e+10"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="unknown",le="2.147483648e+10"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="unknown",le="3.221225472e+10"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="unknown",le="4.294967296e+10"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="unknown",le="5.36870912e+10"} 1
registry_migration_blob_transfer_size_bytes_bucket{blob_type="unknown",le="+Inf"} 1
registry_migration_blob_transfer_size_bytes_sum{blob_type="unknown"} 5.351258e+06
registry_migration_blob_transfer_size_bytes_count{blob_type="unknown"} 1
`)
	durationFullName := fmt.Sprintf("%s_%s_%s", metrics.NamespacePrefix, subsystem, blobTransferDurationName)
	sizeFullName := fmt.Sprintf("%s_%s_%s", metrics.NamespacePrefix, subsystem, blobTransferSizeName)

	err := testutil.GatherAndCompare(prometheus.DefaultGatherer, &expected, durationFullName, sizeFullName)
	require.NoError(t, err)
}

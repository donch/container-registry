package validation_test

import (
	"fmt"
	"math/rand"
	"regexp"
	"testing"

	"github.com/docker/distribution"
	"github.com/docker/distribution/context"
	"github.com/docker/distribution/manifest/ocischema"
	"github.com/docker/distribution/registry/storage/validation"
	"github.com/docker/distribution/testutil"
	"github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/require"
)

func TestVerifyManifest_OCI_NonDistributableLayer(t *testing.T) {
	ctx := context.Background()

	registry := createRegistry(t)
	repo := makeRepository(t, registry, "test")

	manifestService, err := testutil.MakeManifestService(repo)
	require.NoError(t, err)

	template := makeOCIManifestTemplate(t, repo)

	layer, err := repo.Blobs(ctx).Put(ctx, v1.MediaTypeImageLayer, nil)
	require.NoError(t, err)

	nonDistributableLayer := distribution.Descriptor{
		Digest:    digest.FromString("nonDistributableLayer"),
		Size:      6323,
		MediaType: v1.MediaTypeImageLayerNonDistributableGzip,
	}

	type testcase struct {
		BaseLayer distribution.Descriptor
		URLs      []string
		Err       error
	}

	cases := []testcase{
		{
			nonDistributableLayer,
			nil,
			distribution.ErrManifestBlobUnknown{Digest: nonDistributableLayer.Digest},
		},
		{
			layer,
			[]string{"http://foo/bar"},
			nil,
		},
		{
			nonDistributableLayer,
			[]string{"file:///local/file"},
			errInvalidURL,
		},
		{
			nonDistributableLayer,
			[]string{"http://foo/bar#baz"},
			errInvalidURL,
		},
		{
			nonDistributableLayer,
			[]string{""},
			errInvalidURL,
		},
		{
			nonDistributableLayer,
			[]string{"https://foo/bar", ""},
			errInvalidURL,
		},
		{
			nonDistributableLayer,
			[]string{"", "https://foo/bar"},
			errInvalidURL,
		},
		{
			nonDistributableLayer,
			[]string{"http://nope/bar"},
			errInvalidURL,
		},
		{
			nonDistributableLayer,
			[]string{"http://foo/nope"},
			errInvalidURL,
		},
		{
			nonDistributableLayer,
			[]string{"http://foo/bar"},
			nil,
		},
		{
			nonDistributableLayer,
			[]string{"https://foo/bar"},
			nil,
		},
	}

	for _, c := range cases {
		m := template
		l := c.BaseLayer
		l.URLs = c.URLs
		m.Layers = []distribution.Descriptor{l}
		dm, err := ocischema.FromStruct(m)
		if err != nil {
			t.Error(err)
			continue
		}

		v := validation.NewOCIValidator(
			manifestService,
			repo.Blobs(ctx),
			0,
			0,
			validation.ManifestURLs{
				Allow: regexp.MustCompile("^https?://foo"),
				Deny:  regexp.MustCompile("^https?://foo/nope"),
			})

		err = v.Validate(ctx, dm)
		if verr, ok := err.(distribution.ErrManifestVerification); ok {
			// Extract the first error
			if len(verr) == 2 {
				if _, ok = verr[1].(distribution.ErrManifestBlobUnknown); ok {
					err = verr[0]
				}
			} else if len(verr) == 1 {
				err = verr[0]
			}
		}
		require.Equal(t, c.Err, err)
	}
}

func TestVerifyManifest_OCI_InvalidSchemaVersion(t *testing.T) {
	ctx := context.Background()

	registry := createRegistry(t)
	repo := makeRepository(t, registry, "test")

	manifestService, err := testutil.MakeManifestService(repo)
	require.NoError(t, err)

	m := makeOCIManifestTemplate(t, repo)
	m.Versioned.SchemaVersion = 42

	dm, err := ocischema.FromStruct(m)
	require.NoError(t, err)

	v := validation.NewOCIValidator(manifestService, repo.Blobs(ctx), 0, 0, validation.ManifestURLs{})

	err = v.Validate(ctx, dm)
	require.EqualError(t, err, fmt.Sprintf("unrecognized manifest schema version %d", m.Versioned.SchemaVersion))
}

func TestVerifyManifest_OCI_ManifestLayer(t *testing.T) {
	ctx := context.Background()

	registry := createRegistry(t)
	repo := makeRepository(t, registry, "test")

	manifestService, err := testutil.MakeManifestService(repo)
	require.NoError(t, err)

	layer, err := repo.Blobs(ctx).Put(ctx, v1.MediaTypeImageLayer, nil)
	require.NoError(t, err)

	// Test a manifest used as a layer. Looking at the original oci validation
	// logic it appears that oci manifests allow manifests as layers. So we
	// should try to preserve this rather odd behavior.
	depManifest := makeOCIManifestTemplate(t, repo)
	depManifest.Layers = []distribution.Descriptor{layer}

	depM, err := ocischema.FromStruct(depManifest)
	require.NoError(t, err)

	mt, payload, err := depM.Payload()
	require.NoError(t, err)

	// If a manifest is used as a layer, it should have been pushed both as a
	// manifest as well as a blob.
	dgst, err := manifestService.Put(ctx, depM)
	require.NoError(t, err)

	_, err = repo.Blobs(ctx).Put(ctx, mt, payload)
	require.NoError(t, err)

	m := makeOCIManifestTemplate(t, repo)
	m.Layers = []distribution.Descriptor{{Digest: dgst, MediaType: mt}}

	dm, err := ocischema.FromStruct(m)
	require.NoError(t, err)

	v := validation.NewOCIValidator(manifestService, repo.Blobs(ctx), 0, 0, validation.ManifestURLs{})

	err = v.Validate(ctx, dm)
	require.NoErrorf(t, err, fmt.Sprintf("digest: %s", dgst))
}

func TestVerifyManifest_OCI_MultipleErrors(t *testing.T) {
	ctx := context.Background()

	registry := createRegistry(t)
	repo := makeRepository(t, registry, "test")

	manifestService, err := testutil.MakeManifestService(repo)
	require.NoError(t, err)

	layer, err := repo.Blobs(ctx).Put(ctx, v1.MediaTypeImageLayer, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a manifest with three layers, two of which are missing. We should
	// see the digest of each missing layer in the error message.
	m := makeOCIManifestTemplate(t, repo)
	m.Layers = []distribution.Descriptor{
		{Digest: digest.FromString("fake-blob-layer"), MediaType: v1.MediaTypeImageLayer},
		layer,
		{Digest: digest.FromString("fake-manifest-layer"), MediaType: v1.MediaTypeImageManifest},
	}

	dm, err := ocischema.FromStruct(m)
	require.NoError(t, err)

	v := validation.NewOCIValidator(manifestService, repo.Blobs(ctx), 0, 0, validation.ManifestURLs{})

	err = v.Validate(ctx, dm)
	require.Error(t, err)

	require.Contains(t, err.Error(), m.Layers[0].Digest.String())
	require.NotContains(t, err.Error(), m.Layers[1].Digest.String())
	require.Contains(t, err.Error(), m.Layers[2].Digest.String())
}

func TestVerifyManifest_OCI_ReferenceLimits(t *testing.T) {
	ctx := context.Background()

	registry := createRegistry(t)
	repo := makeRepository(t, registry, "test")

	manifestService, err := testutil.MakeManifestService(repo)
	require.NoError(t, err)

	var tests = []struct {
		name           string
		manifestLayers int
		refLimit       int
		wantErr        bool
	}{
		{
			name:           "no reference limit",
			manifestLayers: 10,
			refLimit:       0,
			wantErr:        false,
		},
		{
			name:           "reference limit greater than number of references",
			manifestLayers: 10,
			refLimit:       150,
			wantErr:        false,
		},
		{
			name:           "reference limit equal to number of references",
			manifestLayers: 9, // 9 layers + config = 10
			refLimit:       10,
			wantErr:        false,
		},
		{
			name:           "reference limit less than number of references",
			manifestLayers: 400,
			refLimit:       179,
			wantErr:        true,
		},
		{
			name:           "negative reference limit",
			manifestLayers: 8,
			refLimit:       -17,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := makeOCIManifestTemplate(t, repo)

			// Create a random layer for each of the specified manifest layers.
			for i := 0; i < tt.manifestLayers; i++ {
				b := make([]byte, rand.Intn(20))
				rand.Read(b)

				layer, err := repo.Blobs(ctx).Put(ctx, v1.MediaTypeImageLayer, b)
				require.NoError(t, err)

				m.Layers = append(m.Layers, layer)
			}

			dm, err := ocischema.FromStruct(m)
			require.NoError(t, err)

			v := validation.NewOCIValidator(manifestService, repo.Blobs(ctx), tt.refLimit, 0, validation.ManifestURLs{})

			err = v.Validate(ctx, dm)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestVerifyManifest_OCI_PayloadLimits(t *testing.T) {
	ctx := context.Background()

	registry := createRegistry(t)
	repo := makeRepository(t, registry, "test")

	manifestService, err := testutil.MakeManifestService(repo)
	require.NoError(t, err)

	m := makeOCIManifestTemplate(t, repo)

	dm, err := ocischema.FromStruct(m)
	require.NoError(t, err)

	_, payload, err := dm.Payload()
	require.NoError(t, err)

	baseOCIManifestIndextSize := len(payload)

	tests := map[string]struct {
		payloadLimit int
		wantErr      bool
		expectedErr  error
	}{
		"no payload size limit": {
			payloadLimit: 0,
			wantErr:      false,
		},
		"payload size limit greater than manifest size": {
			payloadLimit: baseOCIManifestIndextSize * 2,
			wantErr:      false,
		},
		"payload size limit equal to manifest size": {
			payloadLimit: baseOCIManifestIndextSize,
			wantErr:      false,
		},
		"payload size limit less than manifest size": {
			payloadLimit: baseOCIManifestIndextSize / 2,
			wantErr:      true,
			expectedErr: distribution.ErrManifestVerification{
				distribution.ErrManifestPayloadSizeExceedsLimit{PayloadSize: baseOCIManifestIndextSize, Limit: baseOCIManifestIndextSize / 2},
			},
		},
		"negative payload size limit": {
			payloadLimit: -baseOCIManifestIndextSize,
			wantErr:      false,
		},
	}

	for tn, tt := range tests {
		t.Run(tn, func(t *testing.T) {
			v := validation.NewOCIValidator(manifestService, repo.Blobs(ctx), 0, tt.payloadLimit, validation.ManifestURLs{})

			err = v.Validate(ctx, dm)
			if tt.wantErr {
				require.Error(t, err)
				require.EqualError(t, err, tt.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

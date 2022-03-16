# Bad Manifest Layer Format

This test fixture simulates a registry with a repository in which a manifest has a layer with an unknown media type.
This can be used to cause errors in order to test the error path. A layer with an unknown media type should cause the
respective manifest and repository import to fail.

## Fixture Creation

This fixture was created by copying the `a-simple` repository and relevant blobs from the `happy-path` text fixture and
changing the manifest referenced layers media type to `application/foo.bar.layer.v1.tar+gzip`.

# Bad Manifest Format

This test fixture simulates a registry with a repository in which a manifest has an unknown media type.
This can be used to cause errors in order to test the error path. A manifest with an unknown media type should cause the
respective manifest and repository import to fail.

## Fixture Creation

This fixture was created by copying the `a-simple` repository and relevant blobs from the `happy-path` text fixture and
changing the manifest media type to `application/foo.bar.manfiest.v1.tar+gzip`.

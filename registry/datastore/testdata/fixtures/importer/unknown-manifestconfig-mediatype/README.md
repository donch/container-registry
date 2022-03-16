# Bad Manifest Format

This test fixture simulates a registry with a repository in which a manifest has configuration with an unknown media type.
This can be used to cause errors in order to test the error path. A manifest's configuration that has an unknown media
type should cause the respective manifest and repository import to fail.

## Fixture Creation

This fixture was created by copying the `a-simple` repository and relevant blobs from the `happy-path` text fixture and
changing the manifest configuration media type to `application/foo.bar.container.image.v1+json`.

# Missing Revisions

This test fixture simulates a registry with a repository in which the tags
path is missing. Malformed registries such as these can appear when an
administrator performs manual maintenance against filesystem metadata.

## Fixture Creation

This fixture was created by uploading three schema 2 images and removing the
`_manifests/tags/` path in one of them.

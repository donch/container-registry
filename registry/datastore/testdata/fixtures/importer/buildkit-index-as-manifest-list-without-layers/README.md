# Buildkit Index With No Layers

This test fixture simulates a registry with a repository in which an OCI image index (manifest list)
is missing layers, as seen in
[this file](./docker/registry/v2/blobs/sha256/5d/5d42ad55895dbd1a4a57d306214c250668f70ba85045a8f432c3806362780621/data).
These manifests are not useful, but they were allowed to be pushed to the registry at some point,
so we should be able to migrate these manifest lists too.
See https://gitlab.com/gitlab-org/container-registry/-/issues/695.

## Fixture Creation

This fixture copied from [../buildkit-index-as-manifest-list](../buildkit-index-as-manifest-list) and
removing the layers from the configuration file, as well as removing the unnecessary blobs manually.

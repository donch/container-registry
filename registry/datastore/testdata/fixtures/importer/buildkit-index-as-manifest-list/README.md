# Bad Manifest Format

This test fixture simulates a registry with a repository in which an OCI image index (manifest list)
contains layers and not just other manifests, as seen in 
[this file](./docker/registry/v2/blobs/sha256/5d/5d42ad55895dbd1a4a57d306214c250668f70ba85045a8f432c3806362780621/data).
This index can be created by `docker buildx` and pushed  to a container registry.

This issue was addressed  in https://gitlab.com/gitlab-org/container-registry/-/issues/407 
for the import handler and the solution is adapted now for the importer
https://gitlab.com/gitlab-org/container-registry/-/issues/414.


## Fixture Creation

This fixture was created by running the following command

```shell
docker buildx build \                                                                                                                                                                                                                       
  --cache-from=type=registry,ref=127.0.0.1:5000/buildx:cache \
  --cache-to=type=registry,ref=127.0.0.1:5000/buildx:cache,mode=max \
  --platform linux/amd64,linux/arm64,linux/arm/v7 \
  -t 127.0.0.1:5000/buildx:image \
  --push \
  - <<<$(echo -e "FROM alpine:3.14.0\nRUN echo 'foo' > /foo")
```

Where `127.0.0.1` is a local instance of a container registry without the metadata database
enabled.

We have then deleted all blobs and manifests except
sha256:5d42ad55895dbd1a4a57d306214c250668f70ba85045a8f432c3806362780621 (the actual buildkit cache image) and its
references.

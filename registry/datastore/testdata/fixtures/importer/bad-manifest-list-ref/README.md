# Bad Manifest List Reference

This test fixture simulates a registry with a repository containing 2 manifests and 1 manifest list referencing them.

To simulate a bad manifest list reference we simply set the content of one of the referenced manifests at
`docker/registry/v2/blobs/sha256/59/597bd5c319cc09d6bb295b4ef23cac50ec7c373fff5fe923cfd246ec09967b31/data` to an JSON
that does not represent a valid manifest payload.
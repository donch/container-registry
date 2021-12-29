# Unlinked Layers

This test fixture simulates a registry with a repository in which some layers
are unlinked from the repository containing the manifest. The
[Delete Blob API](https://gitlab.com/gitlab-org/container-registry/-/blob/master/docs/spec/api.md#delete-blob)
allows users to unlink specific blobs from the repository, preventing manifests
referencing those blobs from being pulled.

## Fixture Creation

This fixure was created by uploading a schema 2 image and removing three layers
of the manifest using the following command:
```
curl -v -X DELETE  localhost:5000/v2/a-unlinked-layers/blobs/<digest>
```

This results in a manifest in a repository with the following layers still
linked to the repository:

- `8fc4b4a7d078aa4e5764f60fd80ab1d0000fb88560909622862f57833cd3e504`
- `0fe9c8bfd649945d25a2744cfd6e59aa4af191f55c4282a202afc880ec130155`
- `ca11fc3041d580ea5eaeb17396c8c806670b4e26b5d0bca88bf337fc836f9a33`
- `189d3ff46c7870d6fb194cbf5581145302a33305446aa65989072ce1c753a36e`

While these layers are unlinked from the repository containing the manifest:

- `c22b5cc4704d517e05e9570a477393c47695ff1702e3c3490be13b0989c28d4a`
- `fade238920e4b7708c097768696f25a1ae1c55458ef3c547d979d0b0fec03940`
- `c9b1b535fdd91a9855fb7f82348177e5f019329a58c53c47272962dd60f71fc9`

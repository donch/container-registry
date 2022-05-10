# Empty Layer Links

This test fixture simulates a registry with a repository in with three images.
One image is expected to import without issue. The second image has a layer
whose link file within the repository contains no content. The third image's
configuration blob's layer link file also contains no content.

## Fixture Creation

This fixure was created by uploading three schema 2 images and truncating the
layer link files for a single filesystem layer of the second image, and the
manifest configuration blob of the third image.

```
truncate -s0 broken-layer-links/_layers/sha256/c3cce1b00b48329400b61ad073da43e8f838d8fcd134b86ac05c7b7a0452992c/link
truncate -s0 broken-layer-links/_layers/sha256/4d59f0a788804aa0f2d4fee2469704767a455e72f0bb9ee9fcc72407692812e7/link
```


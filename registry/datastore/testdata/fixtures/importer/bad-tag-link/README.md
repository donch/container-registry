# Bad Tag Link

This test fixture simulates a registry with a repository containing 3 tags, one of which (3.11.6) is broken.

To simulate a broken tag link we simply erased the content of
`docker/registry/v2/repositories/alpine/_manifests/tags/3.11.6/current/link`.

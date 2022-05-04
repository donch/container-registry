# Bad Tag Link

This test fixture simulates a registry with a repository containing 3 tags, one of which (3.11.6) has a missing link.

To simulate a missing tag link we simply deleted 
`docker/registry/v2/repositories/alpine/_manifests/tags/3.11.6/current/link`.

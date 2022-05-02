# Bad Manifest Link

This test fixture simulates a registry with a repository containing 3 manifests, one of which 
(sha256:39eda93d15866957feaee28f8fc5adb545276a64147445c64992ef69804dbf01) has a broken link.

To simulate a broken manifest link we simply erased the content of
`docker/registry/v2/repositories/alpine/_manifests/revisions/sha256/39eda93d15866957feaee28f8fc5adb545276a64147445c64992ef69804dbf01/link`.

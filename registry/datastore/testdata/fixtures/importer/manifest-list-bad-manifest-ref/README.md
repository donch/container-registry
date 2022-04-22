# Bad Manifest Reference in List

This test fixture simulates a registry with a repository containing 1 manifest and 1 manifest list referencing it.
Then, the referenced manifest has an invalid layer reference.

To simulate an invalid layer reference we simply altered one of the manifest layers media type to
`application/foo.bar.layer.v1.tar+gzip`.
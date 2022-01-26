# Bad Manifest Format

This test fixture simulates a registry with a repository in which a manifest is
in an invalid format. This can be used to cause errors in order to test the
importer error path.

## Fixture Creation

This fixure was created by copying the a-simple repository and relevant blobs
from the happy path text fixure and converting the manifest JSON into YAML.

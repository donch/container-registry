# Last Tag Error

This test fixture simulates a registry with the a final tag link that will not
import. This is meant to be to test that the last tag error is handled correctly.
Previous versions would occasionally ingore the final tag error on final import.

The test repository only contains a single tag link. Since the tests use the
local file system, it is not possible to consistently replicate this error with
multiple tags.

## Fixture Creation

This fixure was created by pushing a single tagged image to the registry. The
tests handle simulating the error, so no modifications are made to the fixture.

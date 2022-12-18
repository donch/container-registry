package s3

import (
	"errors"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/cenkalti/backoff/v4"
	"github.com/stretchr/testify/require"
)

var (
	aPermanentAWSRequestError = awserr.NewRequestFailure(
		awserr.New("test-code", "test-error-message", errors.New("test-original error")),
		http.StatusFailedDependency, "testReqID")

	aNonPermanentAWSRequestError = awserr.NewRequestFailure(
		awserr.New("test-code", "test-error-message", errors.New("test-original error")),
		http.StatusInternalServerError, "testReqID")

	anAWSInvalidPresignExpireError = awserr.New(request.ErrCodeInvalidPresignExpire, "test-error-message", errors.New("test-original error"))

	aRequestAWSSerializationError = awserr.NewRequestFailure(
		awserr.New(request.ErrCodeSerialization,
			"failed to decode REST XML response", errors.New("test-original error")),
		http.StatusAccepted,
		"1234",
	)

	anAWSSerializationError = awserr.New(request.ErrCodeSerialization, "test-error-message", errors.New("test-original error"))
)

func TestWrapAWSerr(t *testing.T) {
	tests := []struct {
		name          string
		inputError    error
		expectedError error
	}{
		{
			name:          "no error",
			inputError:    nil,
			expectedError: nil,
		},
		{
			name:          "aws server request failure",
			inputError:    aPermanentAWSRequestError,
			expectedError: backoff.Permanent(aPermanentAWSRequestError),
		},
		{
			name:          "aws non-server request failure",
			inputError:    aNonPermanentAWSRequestError,
			expectedError: aNonPermanentAWSRequestError,
		},
		{
			name:          "aws InvalidPresignExpireError (not-request-specific)",
			inputError:    anAWSInvalidPresignExpireError,
			expectedError: backoff.Permanent(anAWSInvalidPresignExpireError),
		},
		{
			name:          "aws SerializationError error (not-request-specific)",
			inputError:    anAWSSerializationError,
			expectedError: anAWSSerializationError,
		},
		{
			name:          "aws SerializationError error (request-specific)",
			inputError:    aRequestAWSSerializationError,
			expectedError: aRequestAWSSerializationError,
		},
	}
	t.Logf("Running %s test", t.Name())
	t.Parallel()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.ErrorIs(t, test.expectedError, wrapAWSerr(test.inputError))
		})
	}
}

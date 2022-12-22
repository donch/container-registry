package s3

import (
	"errors"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/cenkalti/backoff/v4"
	"golang.org/x/time/rate"
)

type backoffConstructor func() backoff.BackOff

func withNoExponentialBackoff() backoff.BackOff {
	return &noBackoff{}
}

// noBackoff disables exponential backoffs.
type noBackoff struct{}

// NextBackOff always returns backoff.Stop to signal the caller not to retry the operation.
func (n *noBackoff) NextBackOff() time.Duration {
	return backoff.Stop
}

// Reset to initial state.
func (n *noBackoff) Reset() {}

func withExponentialBackoff(max int64) wrapperOpt {
	if max < 0 {
		max = 0
	}

	b := backoff.NewExponentialBackOff()
	b.InitialInterval = defaultInitialInterval
	b.RandomizationFactor = defaultRandomizationFactor
	b.Multiplier = defaultMultiplier
	b.MaxInterval = defaultMaxInterval
	b.MaxElapsedTime = defaultMaxElapsedTime

	return func(w *s3wrapper) {
		w.backoff = func() backoff.BackOff {
			return backoff.WithMaxRetries(b, uint64(max))
		}
	}
}

// s3wrapper implements a subset of s3iface.S3API allowing us to rate limit,
// retry, add trace logging, or otherwise improve s3 calls made by the driver.
type s3wrapper struct {
	s3 s3iface.S3API
	*rate.Limiter
	backoff backoffConstructor
	notify  backoff.Notify
}

type wrapperOpt func(*s3wrapper)

func withRateLimit(max int64, burst int) wrapperOpt {
	return func(w *s3wrapper) {
		w.Limiter = rate.NewLimiter(rate.Limit(max), burst)
	}
}

func withBackoffNotify(n backoff.Notify) wrapperOpt {
	return func(w *s3wrapper) {
		w.notify = n
	}
}

func newS3Wrapper(s3 s3iface.S3API, opts ...wrapperOpt) *s3wrapper {
	w := &s3wrapper{
		s3:      s3,
		Limiter: rate.NewLimiter(rate.Inf, 0),
		backoff: withNoExponentialBackoff,
	}

	for _, o := range opts {
		o(w)
	}

	return w
}

func (w *s3wrapper) PutObjectWithContext(ctx aws.Context, input *s3.PutObjectInput, opts ...request.Option) (*s3.PutObjectOutput, error) {
	var out *s3.PutObjectOutput

	err := w.waitRetryNotify(ctx, func() error {
		var err error
		out, err = w.s3.PutObjectWithContext(ctx, input, opts...)
		return err
	})

	return out, err
}

func (w *s3wrapper) GetObjectWithContext(ctx aws.Context, input *s3.GetObjectInput, opts ...request.Option) (*s3.GetObjectOutput, error) {
	var out *s3.GetObjectOutput

	err := w.waitRetryNotify(ctx, func() error {
		var err error
		out, err = w.s3.GetObjectWithContext(ctx, input, opts...)
		return err
	})

	return out, err
}

func (w *s3wrapper) CreateMultipartUploadWithContext(ctx aws.Context, input *s3.CreateMultipartUploadInput, opts ...request.Option) (*s3.CreateMultipartUploadOutput, error) {
	var out *s3.CreateMultipartUploadOutput

	err := w.waitRetryNotify(ctx, func() error {
		var err error
		out, err = w.s3.CreateMultipartUploadWithContext(ctx, input, opts...)
		return err
	})

	return out, err
}

func (w *s3wrapper) ListMultipartUploadsWithContext(ctx aws.Context, input *s3.ListMultipartUploadsInput, opts ...request.Option) (*s3.ListMultipartUploadsOutput, error) {
	var out *s3.ListMultipartUploadsOutput

	err := w.waitRetryNotify(ctx, func() error {
		var err error
		out, err = w.s3.ListMultipartUploadsWithContext(ctx, input, opts...)
		return err
	})

	return out, err
}

func (w *s3wrapper) ListPartsWithContext(ctx aws.Context, input *s3.ListPartsInput, opts ...request.Option) (*s3.ListPartsOutput, error) {
	var out *s3.ListPartsOutput

	err := w.waitRetryNotify(ctx, func() error {
		var err error
		out, err = w.s3.ListPartsWithContext(ctx, input, opts...)
		return err
	})

	return out, err
}

func (w *s3wrapper) ListObjectsV2WithContext(ctx aws.Context, input *s3.ListObjectsV2Input, opts ...request.Option) (*s3.ListObjectsV2Output, error) {
	var out *s3.ListObjectsV2Output

	err := w.waitRetryNotify(ctx, func() error {
		var err error
		out, err = w.s3.ListObjectsV2WithContext(ctx, input, opts...)
		return err
	})

	return out, err
}

func (w *s3wrapper) CopyObjectWithContext(ctx aws.Context, input *s3.CopyObjectInput, opts ...request.Option) (*s3.CopyObjectOutput, error) {
	var out *s3.CopyObjectOutput

	err := w.waitRetryNotify(ctx, func() error {
		var err error
		out, err = w.s3.CopyObjectWithContext(ctx, input, opts...)
		return err
	})

	return out, err
}

func (w *s3wrapper) UploadPartCopyWithContext(ctx aws.Context, input *s3.UploadPartCopyInput, opts ...request.Option) (*s3.UploadPartCopyOutput, error) {
	var out *s3.UploadPartCopyOutput

	err := w.waitRetryNotify(ctx, func() error {
		var err error
		out, err = w.s3.UploadPartCopyWithContext(ctx, input, opts...)
		return err
	})

	return out, err
}

func (w *s3wrapper) CompleteMultipartUploadWithContext(ctx aws.Context, input *s3.CompleteMultipartUploadInput, opts ...request.Option) (*s3.CompleteMultipartUploadOutput, error) {
	var out *s3.CompleteMultipartUploadOutput

	err := w.waitRetryNotify(ctx, func() error {
		var err error
		out, err = w.s3.CompleteMultipartUploadWithContext(ctx, input, opts...)
		return err
	})

	return out, err
}

func (w *s3wrapper) DeleteObjectsWithContext(ctx aws.Context, input *s3.DeleteObjectsInput, opts ...request.Option) (*s3.DeleteObjectsOutput, error) {
	var out *s3.DeleteObjectsOutput

	err := w.waitRetryNotify(ctx, func() error {
		var err error
		out, err = w.s3.DeleteObjectsWithContext(ctx, input, opts...)
		return err
	})

	return out, err
}

func (w *s3wrapper) GetObjectRequest(input *s3.GetObjectInput) (*request.Request, *s3.GetObjectOutput) {
	// This does not make network calls, no need to rate limit.
	return w.s3.GetObjectRequest(input)
}

func (w *s3wrapper) HeadObjectRequest(input *s3.HeadObjectInput) (*request.Request, *s3.HeadObjectOutput) {
	// This does not make network calls, no need to rate limit.
	return w.s3.HeadObjectRequest(input)
}

func (w *s3wrapper) ListObjectsV2PagesWithContext(ctx aws.Context, input *s3.ListObjectsV2Input, f func(*s3.ListObjectsV2Output, bool) bool, opts ...request.Option) error {
	return w.waitRetryNotify(ctx, func() error {
		return w.s3.ListObjectsV2PagesWithContext(ctx, input, f, opts...)
	})
}

func (w *s3wrapper) AbortMultipartUploadWithContext(ctx aws.Context, input *s3.AbortMultipartUploadInput, opts ...request.Option) (*s3.AbortMultipartUploadOutput, error) {
	var out *s3.AbortMultipartUploadOutput

	err := w.waitRetryNotify(ctx, func() error {
		var err error
		out, err = w.s3.AbortMultipartUploadWithContext(ctx, input, opts...)
		return err
	})

	return out, err
}

func (w *s3wrapper) UploadPartWithContext(ctx aws.Context, input *s3.UploadPartInput, opts ...request.Option) (*s3.UploadPartOutput, error) {
	var out *s3.UploadPartOutput

	err := w.waitRetryNotify(ctx, func() error {
		var err error
		out, err = w.s3.UploadPartWithContext(ctx, input, opts...)
		return err
	})

	return out, err
}

func (w *s3wrapper) waitRetryNotify(ctx aws.Context, f backoff.Operation) error {
	err := backoff.RetryNotify(func() error {
		if err := w.Wait(ctx); err != nil {
			return backoff.Permanent(err)
		}

		awsErr := f()
		return wrapAWSerr(awsErr)
	},
		w.backoff(),
		w.notify,
	)

	return err
}

// wrapAWSerr wraps the original error with backoff.Permanent if the error
// should not be retried.
func wrapAWSerr(e error) error {
	if e == nil {
		return nil
	}

	// Retry any request failures that are server errors.
	var reqErr awserr.RequestFailure
	if errors.As(e, &reqErr) {
		if reqErr.StatusCode() != http.StatusTooManyRequests && reqErr.StatusCode() < http.StatusInternalServerError &&
			reqErr.Code() != request.ErrCodeSerialization {
			return backoff.Permanent(e)
		}

		return e
	}

	// Some retryable errors are not specifically awserr.RequestFailure, continue
	// evaluating errors to see if we can retry.

	// Don't attempt to backoff from errors that are known to be client errors.
	var awsErr awserr.Error
	if errors.As(e, &awsErr) {
		if awsErr.Code() == request.ErrCodeInvalidPresignExpire {
			return backoff.Permanent(e)
		}
	}

	return e
}

package context

import "context"

const (
	// CDNRedirectKey is used to get the Google Cloud CDN redirection flag from a context.
	CDNRedirectKey = "cdn.redirect"
)

type cdnRedirectContext struct {
	context.Context
	redirect bool
}

// Value implements context.Context.
func (c cdnRedirectContext) Value(key interface{}) interface{} {
	switch key {
	case CDNRedirectKey:
		return c.redirect
	default:
		return c.Context.Value(key)
	}
}

// WithCDNRedirect returns a context with the Google Cloud CDN redirection flag enabled. This is a temporary mechanism
// to flag if we should redirect blob HEAD or GET requests to Google Cloud CDN during the percentage rollout of this
// feature for GitLab.com. See https://gitlab.com/gitlab-org/gitlab/-/issues/349417 for more details.
func WithCDNRedirect(ctx context.Context) context.Context {
	return cdnRedirectContext{
		Context:  ctx,
		redirect: true,
	}
}

// ShouldRedirectToCDN determines whether a request context was marked as eligible for Google Cloud CDN redirection.
func ShouldRedirectToCDN(ctx context.Context) bool {
	if v, ok := ctx.Value(CDNRedirectKey).(bool); ok {
		return v
	}
	return false
}

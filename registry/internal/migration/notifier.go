package migration

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/docker/distribution/log"
	"gitlab.com/gitlab-org/labkit/correlation"
)

const (
	NotifierClientName = "gitlab-container-registry"
)

var (
	errMissingURL       = errors.New("missing URL for import notifier")
	errMissingAPISecret = errors.New("missing API secret for import notifier")

	pathPlaceholder = url.QueryEscape("{path}")
)

// Notifier holds the configuration needed to send an HTTP request
// to the specified endpoint using the secret in the Authorization header.
type Notifier struct {
	endpoint string
	secret   string
	client   *http.Client
}

// Notification defines the fields that will be sent by the Notifier in
// the request body
type Notification struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Status string `json:"status"`
	Detail string `json:"detail"`
}

// NewNotifier creates an instance of the Notifier with a given configuration.
// It returns an error if it cannot parse the endpoint into a valid URL, or
// if the secret is empty.
func NewNotifier(endpoint, secret string, timeout time.Duration) (*Notifier, error) {
	if endpoint == "" {
		return nil, errMissingURL
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("parsing endpoint: %w", err)
	}

	if secret == "" {
		return nil, errMissingAPISecret
	}

	return &Notifier{
		endpoint: u.String(),
		secret:   secret,
		client: &http.Client{
			Transport: correlation.NewInstrumentedRoundTripper(http.DefaultTransport, correlation.WithClientName(NotifierClientName)),
			Timeout:   timeout,
		},
	}, nil
}

// insertPathInEndpoint will try to find the keyword `{path}` in the configured endpoint using a regular expression.
// If it does, the `{path}` will be replaced with the passed variable path.
// Otherwise, the string is returned as is.
func (n *Notifier) insertPathInEndpoint(path string) string {
	return strings.Replace(n.endpoint, pathPlaceholder, path, -1)
}

// Notify sends an HTTP request to the configured endpoint containing the specified body
func (n *Notifier) Notify(ctx context.Context, notification *Notification) error {
	l := log.GetLogger(log.WithContext(ctx)).
		WithFields(log.Fields{
			"name":   notification.Name,
			"path":   notification.Path,
			"status": notification.Status,
			"detail": notification.Detail,
		})

	l.Info("sending import notification")

	b, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("marshalling notification %w", err)
	}

	reqURLWithPath := n.insertPathInEndpoint(notification.Path)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, reqURLWithPath, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("creating notification request %w", err)
	}

	req.Header.Set("Authorization", n.secret)
	req.Header.Set("Content-Type", "application/json")

	res, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("making request %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		err := fmt.Errorf("import notifier received response: %d", res.StatusCode)
		l.WithError(err).Error("unexpected response sending import notification")
		return err
	}

	l.Info("sent import notification successfully")

	return nil
}

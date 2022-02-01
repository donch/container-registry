package migration

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/docker/distribution/log"
)

var (
	errMissingURL       = errors.New("missing URL for import notifier")
	errMissingAPISecret = errors.New("missing API secret for import notifier")
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
			Timeout: timeout,
		},
	}, nil
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, n.endpoint, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("creating notification request %w", err)
	}

	req.Header.Set("Authorization", n.secret)

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

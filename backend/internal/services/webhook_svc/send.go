package webhook_svc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	// requestTimeout bounds one call to the receiver.
	requestTimeout = 10 * time.Second
	// maxErrorBody is how much of an error response is kept for the status.
	maxErrorBody = 200
)

// header is one stored request header.
type header struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// payload is the JSON body sent for POST, PUT and PATCH.
type payload struct {
	Event          string `json:"event"`
	PackID         uint   `json:"packId"`
	Slug           string `json:"slug"`
	Name           string `json:"name"`
	HeadSnapshotID *uint  `json:"headSnapshotId"`
	At             string `json:"at"`
}

// outcome is the result of one call.
type outcome struct {
	StatusCode int
	// Err is empty on success.
	Err string
	// Retry reports that a later attempt may succeed (network error, 429, 5xx).
	Retry bool
}

func (o outcome) ok() bool { return o.Err == "" }

// newHTTPClient builds the client used for webhook calls. The receiver is
// chosen by a pack owner and is typically on the private network, so private
// addresses are allowed; redirects are not followed (a 3xx is a failure), and
// environment proxies are ignored.
func newHTTPClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	return &http.Client{
		Timeout:   requestTimeout,
		Transport: transport,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

// buildRequest assembles the request for a webhook call.
func buildRequest(ctx context.Context, method, url string, headers []header, body payload) (*http.Request, error) {
	var reader io.Reader
	if method != http.MethodGet {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("encode webhook payload: %w", err)
		}
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return nil, fmt.Errorf("build webhook request: %w", err)
	}
	req.Header.Set("User-Agent", "packwiz-web")
	if reader != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for _, h := range headers {
		req.Header.Set(h.Name, h.Value)
	}
	return req, nil
}

// classify turns a response status into an outcome.
func classify(status int, body string) outcome {
	switch {
	case status >= 200 && status < 300:
		return outcome{StatusCode: status}
	case status == http.StatusTooManyRequests || status >= 500:
		return outcome{StatusCode: status, Err: statusMessage(status, body), Retry: true}
	default:
		return outcome{StatusCode: status, Err: statusMessage(status, body)}
	}
}

func statusMessage(status int, body string) string {
	if len(body) > maxErrorBody {
		body = body[:maxErrorBody]
	}
	if body == "" {
		return fmt.Sprintf("HTTP %d", status)
	}
	return fmt.Sprintf("HTTP %d: %s", status, body)
}

// send performs one call.
func send(ctx context.Context, client *http.Client, method, url string, headers []header, body payload) outcome {
	req, err := buildRequest(ctx, method, url, headers, body)
	if err != nil {
		return outcome{Err: err.Error()}
	}
	resp, err := client.Do(req)
	if err != nil {
		// transport errors can carry the URL; it is not secret (no credentials
		// are allowed in it)
		return outcome{Err: err.Error(), Retry: true}
	}
	defer resp.Body.Close()
	snippet, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody)) // best effort, only for the status text
	return classify(resp.StatusCode, string(snippet))
}

package dto

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	maxWebhookHeaders     = 20
	maxWebhookHeaderValue = 4096
)

// webhookMethods are the HTTP methods a webhook may use.
var webhookMethods = map[string]struct{}{
	http.MethodPost: {}, http.MethodPut: {}, http.MethodPatch: {}, http.MethodGet: {},
}

// forbiddenWebhookHeaders are set by the HTTP client itself.
var forbiddenWebhookHeaders = map[string]struct{}{
	"host": {}, "content-length": {}, "transfer-encoding": {}, "connection": {},
}

var webhookHeaderName = regexp.MustCompile("^[!#$%&'*+.^_`|~0-9A-Za-z-]{1,128}$")

// ValidateWebhookURL checks the URL is an absolute http(s) URL without
// embedded credentials (credentials belong in headers).
func ValidateWebhookURL(raw string) error {
	if len(raw) > 2048 {
		return errors.New("webhook URL is too long")
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return errors.New("webhook URL must be an absolute URL")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("webhook URL must use http or https")
	}
	if u.User != nil {
		return errors.New("webhook URL must not contain credentials; use a header")
	}
	return nil
}

// WebhookHeaderForm is one request header. On update a blank Value keeps the
// stored value of the header with the same name.
type WebhookHeaderForm struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// PackWebhookForm is the save payload of a pack webhook. Headers replace the
// stored set; header values are write-only.
type PackWebhookForm struct {
	Enabled bool                `json:"enabled"`
	URL     string              `json:"url"`
	Method  string              `json:"method"`
	Headers []WebhookHeaderForm `json:"headers"`
}

func (f *PackWebhookForm) Validate() error {
	f.URL = strings.TrimSpace(f.URL)
	f.Method = strings.ToUpper(strings.TrimSpace(f.Method))
	if f.Method == "" {
		f.Method = http.MethodPost
	}
	if _, ok := webhookMethods[f.Method]; !ok {
		return errors.New("method must be POST, PUT, PATCH or GET")
	}
	// a disabled webhook may be saved half filled in
	if f.URL != "" || f.Enabled {
		if err := ValidateWebhookURL(f.URL); err != nil {
			return err
		}
	}
	if len(f.Headers) > maxWebhookHeaders {
		return fmt.Errorf("at most %d headers are allowed", maxWebhookHeaders)
	}
	seen := make(map[string]struct{}, len(f.Headers))
	for i := range f.Headers {
		h := &f.Headers[i]
		h.Name = strings.TrimSpace(h.Name)
		if !webhookHeaderName.MatchString(h.Name) {
			return fmt.Errorf("invalid header name %q", h.Name)
		}
		key := strings.ToLower(h.Name)
		if _, bad := forbiddenWebhookHeaders[key]; bad {
			return fmt.Errorf("header %q cannot be set", h.Name)
		}
		if _, dup := seen[key]; dup {
			return fmt.Errorf("duplicate header %q", h.Name)
		}
		seen[key] = struct{}{}
		if len(h.Value) > maxWebhookHeaderValue || strings.ContainsAny(h.Value, "\r\n") {
			return fmt.Errorf("invalid value for header %q", h.Name)
		}
	}
	return nil
}

// WebhookHeaderResponse names a stored header. Values are never returned.
type WebhookHeaderResponse struct {
	Name string `json:"name"`
}

// PackWebhookResponse is the view of a pack webhook.
type PackWebhookResponse struct {
	Enabled bool                    `json:"enabled"`
	URL     string                  `json:"url"`
	Method  string                  `json:"method"`
	Headers []WebhookHeaderResponse `json:"headers"`
	// Broken is set when the stored headers can no longer be decrypted (the
	// session secret changed); they must be entered again.
	Broken         bool       `json:"broken"`
	LastStatusCode int        `json:"lastStatusCode"`
	LastError      string     `json:"lastError"`
	LastAttemptAt  *time.Time `json:"lastAttemptAt"`
	LastSuccessAt  *time.Time `json:"lastSuccessAt"`
}

// PackWebhookTestResponse is the outcome of a test delivery.
type PackWebhookTestResponse struct {
	OK         bool   `json:"ok"`
	StatusCode int    `json:"statusCode"`
	Error      string `json:"error"`
}

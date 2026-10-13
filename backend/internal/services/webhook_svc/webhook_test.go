package webhook_svc

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"packwiz-web/internal/types/dto"
)

func TestMergeHeaders(t *testing.T) {
	existing := []header{{Name: "Authorization", Value: "Bearer old"}}

	got, err := mergeHeaders(existing, []dto.WebhookHeaderForm{
		{Name: "authorization"},
		{Name: "X-New", Value: "v"},
	})
	if err != nil || len(got) != 2 || got[0].Value != "Bearer old" || got[1].Value != "v" {
		t.Fatalf("keep + add: %+v, %v", got, err)
	}

	got, err = mergeHeaders(existing, []dto.WebhookHeaderForm{{Name: "Authorization", Value: "Bearer new"}})
	if err != nil || got[0].Value != "Bearer new" {
		t.Fatalf("replace: %+v, %v", got, err)
	}

	if _, err = mergeHeaders(existing, []dto.WebhookHeaderForm{{Name: "Other"}}); err == nil {
		t.Fatal("blank value for an unknown header must fail")
	}

	got, err = mergeHeaders(existing, nil)
	if err != nil || len(got) != 0 {
		t.Fatalf("removal: %+v, %v", got, err)
	}
}

func TestClassify(t *testing.T) {
	cases := []struct {
		status int
		ok     bool
		retry  bool
	}{
		{200, true, false}, {204, true, false},
		{301, false, false}, {401, false, false}, {404, false, false},
		{429, false, true}, {500, false, true}, {503, false, true},
	}
	for _, c := range cases {
		o := classify(c.status, "")
		if o.ok() != c.ok || o.Retry != c.retry {
			t.Errorf("status %d: ok=%v retry=%v", c.status, o.ok(), o.Retry)
		}
	}
}

func TestSend(t *testing.T) {
	var gotAuth, gotMethod, gotType string
	var gotBody payload
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth, gotMethod, gotType = r.Header.Get("Authorization"), r.Method, r.Header.Get("Content-Type")
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	out := send(context.Background(), newHTTPClient(), http.MethodPost, srv.URL,
		[]header{{Name: "Authorization", Value: "Bearer t"}}, payload{Event: "pack.changed", PackID: 7, Slug: "p"})
	if !out.ok() || out.StatusCode != 200 {
		t.Fatalf("outcome: %+v", out)
	}
	if gotAuth != "Bearer t" || gotMethod != "POST" || gotType != "application/json" || gotBody.PackID != 7 || gotBody.Event != "pack.changed" {
		t.Errorf("request: auth=%q method=%q type=%q body=%+v", gotAuth, gotMethod, gotType, gotBody)
	}
}

func TestSendGetHasNoBody(t *testing.T) {
	var gotType string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotType = r.Header.Get("Content-Type")
	}))
	defer srv.Close()
	if out := send(context.Background(), newHTTPClient(), http.MethodGet, srv.URL, nil, payload{}); !out.ok() {
		t.Fatalf("outcome: %+v", out)
	}
	if gotType != "" {
		t.Errorf("GET sent content type %q", gotType)
	}
}

func TestSendDoesNotFollowRedirects(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("redirect target was called")
	}))
	defer target.Close()
	srv := httptest.NewServer(http.RedirectHandler(target.URL, http.StatusFound))
	defer srv.Close()

	out := send(context.Background(), newHTTPClient(), http.MethodPost, srv.URL, nil, payload{})
	if out.ok() || out.Retry || out.StatusCode != http.StatusFound {
		t.Errorf("outcome: %+v", out)
	}
}

func TestSendRetriesServerErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()
	out := send(context.Background(), newHTTPClient(), http.MethodPost, srv.URL, nil, payload{})
	if out.ok() || !out.Retry || out.Err != "HTTP 500: boom\n" {
		t.Errorf("outcome: %+v", out)
	}
}

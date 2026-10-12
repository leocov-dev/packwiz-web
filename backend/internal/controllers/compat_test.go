package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"packwiz-web/internal/jobs"
	"packwiz-web/internal/services/compat_svc"
	"packwiz-web/internal/tables"
)

func TestScanCompatStatus(t *testing.T) {
	if got := scanCompatStatus(true); got != http.StatusAccepted {
		t.Errorf("enqueued: got %d want 202", got)
	}
	if got := scanCompatStatus(false); got != http.StatusOK {
		t.Errorf("not enqueued: got %d want 200", got)
	}
}

func TestCompatServerError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int // 0 means nil
	}{
		{"nil", nil, 0},
		{"missing pack", fmt.Errorf("load pack 1: %w", gorm.ErrRecordNotFound), http.StatusNotFound},
		{"archived", fmt.Errorf("compat scan of pack 1: %w", compat_svc.ErrPackArchived), http.StatusConflict},
		{"no job client", jobs.ErrClientUnavailable, http.StatusServiceUnavailable},
		{"other", errors.New("boom"), http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compatServerError(tt.err)
			if tt.want == 0 {
				if got != nil {
					t.Fatalf("want nil, got %v", got)
				}
				return
			}
			if got == nil {
				t.Fatal("want error, got nil")
			}
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			got.JSON(c)
			if rec.Code != tt.want {
				t.Errorf("status: got %d want %d", rec.Code, tt.want)
			}
		})
	}
}

func TestGetCompatRejectsBadPackId(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cc := NewCompatController(nil)
	router.GET("/pack/:packId/compat", cc.GetCompat)
	router.POST("/pack/:packId/compat/scan", cc.ScanCompat)

	for _, req := range []*http.Request{
		httptest.NewRequest(http.MethodGet, "/pack/abc/compat", nil),
		httptest.NewRequest(http.MethodPost, "/pack/abc/compat/scan", nil),
	} {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code < 400 || rec.Code >= 500 {
			t.Errorf("%s %s: got %d want 4xx", req.Method, req.URL.Path, rec.Code)
		}
	}
}

type fakeCompatSvc struct {
	report compat_svc.Report
	scan   compat_svc.ManualScanResult
	err    error
}

func (f fakeCompatSvc) GetCompat(context.Context, uint) (compat_svc.Report, error) {
	return f.report, f.err
}

func (f fakeCompatSvc) RequestManualScan(context.Context, uint) (compat_svc.ManualScanResult, error) {
	return f.scan, f.err
}

func serveCompat(t *testing.T, svc compatService, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cc := NewCompatController(svc)
	router.GET("/pack/:packId/compat", cc.GetCompat)
	router.POST("/pack/:packId/compat/scan", cc.ScanCompat)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(method, path, nil))
	return rec
}

func TestGetCompatNeverScannedShape(t *testing.T) {
	rec := serveCompat(t, fakeCompatSvc{report: compat_svc.Report{Status: compat_svc.StatusUnknown}}, http.MethodGet, "/pack/1/compat")
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"status":    `"unknown"`,
		"counts":    `{"errors":0,"warnings":0,"info":0}`,
		"unscanned": `0`,
		"scannedAt": `null`,
		"runError":  `""`,
		"findings":  `[]`,
	}
	for k, v := range want {
		if got := string(raw[k]); got != v {
			t.Errorf("%s: got %s want %s", k, got, v)
		}
	}
}

func TestGetCompatFindingShape(t *testing.T) {
	report := compat_svc.Report{
		Status:   compat_svc.StatusErrors,
		Run:      &tables.PackCompatRun{Status: tables.CompatDone, ErrorCount: 1},
		Findings: []tables.PackCompatFinding{{Slug: "sodium", Severity: "error", Code: "x", Message: "m"}},
	}
	rec := serveCompat(t, fakeCompatSvc{report: report}, http.MethodGet, "/pack/1/compat")
	var got struct {
		Findings []map[string]json.RawMessage `json:"findings"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Findings) != 1 {
		t.Fatalf("findings: %s", rec.Body.String())
	}
	f := got.Findings[0]
	if string(f["modId"]) != "null" || string(f["detail"]) != "{}" {
		t.Errorf("modId=%s detail=%s, want null and {}", f["modId"], f["detail"])
	}
}

func TestScanCompatResponses(t *testing.T) {
	tests := []struct {
		name     string
		svc      fakeCompatSvc
		wantCode int
		wantBody string
	}{
		{"enqueued", fakeCompatSvc{scan: compat_svc.ManualScanResult{Enqueued: true, Status: compat_svc.StatusQueued}}, http.StatusAccepted, `{"status":"queued","enqueued":true}`},
		{"not enqueued", fakeCompatSvc{scan: compat_svc.ManualScanResult{Status: compat_svc.StatusStale}}, http.StatusOK, `{"status":"stale","enqueued":false}`},
		{"missing pack", fakeCompatSvc{err: fmt.Errorf("load pack 1: %w", gorm.ErrRecordNotFound)}, http.StatusNotFound, ""},
		{"archived pack", fakeCompatSvc{err: compat_svc.ErrPackArchived}, http.StatusConflict, ""},
		{"no job client", fakeCompatSvc{err: jobs.ErrClientUnavailable}, http.StatusServiceUnavailable, ""},
		{"other", fakeCompatSvc{err: errors.New("boom")}, http.StatusInternalServerError, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serveCompat(t, tt.svc, http.MethodPost, "/pack/1/compat/scan")
			if rec.Code != tt.wantCode {
				t.Fatalf("status: got %d want %d", rec.Code, tt.wantCode)
			}
			if tt.wantBody != "" && rec.Body.String() != tt.wantBody {
				t.Errorf("body: got %s want %s", rec.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestGetCompatErrors(t *testing.T) {
	for name, tt := range map[string]struct {
		err  error
		want int
	}{
		"missing":     {fmt.Errorf("load pack 1: %w", gorm.ErrRecordNotFound), http.StatusNotFound},
		"unavailable": {jobs.ErrClientUnavailable, http.StatusServiceUnavailable},
		"other":       {errors.New("boom"), http.StatusInternalServerError},
	} {
		t.Run(name, func(t *testing.T) {
			rec := serveCompat(t, fakeCompatSvc{err: tt.err}, http.MethodGet, "/pack/1/compat")
			if rec.Code != tt.want {
				t.Errorf("got %d want %d", rec.Code, tt.want)
			}
		})
	}
}

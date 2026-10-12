package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"packwiz-web/internal/jobs"
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

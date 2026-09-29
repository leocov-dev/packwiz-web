package controllers

import (
	"net/http"
	"testing"
)

func TestCheckForUpdatesStatus(t *testing.T) {
	if got := checkForUpdatesStatus(true); got != http.StatusAccepted {
		t.Errorf("enqueued: got %d want 202", got)
	}
	if got := checkForUpdatesStatus(false); got != http.StatusOK {
		t.Errorf("not enqueued: got %d want 200", got)
	}
}

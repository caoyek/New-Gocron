package service

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/caoyek/New-Gocron/internal/models"
)

func TestHTTPHandlerUsesConfiguredTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	started := time.Now()
	result, err := new(HTTPHandler).Run(models.Task{
		Command:    server.URL,
		HttpMethod: models.TaskHTTPMethodGet,
		Timeout:    1,
	}, 1)
	if err == nil || !strings.Contains(result, "Client.Timeout") {
		t.Fatalf("expected configured HTTP timeout, result=%q err=%v", result, err)
	}
	if elapsed := time.Since(started); elapsed > 2*time.Second {
		t.Fatalf("configured timeout took too long: %s", elapsed)
	}
}

func TestHTTPHandlerAllowsUnlimitedTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	result, err := new(HTTPHandler).Run(models.Task{
		Command:    server.URL,
		HttpMethod: models.TaskHTTPMethodGet,
		Timeout:    0,
	}, 1)
	if err != nil || result != "ok" {
		t.Fatalf("expected unlimited HTTP request to finish, result=%q err=%v", result, err)
	}
}

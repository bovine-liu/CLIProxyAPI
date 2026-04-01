package management

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/usage"
	coreauth "github.com/router-for-me/CLIProxyAPI/v6/sdk/cliproxy/auth"
	coreusage "github.com/router-for-me/CLIProxyAPI/v6/sdk/cliproxy/usage"
)

func TestListAuthFilesIncludesUsageCounts(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	manager := coreauth.NewManager(nil, nil, nil)
	auth := &coreauth.Auth{
		ID:       "auth-1",
		FileName: "auth-1.json",
		Provider: "codex",
		Attributes: map[string]string{
			"path": "/tmp/auth-1.json",
		},
	}
	if _, err := manager.Register(context.Background(), auth); err != nil {
		t.Fatalf("register auth: %v", err)
	}
	authIndex := auth.EnsureIndex()

	stats := usage.NewRequestStatistics()
	stats.Record(context.Background(), coreusage.Record{
		Provider:  "codex",
		Model:     "gpt-5.4",
		AuthIndex: authIndex,
		Failed:    false,
		Detail:    coreusage.Detail{TotalTokens: 10},
		RequestedAt: time.Now(),
	})
	stats.Record(context.Background(), coreusage.Record{
		Provider:  "codex",
		Model:     "gpt-5.4",
		AuthIndex: authIndex,
		Failed:    true,
		Detail:    coreusage.Detail{},
		RequestedAt: time.Now(),
	})

	h := &Handler{
		authManager: manager,
		usageStats:  stats,
	}

	req := httptest.NewRequest(http.MethodGet, "/v0/management/auth-files", nil)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req

	h.ListAuthFiles(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp struct {
		Files []map[string]any `json:"files"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Files) != 1 {
		t.Fatalf("files len = %d, want 1", len(resp.Files))
	}
	file := resp.Files[0]
	if got := int64(file["success_count"].(float64)); got != 1 {
		t.Fatalf("success_count = %d, want 1", got)
	}
	if got := int64(file["failure_count"].(float64)); got != 1 {
		t.Fatalf("failure_count = %d, want 1", got)
	}
	if got := int64(file["total_requests"].(float64)); got != 2 {
		t.Fatalf("total_requests = %d, want 2", got)
	}
}

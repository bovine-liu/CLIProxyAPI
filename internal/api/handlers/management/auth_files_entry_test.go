package management

import (
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	coreauth "github.com/router-for-me/CLIProxyAPI/v6/sdk/cliproxy/auth"
)

func TestBuildAuthFileEntryIncludesDetailedFailureFields(t *testing.T) {
	t.Parallel()

	now := time.Now()
	h := &Handler{}
	auth := &coreauth.Auth{
		ID:            "aubrey.json",
		FileName:      "aubrey.json",
		Provider:      "codex",
		Status:        coreauth.StatusError,
		StatusMessage: `{"detail":"Unauthorized"}`,
		Unavailable:   true,
		LastError: &coreauth.Error{
			Code:       "token_invalidated",
			Message:    `{"error":{"message":"Your authentication token has been invalidated.","code":"token_invalidated"},"status":401}`,
			HTTPStatus: 401,
		},
		NextRetryAfter: now.Add(30 * time.Minute),
		Attributes: map[string]string{
			"path":  "/tmp/aubrey.json",
			"email": "aubrey@example.com",
		},
		ModelStates: map[string]*coreauth.ModelState{
			"gpt-5.4": {
				Status:         coreauth.StatusError,
				StatusMessage:  `{"detail":"Unsupported model"}`,
				Unavailable:    true,
				NextRetryAfter: now.Add(12 * time.Hour),
				LastError: &coreauth.Error{
					Code:       "model_not_supported",
					Message:    `{"error":{"message":"Model not available for your account","code":"model_not_supported"},"status":404}`,
					HTTPStatus: 404,
				},
				Quota: coreauth.QuotaState{
					Exceeded:     true,
					Reason:       "quota",
					BackoffLevel: 2,
				},
				UpdatedAt: now,
			},
		},
	}

	entry := h.buildAuthFileEntry(auth)
	if entry == nil {
		t.Fatal("expected auth file entry")
	}

	lastError, ok := entry["last_error"].(*coreauth.Error)
	if !ok || lastError == nil {
		t.Fatalf("expected last_error in entry, got %#v", entry["last_error"])
	}
	if lastError.Code != "token_invalidated" {
		t.Fatalf("last_error.code = %q, want %q", lastError.Code, "token_invalidated")
	}

	modelStates, ok := entry["model_states"].(gin.H)
	if !ok {
		t.Fatalf("expected model_states gin.H, got %#v", entry["model_states"])
	}
	modelRaw, ok := modelStates["gpt-5.4"]
	if !ok {
		t.Fatalf("expected gpt-5.4 model state, got %#v", modelStates)
	}
	modelEntry, ok := modelRaw.(gin.H)
	if !ok {
		t.Fatalf("expected gpt-5.4 entry gin.H, got %#v", modelRaw)
	}
	if modelEntry["status_message"] != `{"detail":"Unsupported model"}` {
		t.Fatalf("model status_message = %#v", modelEntry["status_message"])
	}
	if _, ok := modelEntry["last_error"]; !ok {
		t.Fatalf("expected model last_error, got %#v", modelEntry)
	}
	if _, ok := modelEntry["quota"]; !ok {
		t.Fatalf("expected model quota, got %#v", modelEntry)
	}
}

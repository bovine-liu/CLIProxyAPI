package management

import (
	"testing"

	coreauth "github.com/router-for-me/CLIProxyAPI/v6/sdk/cliproxy/auth"
)

func TestExtractCodexIDTokenClaimsFallsBackToAccessTokenAndAccountID(t *testing.T) {
	auth := &coreauth.Auth{
		Provider: "codex",
		Metadata: map[string]any{
			"id_token":     "",
			"access_token": "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJodHRwczovL2FwaS5vcGVuYWkuY29tL2F1dGgiOnsiY2hhdGdwdF9hY2NvdW50X2lkIjoiZDlkYThhMjItOTI5Mi00MGNjLTllOWQtYzc0Yzc4YzA0ZDFmIiwiY2hhdGdwdF9wbGFuX3R5cGUiOiJmcmVlIn19.sig",
			"account_id":   "d9da8a22-9292-40cc-9e9d-c74c78c04d1f",
		},
	}

	claims := extractCodexIDTokenClaims(auth)
	if claims == nil {
		t.Fatal("expected codex claims fallback, got nil")
	}
	if got := claims["chatgpt_account_id"]; got != "d9da8a22-9292-40cc-9e9d-c74c78c04d1f" {
		t.Fatalf("chatgpt_account_id = %v, want %q", got, "d9da8a22-9292-40cc-9e9d-c74c78c04d1f")
	}
	if got := claims["plan_type"]; got != "free" {
		t.Fatalf("plan_type = %v, want %q", got, "free")
	}
}

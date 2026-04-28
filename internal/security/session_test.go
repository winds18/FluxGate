package security

import (
	"testing"
	"time"
)

func TestSessionTokenVerify(t *testing.T) {
	now := time.Date(2026, 4, 28, 8, 0, 0, 0, time.UTC)
	token, err := NewSessionToken("secret", 42, now, time.Hour)
	if err != nil {
		t.Fatalf("new session token: %v", err)
	}
	adminID, ok := VerifySessionToken("secret", token, now.Add(30*time.Minute))
	if !ok || adminID != 42 {
		t.Fatalf("expected valid session, got adminID=%d ok=%v", adminID, ok)
	}
	if _, ok := VerifySessionToken("other-secret", token, now.Add(30*time.Minute)); ok {
		t.Fatal("session should reject wrong secret")
	}
	if _, ok := VerifySessionToken("secret", token, now.Add(2*time.Hour)); ok {
		t.Fatal("session should expire")
	}
}

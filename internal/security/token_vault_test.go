package security

import "testing"

func TestTokenVaultRoundTrip(t *testing.T) {
	encrypted, err := EncryptTokenSecret("secret", "fg_example")
	if err != nil {
		t.Fatalf("encrypt token: %v", err)
	}
	if encrypted == "" || encrypted == "fg_example" {
		t.Fatalf("unexpected encrypted token: %q", encrypted)
	}
	decrypted, err := DecryptTokenSecret("secret", encrypted)
	if err != nil {
		t.Fatalf("decrypt token: %v", err)
	}
	if decrypted != "fg_example" {
		t.Fatalf("unexpected decrypted token: %q", decrypted)
	}
	if _, err := DecryptTokenSecret("other-secret", encrypted); err == nil {
		t.Fatal("decrypt with a different secret should fail")
	}
}

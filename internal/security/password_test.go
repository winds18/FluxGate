package security

import "testing"

func TestPasswordHashVerify(t *testing.T) {
	hash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if hash == "" {
		t.Fatal("password hash missing")
	}
	if !VerifyPassword("correct horse battery staple", hash) {
		t.Fatal("expected password to verify")
	}
	if VerifyPassword("wrong password", hash) {
		t.Fatal("wrong password should not verify")
	}
}

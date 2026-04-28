package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func NewSessionToken(secret string, adminID int64, now time.Time, ttl time.Duration) (string, error) {
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	expires := now.Add(ttl).Unix()
	payload := fmt.Sprintf("%d.%d.%s", adminID, expires, base64.RawURLEncoding.EncodeToString(nonce))
	signature := signSessionPayload(secret, payload)
	return payload + "." + signature, nil
}

func VerifySessionToken(secret, token string, now time.Time) (int64, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 4 {
		return 0, false
	}
	payload := strings.Join(parts[:3], ".")
	if !hmac.Equal([]byte(parts[3]), []byte(signSessionPayload(secret, payload))) {
		return 0, false
	}
	adminID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || adminID <= 0 {
		return 0, false
	}
	expires, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || now.Unix() > expires {
		return 0, false
	}
	return adminID, true
}

func signSessionPayload(secret, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

package limiter

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCookieIdentifier_NewCookie(t *testing.T) {
	ci := newCookieIdentifierWithSecret([]byte("test-secret-key-1234567890123456"))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	id := ci.Identify(w, r)

	assert.NotEmpty(t, id)
	assert.Len(t, id, 32, "ID should be 32 hex chars (16 bytes)")

	cookies := w.Result().Cookies()
	assert.Len(t, cookies, 1)
	assert.Equal(t, "_gl_id", cookies[0].Name)
	assert.True(t, cookies[0].HttpOnly)
	assert.Equal(t, "/", cookies[0].Path)
}

func TestCookieIdentifier_ExistingValidCookie(t *testing.T) {
	ci := newCookieIdentifierWithSecret([]byte("test-secret-key-1234567890123456"))

	// First request: get a cookie
	w1 := httptest.NewRecorder()
	r1 := httptest.NewRequest(http.MethodGet, "/", nil)
	firstID := ci.Identify(w1, r1)

	// Second request: reuse the cookie
	cookieValue := w1.Result().Cookies()[0].Value
	w2 := httptest.NewRecorder()
	r2 := httptest.NewRequest(http.MethodGet, "/", nil)
	r2.AddCookie(&http.Cookie{Name: "_gl_id", Value: cookieValue})
	secondID := ci.Identify(w2, r2)

	assert.Equal(t, firstID, secondID, "Same cookie should return same ID")
	assert.Empty(t, w2.Result().Cookies(), "No new cookie should be set for valid existing cookie")
}

func TestCookieIdentifier_TamperedCookie(t *testing.T) {
	ci := newCookieIdentifierWithSecret([]byte("test-secret-key-1234567890123456"))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.AddCookie(&http.Cookie{Name: "_gl_id", Value: "tampered-id.invalid-signature"})

	id := ci.Identify(w, r)

	assert.NotEmpty(t, id)
	assert.NotEqual(t, "tampered-id", id, "Tampered cookie should generate a new ID")
	cookies := w.Result().Cookies()
	assert.Len(t, cookies, 1, "A new cookie should be set")
}

func TestCookieIdentifier_SignVerifyRoundTrip(t *testing.T) {
	ci := newCookieIdentifierWithSecret([]byte("my-secret"))

	id := ci.generateID()
	sig := ci.sign(id)

	verified, ok := ci.verify(id + "." + sig)
	assert.True(t, ok)
	assert.Equal(t, id, verified)
}

func TestCookieIdentifier_VerifyRejectsBadSignature(t *testing.T) {
	ci := newCookieIdentifierWithSecret([]byte("my-secret"))

	_, ok := ci.verify("someid.badsignature")
	assert.False(t, ok)
}

func TestCookieIdentifier_VerifyRejectsMalformedValue(t *testing.T) {
	ci := newCookieIdentifierWithSecret([]byte("my-secret"))

	_, ok := ci.verify("no-dot-separator")
	assert.False(t, ok)
}

func TestIpIdentifier(t *testing.T) {
	id := &IpIdentifier{Header: "X-Forwarded-For"}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Forwarded-For", "192.168.1.1")

	result := id.Identify(w, r)
	assert.Equal(t, "192.168.1.1", result)
}

package limiter

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
)

type ClientIdentifier interface {
	Identify(w http.ResponseWriter, r *http.Request) string
}

// IpIdentifier identifies clients by an IP address header (e.g. X-Forwarded-For).
type IpIdentifier struct {
	Header string
}

func (id *IpIdentifier) Identify(_ http.ResponseWriter, r *http.Request) string {
	return r.Header.Get(id.Header)
}

const cookieName = "_gl_id"

// CookieIdentifier identifies clients via an HMAC-signed cookie.
type CookieIdentifier struct {
	secret []byte
}

func NewCookieIdentifier() *CookieIdentifier {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		panic("failed to generate HMAC secret: " + err.Error())
	}
	return &CookieIdentifier{secret: secret}
}

// newCookieIdentifierWithSecret is used for testing with a deterministic secret.
func newCookieIdentifierWithSecret(secret []byte) *CookieIdentifier {
	return &CookieIdentifier{secret: secret}
}

func (c *CookieIdentifier) Identify(w http.ResponseWriter, r *http.Request) string {
	if cookie, err := r.Cookie(cookieName); err == nil {
		if id, ok := c.verify(cookie.Value); ok {
			return id
		}
	}

	id := c.generateID()
	signed := id + "." + c.sign(id)
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    signed,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return id
}

func (c *CookieIdentifier) generateID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic("failed to generate random ID: " + err.Error())
	}
	return hex.EncodeToString(b)
}

func (c *CookieIdentifier) sign(id string) string {
	mac := hmac.New(sha256.New, c.secret)
	mac.Write([]byte(id))
	return hex.EncodeToString(mac.Sum(nil))
}

func (c *CookieIdentifier) verify(value string) (string, bool) {
	parts := strings.SplitN(value, ".", 2)
	if len(parts) != 2 {
		return "", false
	}
	id, sig := parts[0], parts[1]
	expected := c.sign(id)
	if !hmac.Equal([]byte(sig), []byte(expected)) {
		return "", false
	}
	return id, true
}

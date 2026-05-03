// Package license validates KiwiFS Enterprise license keys.
//
// License keys are Ed25519-signed JWTs containing the customer ID,
// seat count, expiry, and enabled feature set. The public key is
// embedded in the binary; the private key is held by KiwiFS Authors.
package license

import (
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"sync"
	"time"
)

var (
	ErrNoLicense      = errors.New("no enterprise license configured")
	ErrInvalidLicense = errors.New("invalid license key")
	ErrExpired        = errors.New("license expired")
	ErrSeatLimit      = errors.New("seat limit exceeded")
)

// Feature represents a gated enterprise capability.
type Feature string

const (
	FeatureSSO         Feature = "sso"
	FeatureLDAP        Feature = "ldap"
	FeatureSCIM        Feature = "scim"
	FeatureMFA         Feature = "mfa"
	FeaturePagePerms   Feature = "page_permissions"
	FeatureAuditLog    Feature = "audit_log"
	FeatureAIChat      Feature = "ai_chat"
	FeatureAIAssistant Feature = "ai_assistant"
	FeatureCollab      Feature = "collab"
	FeatureConnectors  Feature = "connectors"
	FeatureVectorExt   Feature = "vector_external"
)

// Claims holds the decoded license payload.
type Claims struct {
	CustomerID string    `json:"customer_id"`
	Plan       string    `json:"plan"` // "enterprise" | "business"
	Seats      int       `json:"seats"`
	Features   []Feature `json:"features"`
	IssuedAt   time.Time `json:"issued_at"`
	ExpiresAt  time.Time `json:"expires_at"`
}

// HasFeature reports whether the license includes the given feature.
func (c *Claims) HasFeature(f Feature) bool {
	for _, feat := range c.Features {
		if feat == f {
			return true
		}
	}
	return false
}

// Validator checks and caches a license key.
type Validator struct {
	mu     sync.RWMutex
	claims *Claims
	pubKey ed25519.PublicKey
}

// New creates a Validator with the embedded public key.
func New() *Validator {
	return &Validator{
		// TODO: embed the real Ed25519 public key once generated
		pubKey: nil,
	}
}

// Load parses and validates a license key string.
// Call this once at server startup with the value of KIWI_LICENSE_KEY.
func (v *Validator) Load(key string) error {
	if key == "" {
		return ErrNoLicense
	}

	// TODO: implement Ed25519 JWT verification
	// For now, parse as plain JSON for development/testing
	var claims Claims
	if err := json.Unmarshal([]byte(key), &claims); err != nil {
		return ErrInvalidLicense
	}

	if !claims.ExpiresAt.IsZero() && time.Now().After(claims.ExpiresAt) {
		return ErrExpired
	}

	v.mu.Lock()
	v.claims = &claims
	v.mu.Unlock()
	return nil
}

// Claims returns the current license claims, or nil if unlicensed.
func (v *Validator) Claims() *Claims {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.claims
}

// IsEnterprise reports whether any valid license is loaded.
func (v *Validator) IsEnterprise() bool {
	return v.Claims() != nil
}

// Check reports whether the given feature is available.
func (v *Validator) Check(f Feature) bool {
	c := v.Claims()
	if c == nil {
		return false
	}
	return c.HasFeature(f)
}

// CheckSeats reports whether the active seat count is within limits.
func (v *Validator) CheckSeats(active int) error {
	c := v.Claims()
	if c == nil {
		return ErrNoLicense
	}
	if c.Seats > 0 && active > c.Seats {
		return ErrSeatLimit
	}
	return nil
}

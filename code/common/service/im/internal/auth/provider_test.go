package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"im/internal/domain"

	"github.com/golang-jwt/jwt/v4"
)

func TestJWTProviderAuthenticate(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	pubPath := filepath.Join(t.TempDir(), "public.pem")
	publicKeyDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("MarshalPKIXPublicKey() error = %v", err)
	}
	if err := os.WriteFile(pubPath, pemEncodePublicKey(publicKeyDER), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	provider, err := NewJWTProvider(pubPath)
	if err != nil {
		t.Fatalf("NewJWTProvider() error = %v", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{"uid": 123})
	signed, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	principal, err := provider.Authenticate(signed, domain.DomainTenant, domain.Scope{
		TenantID:    "tenant-1",
		ProjectID:   "project-1",
		Environment: "prod",
	})
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if principal.UserID != 123 {
		t.Fatalf("principal.UserID = %d, want 123", principal.UserID)
	}
}

func pemEncodePublicKey(raw []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: raw,
	})
}

package auth

import (
	"fmt"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type EntraClaims struct {
	Oid   string `json:"oid"`
	Roles []Role `json:"roles"`
	jwt.RegisteredClaims
}

type Role string

const (
	Student Role = "student"
	Lector       = "lector"
	Admin        = "admin"
)

var jwks *keyfunc.JWKS

func InitAuth() {
	var err error
	jwksURL := fmt.Sprintf(
		"https://login.microsoftonline.com/%s/discovery/v2.0/keys",
		os.Getenv("TENANT_ID"),
	)
	jwks, err = keyfunc.Get(jwksURL, keyfunc.Options{
		RefreshInterval: time.Hour,
	})

	if err != nil {
		panic("Failed to get JWKS: " + err.Error())
	}
}

// devProfiles maps a static test token to a fixed profile UUID.
var devProfiles = map[string]uuid.UUID{
	"dev-jon":   uuid.MustParse("00000000-0000-0000-0000-000000000001"),
	"dev-wout":  uuid.MustParse("00000000-0000-0000-0000-000000000002"),
	"dev-bart":  uuid.MustParse("00000000-0000-0000-0000-000000000003"),
	"dev-steen": uuid.MustParse("00000000-0000-0000-0000-000000000004"),
	"dev-tim":   uuid.MustParse("00000000-0000-0000-0000-000000000005"),
	"dev-lien":  uuid.MustParse("00000000-0000-0000-0000-000000000006"),
	"dev-kris":  uuid.MustParse("00000000-0000-0000-0000-000000000007"),
	"dev-ann":   uuid.MustParse("00000000-0000-0000-0000-000000000008"),
	"dev-rein":  uuid.MustParse("00000000-0000-0000-0000-000000000009"),
	"dev-noel":  uuid.MustParse("00000000-0000-0000-0000-000000000010"),
	"dev-mark":  uuid.MustParse("00000000-0000-0000-0000-000000000011"),
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "No token"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// ── Dev bypass ────────────────────────────────────────────────────────

		if profileID, ok := devProfiles[tokenString]; ok {
			c.Set("profileID", profileID)
			c.Set("roles", []Role{Student})
			c.Next()
			return
		}

		// ─────────────────────────────────────────────────────────────────────

		// Token validation
		claims := &EntraClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, jwks.Keyfunc)
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// Audience check
		expectedAud := os.Getenv("CLIENT_ID")
		found := false
		for _, aud := range claims.Audience {
			if aud == expectedAud {
				found = true
				break
			}
		}
		if !found {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid audience"})
			return
		}

		// Issuer check
		expectedIssuer := fmt.Sprintf("https://login.microsoftonline.com/%s/v2.0", os.Getenv("TENANT_ID"))
		if claims.Issuer != expectedIssuer {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid issuer"})
			return
		}

		if claims.Oid != "" {
			if profileID, err := uuid.Parse(claims.Oid); err == nil {
				c.Set("profileID", profileID)
			}
		}
		c.Set("roles", claims.Roles)

		c.Next()
	}
}

func RequireRole(role Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, exists := c.Get("roles")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "No roles"})
			return
		}

		if slices.Contains(roles.([]Role), role) {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
	}
}

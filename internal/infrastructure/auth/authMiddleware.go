package auth

import (
	"fmt"
	"net/http"
	"os"
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
)

var jwks *keyfunc.JWKS

func init() {
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

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "No token"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

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

		for _, r := range roles.([]Role) {
			if r == role {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
	}
}

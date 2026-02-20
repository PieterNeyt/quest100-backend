package demo

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// DemoSecret is used to sign/verify demo tokens
// In production: delete this entire file
var DemoSecret = []byte(os.Getenv("DEMO_SECRET"))

const DemoIssuer = "demo-auth"

// IsDemoToken checks if a token is a demo token and sets profileID if so.
// Returns true if it was a demo token (regardless of validity).
func IsDemoToken(c *gin.Context) (handled bool) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return false
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Peek at claims without verifying signature yet
	parser := jwt.NewParser()
	unverified, _, err := parser.ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return false
	}

	mapClaims, ok := unverified.Claims.(jwt.MapClaims)
	if !ok {
		return false
	}

	// Only handle tokens issued by our demo issuer
	iss, _ := mapClaims["iss"].(string)
	if iss != DemoIssuer {
		return false
	}

	// Now verify with our demo secret
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return DemoSecret, nil
	})

	if err != nil || !token.Valid {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid demo token"})
		return true
	}

	sub, _ := mapClaims["sub"].(string)
	profileID, err := uuid.Parse(sub)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid demo token subject"})
		return true
	}

	c.Set("profileID", profileID)
	c.Next()
	return true
}

// GenerateDemoToken creates a signed JWT for a demo profile
func GenerateDemoToken(profileID uuid.UUID) (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer:    DemoIssuer,
		Subject:   profileID.String(),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(DemoSecret)
}

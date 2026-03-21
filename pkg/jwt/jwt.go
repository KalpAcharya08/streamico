package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims defines the payload included in the JSON Web Token.
// We embed jwt.RegisteredClaims to handle standard fields like exp (expiration).
type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

var (
	ErrTokenExpired = errors.New("token has expired")
	ErrTokenInvalid = errors.New("token is invalid")
)

// GenerateToken creates a new signed JWT for a given user.
//
// Parameters:
//   - userID: The UUID of the user.
//   - role: The user's role (e.g., "viewer", "admin").
//   - secret: The secret key used to sign the token (keep this safe!).
//   - ttl: How long until the token expires (e.g., 15 * time.Minute).
//
// Returns:
//   - string: The encoded, signed JWT token string.
//   - error: Any error that occurred during generation.
func GenerateToken(userID, role, secret string, ttl time.Duration) (string, error) {
	now := time.Now()

	// 1. Create the claims (the payload)
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	// 2. Create the token using HMAC SHA256 (HS256)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 3. Sign the token using our secret key
	// Note: We use []byte(secret) because the HS256 algorithm requires a byte slice for the signature.
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

// ValidateToken parses a token string and verifies its signature and claims (like expiration).
//
// Parameters:
//   - tokenString: The JWT sent by the client.
//   - secret: The secret key to verify the signature.
//
// Returns:
//   - *Claims: The decoded payload if the token is valid.
//   - error: An error if the token is invalid, expired, or tampered with.
func ValidateToken(tokenString, secret string) (*Claims, error) {
	// Let the jwt library parse the token string and verify the signature using our secret.
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Ensure the token's signing method is what we expect (HMAC in this case).
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// Return the secret key to be used for validation.
		return []byte(secret), nil
	})

	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			return nil, ErrTokenExpired
		default:
			return nil, fmt.Errorf("%w: %v", ErrTokenInvalid, err)
		}
	}

	// Double-check if the token was parsed successfully and extraction succeeds
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrTokenInvalid
}

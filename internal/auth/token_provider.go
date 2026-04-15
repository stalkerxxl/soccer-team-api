package auth

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken        = errors.New("invalid token")
	ErrTokenExpired        = errors.New("token expired")
	ErrInvalidTokenSubject = errors.New("invalid token subject")
)

// TokenProvider defines access token generation and parsing operations.
type TokenProvider interface {
	GenerateAccessToken(userID int64) (token string, expiresAt time.Time, err error)
	ParseAccessToken(token string) (TokenClaims, error)
}

// TokenClaims contains the application-level fields extracted from an access token.
type TokenClaims struct {
	UserID    int64
	IssuedAt  time.Time
	ExpiresAt time.Time
}

// JWTTokenProvider issues and validates HS256-signed access tokens.
type JWTTokenProvider struct {
	accessSecret []byte
	accessTTL    time.Duration
	now          func() time.Time
}

type accessTokenClaims struct {
	jwt.RegisteredClaims
}

// NewJWTTokenProvider constructs a JWT token provider with the given signing secret and TTL.
func NewJWTTokenProvider(accessSecret string, accessTTL time.Duration) (*JWTTokenProvider, error) {
	if accessSecret == "" {
		return nil, errors.New("access secret is required")
	}

	if accessTTL <= 0 {
		return nil, errors.New("access ttl must be greater than 0")
	}

	return &JWTTokenProvider{
		accessSecret: []byte(accessSecret),
		accessTTL:    accessTTL,
		now:          time.Now,
	}, nil
}

// GenerateAccessToken creates a signed access token for the given user ID.
func (p *JWTTokenProvider) GenerateAccessToken(userID int64) (string, time.Time, error) {
	now := p.now().UTC()
	expiresAt := now.Add(p.accessTTL)

	claims := accessTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(userID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(p.accessSecret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign access token: %w", err)
	}

	return signedToken, expiresAt, nil
}

// ParseAccessToken validates a signed access token and returns normalized claims.
func (p *JWTTokenProvider) ParseAccessToken(tokenString string) (TokenClaims, error) {
	claims := new(accessTokenClaims)

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		// Accept only the configured symmetric algorithm to avoid algorithm confusion.
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, ErrInvalidToken
		}

		return p.accessSecret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return TokenClaims{}, ErrTokenExpired
		}

		return TokenClaims{}, ErrInvalidToken
	}

	if !token.Valid {
		return TokenClaims{}, ErrInvalidToken
	}

	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		return TokenClaims{}, ErrInvalidTokenSubject
	}

	parsedClaims := TokenClaims{UserID: userID}
	if claims.IssuedAt != nil {
		parsedClaims.IssuedAt = claims.IssuedAt.Time
	}
	if claims.ExpiresAt != nil {
		parsedClaims.ExpiresAt = claims.ExpiresAt.Time
	}

	return parsedClaims, nil
}

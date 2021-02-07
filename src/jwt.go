package main

import (
	"context"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-kit/kit/endpoint"
)

type contextKey string

const (
	// JWTTokenContextKey holds the key used to store a JWT Token in the
	// context.
	JWTTokenContextKey contextKey = "JWTToken"

	// JWTClaimsContextKey holds the key used to store the JWT Claims in the
	// context.
	JWTClaimsContextKey contextKey = "JWTClaims"
)

var (
	// ErrTokenContextMissing denotes a token was not passed into the parsing
	// middleware's context.
	ErrTokenContextMissing = errors.New("token up for parsing was not passed through the context")

	// ErrTokenInvalid denotes a token was not able to be validated.
	ErrTokenInvalid = errors.New("JWT Token was invalid")

	// ErrTokenExpired denotes a token's expire header (exp) has since passed.
	ErrTokenExpired = errors.New("JWT Token is expired")

	// ErrTokenMalformed denotes a token was not formatted as a JWT token.
	ErrTokenMalformed = errors.New("JWT Token is malformed")

	// ErrTokenNotActive denotes a token's not before header (nbf) is in the
	// future.
	ErrTokenNotActive = errors.New("token is not valid yet")

	// ErrUnexpectedSigningMethod denotes a token was signed with an unexpected
	// signing method.
	ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
)

// ClaimsFactory is a factory for jwt.Claims.
// Useful in NewParser middleware.
type ClaimsFactory func() jwt.Claims

// MapClaimsFactory is a ClaimsFactory that returns
// an empty jwt.MapClaims.
func MapClaimsFactory() jwt.Claims {
	return jwt.MapClaims{}
}

// StandardClaimsFactory is a ClaimsFactory that returns
// an empty jwt.StandardClaims.
func StandardClaimsFactory() jwt.Claims {
	return &jwt.StandardClaims{}
}

// NewParser creates a new JWT token parsing middleware, specifying a
// jwt.Keyfunc interface, the signing method and the claims type to be used. NewParser
// adds the resulting claims to endpoint context or returns error on invalid token.
// Particularly useful for servers.
func NewParser(keyFunc jwt.Keyfunc, method jwt.SigningMethod, newClaims ClaimsFactory) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			// tokenString is stored in the context from the transport handlers.
			tokenString, ok := ctx.Value(JWTTokenContextKey).(string)
			if !ok {
				return nil, &jwt.ValidationError{Inner: ErrTokenContextMissing}
			}

			// Parse takes the token string and a function for looking up the
			// key. The latter is especially useful if you use multiple keys
			// for your application.  The standard is to use 'kid' in the head
			// of the token to identify which key to use, but the parsed token
			// (head and claims) is provided to the callback, providing
			// flexibility.
			token, err := jwt.ParseWithClaims(tokenString, newClaims(), func(token *jwt.Token) (interface{}, error) {
				// Don't forget to validate the alg is what you expect:
				if token.Method != method {
					return nil, &jwt.ValidationError{Inner: ErrUnexpectedSigningMethod}
				}

				return keyFunc(token)
			})
			if err != nil {
				if e, ok := err.(*jwt.ValidationError); ok {
					switch {
					case e.Errors&jwt.ValidationErrorMalformed != 0:
						// Token is malformed
						return nil, e
					case e.Errors&jwt.ValidationErrorExpired != 0:
						// Token is expired
						return nil, e
					case e.Errors&jwt.ValidationErrorNotValidYet != 0:
						// Token is not active yet
						return nil, e
					case e.Inner != nil:
						// report e.Inner
						return nil, e.Inner
					}
					// We have a ValidationError but have no specific Go kit error for it.
					// Fall through to return original error.
				}
				return nil, err
			}

			if !token.Valid {
				return nil, &jwt.ValidationError{Inner: ErrTokenInvalid}
			}

			ctx = context.WithValue(ctx, JWTClaimsContextKey, token.Claims)

			return next(ctx, request)
		}
	}
}

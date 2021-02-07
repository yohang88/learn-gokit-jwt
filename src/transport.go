package main

import (
	"context"
	"encoding/json"
	stdjwt "github.com/dgrijalva/jwt-go"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport"
	"strings"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"net/http"
)

func MakeHTTPHandler(jwtPublicKey string, s Service, logger log.Logger) http.Handler {
	r := mux.NewRouter()
	e := makeServerEndpoints(s)

	publicKey := []byte(jwtPublicKey)
	pem, _ := stdjwt.ParseRSAPublicKeyFromPEM(publicKey)
	kf := func(token *stdjwt.Token) (interface{}, error) { return pem, nil }

	options := []httptransport.ServerOption{
		httptransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		httptransport.ServerErrorEncoder(encodeError),
		httptransport.ServerBefore(HTTPToContext()),
	}

	r.Methods("GET").Path("/health").Handler(httptransport.NewServer(
		e.GetHealthCheckEndpoint,
		decodeHealthCheckRequest,
		encodeResponse,
		options...,
	))

	r.Methods("GET").Path("/user").Handler(httptransport.NewServer(
		NewParser(kf, stdjwt.SigningMethodRS256, MapClaimsFactory)(e.GetUserEndpoint),
		decodeGetUserRequest,
		encodeResponse,
		options...,
	))

	return r
}

type HealthCheckRequest struct{}

type HealthCheckResponse struct {
	Status string `json:"status"`
}

func decodeHealthCheckRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req HealthCheckRequest
	return req, nil
}

type GetUserRequest struct{}

type GetUserResponse struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func decodeGetUserRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req GetUserRequest
	return req, nil
}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	return json.NewEncoder(w).Encode(response)
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	if err == nil {
		panic("encodeError with nil error")
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	switch err.(type) {
	case *stdjwt.ValidationError:
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Unauthenticated (Invalid Token).",
		})
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": err.Error(),
	})
}

const (
	bearer       string = "bearer"
	bearerFormat string = "Bearer %s"
)

// HTTPToContext moves a JWT from request header to context. Particularly
// useful for servers.
func HTTPToContext() httptransport.RequestFunc {
	return func(ctx context.Context, r *http.Request) context.Context {
		token, ok := extractTokenFromAuthHeader(r.Header.Get("Authorization"))
		if !ok {
			return ctx
		}

		return context.WithValue(ctx, JWTTokenContextKey, token)
	}
}

func extractTokenFromAuthHeader(val string) (token string, ok bool) {
	authHeaderParts := strings.Split(val, " ")
	if len(authHeaderParts) != 2 || !strings.EqualFold(authHeaderParts[0], bearer) {
		return "", false
	}

	return authHeaderParts[1], true
}

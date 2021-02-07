package main

import (
	"context"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-kit/kit/endpoint"
)

type Endpoints struct {
	GetHealthCheckEndpoint endpoint.Endpoint
	GetUserEndpoint        endpoint.Endpoint
}

func makeServerEndpoints(s Service) Endpoints {
	return Endpoints{
		GetHealthCheckEndpoint: makeHealthCheckEndpoint(s),
		GetUserEndpoint:        makeGetUserEndpoint(s),
	}
}
func makeHealthCheckEndpoint(svc Service) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		status, _ := svc.HealthCheck()

		return HealthCheckResponse{status}, nil
	}
}
func makeGetUserEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		claims := ctx.Value(JWTClaimsContextKey).(jwt.MapClaims)

		user, _ := svc.GetUser(claims["sub"].(string), claims["name"].(string), claims["email"].(string))

		return GetUserResponse{
			Id:    user.Id,
			Name:  user.Name,
			Email: user.Email,
		}, nil
	}
}

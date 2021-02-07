package main

import (
	"github.com/go-kit/kit/log"
	"time"
)

type Middleware func(Service) Service

func LoggingMiddleware(logger log.Logger) Middleware {
	return func(next Service) Service {
		return &loggingMiddleware{next: next, logger: logger}
	}
}

type loggingMiddleware struct {
	next   Service
	logger log.Logger
}

func (mw loggingMiddleware) GetUser(id string, name string, email string) (user User, err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "GetUser", "time", time.Since(begin), "err", err)
	}(time.Now())

	return mw.next.GetUser(id, name, email)
}

func (mw loggingMiddleware) HealthCheck() (s string, err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "GetHealthCheck", "time", time.Since(begin), "err", err)
	}(time.Now())

	return mw.next.HealthCheck()
}

package main

import (
	"flag"
	"fmt"
	"github.com/go-kit/kit/log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	jwtPublicKeyStringSrc := os.Getenv("JWT_PUBLIC_KEY")
	jwtPublicKeyString := strings.Replace(jwtPublicKeyStringSrc, `\n`, "\n", -1)

	var (
		httpAddr = flag.String("http", ":8000", "http listen address")
	)
	flag.Parse()

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "timestamp", log.DefaultTimestampUTC)
	}

	var s Service
	{
		s = NewService()
		s = LoggingMiddleware(logger)(s)
	}

	var h http.Handler
	{
		h = MakeHTTPHandler(jwtPublicKeyString, s, log.With(logger, "component", "HTTP"))
	}

	errs := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	go func() {
		logger.Log("transport", "HTTP", "addr", *httpAddr)
		errs <- http.ListenAndServe(*httpAddr, h)
	}()

	logger.Log("exit", <-errs)
}

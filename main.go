package main

import (
	"os"
	"log"
	"fmt"
	"flag"
	"time"
	"context"
	"net/http"
	"os/signal"

	"github.com/made2591/jaker/lib"
)

type key int

const (
	REQUEST_ID_KEY key = 0
)

var (
	PORT    = os.Getenv("PORT")
	ADDRESS = os.Getenv("ADDRESS")
)

func main() {

	// parse configuration
	flag.StringVar(&ADDRESS, "listen-addr", ":"+PORT, "server listen address")
	flag.Parse()

	// log to output
	logger := log.New(os.Stdout, "http: ", log.LstdFlags)
	logger.Println("Server is starting...")

	// setup routing
	router := http.NewServeMux()

	// list images
	router.Handle("/images/list", lib.ListImages())

	// list images
	router.Handle("/images/global-size", lib.GetImagesSize())

	// list containers
	router.Handle("/containers/list", lib.ListContainers())

	// config
	router.Handle("/config", lib.Configuration())

	// notify
	router.Handle("/notify/local/repository/size", lib.NotifyLocalRepositorySize())

	// new request ID
	nextRequestID := func() string {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	// instantiate server
	server := &http.Server{
		Addr:         ADDRESS,
		Handler:      tracing(nextRequestID)(logging(logger)(router)),
		ErrorLog:     logger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	// channels
	done := make(chan bool)
	quit := make(chan os.Signal, 1)

	// stop
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		logger.Println("Server is shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			logger.Fatalf("Could not gracefully shutdown the server: %v\n", err)
		}
		close(done)
	}()

	logger.Println("Server is ready to handle requests at", ADDRESS)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("Could not listen on %s: %v\n", ADDRESS, err)
	}

	<-done
	logger.Println("Server stopped")
}

func logging(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				requestID, ok := r.Context().Value(REQUEST_ID_KEY).(string)
				if !ok {
					requestID = "unknown"
				}
				logger.Println(requestID, r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func tracing(nextRequestID func() string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-Id")
			if requestID == "" {
				requestID = nextRequestID()
			}
			ctx := context.WithValue(r.Context(), REQUEST_ID_KEY, requestID)
			w.Header().Set("X-Request-Id", requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

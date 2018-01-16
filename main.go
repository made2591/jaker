package main

import (

	"os"
	"log"
	"fmt"
	"flag"
	"time"
	"strconv"
	"context"
	"net/http"
	"os/signal"
	"sync/atomic"
	"encoding/json"

	"github.com/made2591/jaker/lib"

	"github.com/docker/docker/client"
	"github.com/docker/docker/api/types"

)

type key int

const (
	requestIDKey key = 0
)

var (
	listenAddr string
	healthy    int32
	jonfiguration = lib.Jonfiguration{ Port: 5000, Alerts: []lib.Jalert{}}
)

func main() {
	flag.StringVar(&listenAddr, "listen-addr", ":"+strconv.Itoa(jonfiguration.Port), "server listen address")
	flag.Parse()

	logger := log.New(os.Stdout, "http: ", log.LstdFlags)
	logger.Println("Server is starting...")

	router := http.NewServeMux()
	router.Handle("/listc", listc())
	router.Handle("/listi", listi())
	router.Handle("/config", config())
	router.Handle("/notify", lib.Notify("Docker image repository", "Limit reached: local repository size 9.8Gb"))

	nextRequestID := func() string {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	server := &http.Server{
		Addr:         listenAddr,
		Handler:      tracing(nextRequestID)(logging(logger)(router)),
		ErrorLog:     logger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		logger.Println("Server is shutting down...")
		atomic.StoreInt32(&healthy, 0)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			logger.Fatalf("Could not gracefully shutdown the server: %v\n", err)
		}
		close(done)
	}()

	logger.Println("Server is ready to handle requests at", listenAddr)
	atomic.StoreInt32(&healthy, 1)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("Could not listen on %s: %v\n", listenAddr, err)
	}

	<-done
	logger.Println("Server stopped")
}

func config() http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Body == nil {
			http.Error(w, "Please send a request body", http.StatusBadRequest)
			return
		}
		err := json.NewDecoder(r.Body).Decode(&jonfiguration)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		b, err := json.Marshal(jonfiguration)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(b)
	})

}

func listc() http.Handler {

	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	jontainers := []lib.Jontainer{}
	for _, container := range containers {
		//fmt.Printf("%s %s\n", container.ID[:10], container.Image)
		jontainers = append(jontainers, lib.Jontainer{Id: container.ID[:10], Name: container.Names[0], Image: container.Image, Status: container.Status})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(containers) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		} else {
			b, err := json.Marshal(jontainers)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Write(b)
		}
		w.WriteHeader(http.StatusServiceUnavailable)
	})

}

func listi() http.Handler {

	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	images, err := cli.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		panic(err)
	}

	jmages := []lib.Jmage{}
	for _, image := range images {
		jmages = append(jmages, lib.Jmage{Id: image.ID[:10], Name: image.RepoDigests, Size: image.Size})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(images) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		} else {
			b, err := json.Marshal(jmages)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Write(b)
		}
		w.WriteHeader(http.StatusServiceUnavailable)
	})

}

func logging(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				requestID, ok := r.Context().Value(requestIDKey).(string)
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
			ctx := context.WithValue(r.Context(), requestIDKey, requestID)
			w.Header().Set("X-Request-Id", requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

package main

import (
	"context"
	"fmt"

	"os"
	"log"
	"flag"
	"time"
	"net/http"
	"os/signal"
	"sync/atomic"
	"encoding/json"

	"github.com/made2591/lib"

	"github.com/0xAX/notificator"

	"github.com/docker/docker/client"
	"github.com/docker/docker/api/types"

)

var (
	listenAddr string
	healthy    int32
)

func notify() http.Handler {

	var notify *notificator.Notificator

	notify = notificator.New(notificator.Options{
		DefaultIcon: "icon/default.png",
		AppName:     "My test App",
	})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		notify.Push("title", "text", "/home/user/icon.png", notificator.UR_NORMAL)
		w.WriteHeader(http.StatusOK)
	})
}

func main() {
	flag.StringVar(&listenAddr, "listen-addr", ":5000", "server listen address")
	flag.Parse()

	logger := log.New(os.Stdout, "http: ", log.LstdFlags)
	logger.Println("Server is starting...")

	router := http.NewServeMux()
	router.Handle("/", index())
	router.Handle("/listc", listc())
	router.Handle("/notify", notify())

	server := &http.Server{
		Addr:         listenAddr,
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

func index() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Hello, World!")
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

	jontainers := []Jontainer{}
	for _, container := range containers {
		fmt.Printf("%s %s\n", container.ID[:10], container.Image)
		jontainers = append(jontainers, Jontainer{Id: container.ID[:10], Name: container.Names[0], Image: container.Image, Status: container.Status})
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
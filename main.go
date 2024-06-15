package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

func main() {
    ctx := context.Background()

    getEnv := func(key string) string {
        switch key {
        case "MYAPP_FORMAT":
            return "myformat"
        case "ENV":
            return "dev"
        default:
            return ""
    }
    }

    if err := run(ctx, os.Args, getEnv, os.Stdin, os.Stdout, os.Stderr); err != nil {
        fmt.Fprintf(os.Stderr, "%s\n", err)
        os.Exit(1)
    }
}

func run(
    ctx context.Context,
    args []string,
    getenv func(string)string,
    stdin io.Reader,
    stdout, stderr io.Writer,
) error {

    ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
    defer cancel()

    logger := &Logger{}
    config := &Config{
        Host: "",
        Port: "",
    }
    store := &Store{}

    srv := NewServer(
        logger,
        config,
        store,
        )

    httpServer := &http.Server{
        Addr: net.JoinHostPort(config.Host, config.Port),
        Handler: srv,
    }

    go func() {
        log.Printf("listening on %s\n", httpServer.Addr)
        if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
        }
    }()

    var wg sync.WaitGroup
    wg.Add(1)

    go func() {
        defer wg.Done()
        <-ctx.Done()

        // make a new context for the shutdown
        shutdownCtx := context.Background()
        shutdownCtx, cancel := context.WithTimeout(ctx, 10 * time.Second)
        defer cancel()

        if err := httpServer.Shutdown(shutdownCtx); err != nil {
            fmt.Fprintf(os.Stderr, "error shuting down http server: %s\n", err)
        }
    }()

    wg.Wait()

    return nil
}

func getEnv(name string) string {
    return name
}

type Logger struct {
}

type Config struct {
    Host string
    Port string
}

type Store struct {
}

// Responsible for all top-level http configurations that applies to all endpoints
// like CORS, auth middleware, and logging
func NewServer(
    logger *Logger,
    config *Config,
    store *Store,
) http.Handler {
    mux := http.NewServeMux()

    addRoutes(
        mux,
        logger,
        config,
        store,
        )

    var handler http.Handler = mux

    //handler = someMiddelware(handler)
    //handler = someMiddelware2(handler)
    //handler = someMiddelware3(handler)

    return handler
}

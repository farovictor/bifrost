package main

import (
    "fmt"
    "log"
    "net/http"
    "os"

    "github.com/go-chi/chi/v5"
)

func main() {
    port := os.Getenv("TESTSERVER_PORT")
    if port == "" {
        port = "8081"
    }

    r := chi.NewRouter()
    r.Get("/check", func(w http.ResponseWriter, r *http.Request) {
        if r.Header.Get("X-API-Key") != "secret" {
            w.WriteHeader(http.StatusInternalServerError)
            return
        }
        w.WriteHeader(http.StatusOK)
        fmt.Fprint(w, "pong")
    })

    addr := ":" + port
    log.Printf("starting test server on %s", addr)
    if err := http.ListenAndServe(addr, r); err != nil {
        log.Fatalf("listen and serve: %v", err)
    }
}

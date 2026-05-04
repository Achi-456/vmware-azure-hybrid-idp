package main

import (
    "fmt"
    "net/http"
    "os"
)

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        hostname, _ := os.Hostname()
        fmt.Fprintf(w, "DCLab IDP Platform | Pod: %s | Built by Gitea Actions\n", hostname)
    })
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(200)
        fmt.Fprint(w, "ok")
    })
    http.ListenAndServe(":8080", nil)
}

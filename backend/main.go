package main

import (
    "net/http"
    "log"
)

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("OpenWaterMap Backend"))
    })

    log.Println("Server is starting on port 8080...")
    http.ListenAndServe(":8080", nil)
}
package main

import (
    "fmt"
    "net/http"
)

func healthz(w http.ResponseWriter, req *http.Request) {

    fmt.Fprintf(w, "OK\n")
}


func main() {

    http.HandleFunc("/healthz", healthz)
    // http.HandleFunc("/headers", headers)

    http.ListenAndServe(":8090", nil)
	// fmt.Println("hello world")
}

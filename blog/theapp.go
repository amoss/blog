package main

import (
      "net/http"
      "fmt"
)

func handler(out http.ResponseWriter, req *http.Request) {
  fmt.Println("Req:", req.URL)
}

func main() {
    http.HandleFunc("/", handler)
    http.ListenAndServe(":8080", nil)
}

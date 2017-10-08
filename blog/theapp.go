package main

import (
      "net/http",
      "fmt"
)

func handler(out http.ResponseWriter, req *http.Request) 
{
  fmt.PrintLn("Req:", req.url)
}

func main() 
{
    http.HandleFunc("/", handler())
    http.ListenAndServe(":8080", nil)
}

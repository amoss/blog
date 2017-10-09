package main

import (
      "net/http"
      "fmt"
      "io/ioutil"
      "strings"
)

func parseRst(src string) {
  lines := strings.Split(src,"\n")
  for i:= 0; i<len(lines); i++ {
    fmt.Println(i, lines[i])
  }
}

func handler(out http.ResponseWriter, req *http.Request) {
  fmt.Println("Req:", req.URL)
  filename := "data" + req.URL.Path + ".rst"
  cnt, err := ioutil.ReadFile(filename)
  if err==nil {
    parseRst( string(cnt) )
  } else {
    fmt.Println(err)
  }
}

func main() {
    http.HandleFunc("/", handler)
    http.ListenAndServe(":8080", nil)
}

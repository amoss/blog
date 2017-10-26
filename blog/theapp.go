package main

import (
    "runtime/debug"
      "net/http"
      "fmt"
      "io/ioutil"
      "errors"
      "os"
      "path"
      "time"
      //"strings"
      //"regexp"
)

func handler(out http.ResponseWriter, req *http.Request) {
    t := time.Now()
    fmt.Println(t.Format(time.RFC1123),"Request", *req)
    defer func() {
        r:= recover()
        if r!=nil {
            fmt.Printf("Panic during page handler! %s %s\n", r, debug.Stack() )
            http.Error(out, errors.New("Something went wrong :(").Error(),
                       http.StatusInternalServerError)
        }
    }()
    switch req.URL.Path {
        case "/page.css", "/styles.css", "/book-icon.png",
             "/Basic-Regular.ttf", "/Inconsolata-Regular.ttf",
             "/SourceSansPro-Regular.otf", "/slides.css",
             "/slides.js", "/logo.svg", "/leftarrow.svg",
             "/rightarrow.svg", "/closearrow.svg", "/settings.svg":
            target := "data" + req.URL.Path
            fmt.Printf("%29s: Path whitelisted - served from", "handler", target)
            cnt,_ := ioutil.ReadFile(target)
            switch path.Ext(req.URL.Path) {
                case ".svg": out.Header().Set("Content-type", "image/svg+xml")
            }
            out.Write(cnt)
        default:
            filename := "data" + req.URL.Path + ".rst"
            // PANIC in trace comes from lack of error checking
            // Video and images not available...
            lines  := LineScanner(filename)
            if lines!=nil {
                fmt.Printf("%29s: Path default - served from", "handler", filename)
                blocks := parse(*lines)
                out.Write( renderHtml(blocks) )
            } else {
              fmt.Printf("%29s: File not found!", "handler", filename)
            }
    }
}

func main() {
    for _,arg := range os.Args[1:] {
        fmt.Println(arg)
        switch arg {
            case "--lines":  lineScannerDbg  = true
            case "--parse":  lineParserStDbg = true
            case "--blocks": lineParserDbg   = true
            default: panic("Unrecognised arg "+arg)
        }
    }
    http.HandleFunc("/", handler)
    http.ListenAndServe(":8080", nil)
}

/*
func main() {
    lines  := LineScanner(os.Args[1])
    blocks := parse(*lines)
    renderHtml(blocks)
}*/


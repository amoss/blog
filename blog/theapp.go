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
      "strings"
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
            fmt.Printf("%29s: Path whitelisted - served from %s\n", "handler", target)
            cnt,_ := ioutil.ReadFile(target)
            switch path.Ext(req.URL.Path) {
                case ".svg": out.Header().Set("Content-type", "image/svg+xml")
            }
            out.Write(cnt)
        default:
            if req.URL.Path[ len(req.URL.Path)-1 ] == '/' {
                http.Redirect(out,req,req.URL.Path+"index.html",302)
            }
            switch( path.Ext(req.URL.Path) ) {
                case ".jpg":
                    filename := "data" + req.URL.Path
                    cnt,err := ioutil.ReadFile(target)
                    if err!=nil && os.IsNotExist(err) {
                        out.WriteHeader(404)
                        fmt.Printf("%29s: Missing image %s\n",
                                   "handler", filename)
                        out.Write( []byte("File not found") )
                    }
                    else
                    {
                        out.Header().Set("Content-type", "image/jpg")
                        out.Write(cnt)
                    }
                case ".html":
                    if strings.HasSuffix(req.URL.Path,"index.html") {
                        dir := path.Dir(req.URL.Path)
                        fmt.Printf("Index detected %s <- %s\n",dir,req.URL.Path)
                        inside  := "data" + dir + "/index.rst"
                        outside := "data" + dir + ".rst"
                        insideFI, insideErr := os.Stat(inside)
                        fmt.Printf("%s: %s, %s\n", inside, insideFI, insideErr)
                        outsideFI, outsideErr := os.Stat(outside)
                        fmt.Printf("%s: %s, %s\n", outside, outsideFI, outsideErr)
                        if insideFI==nil && outsideFI==nil {
                            out.WriteHeader(404)
                            fmt.Printf("%29s: Can't resolve %s or %s\n",
                                       "handler", inside, outside)
                            out.Write( []byte("File not found") )
                            return
                        }
                        if insideFI!=nil && outsideFI!=nil {
                            out.WriteHeader(404)
                            fmt.Printf("%29s: Ambiguous! Can resolve %s AND %s\n",
                                       "handler", inside, outside)
                            out.Write( []byte("File not found (ambiguous configuration)") )
                            return
                        }
                        var filename string
                        if insideFI!=nil {
                            filename = inside
                        } else {
                            filename = outside
                        }
                        lines  := LineScanner(filename)
                        if lines!=nil {
                            fmt.Printf("%29s: Path default - served from %s\n", "handler", filename)
                            blocks := parse(*lines)
                            out.Write( renderHtml(blocks) )
                            return
                        } else {
                            fmt.Printf("%29s: File not found AFTER check! %s\n", "handler", filename)
                            out.WriteHeader(404)
                            out.Write( []byte("File not found") )
                            return
                        }
                    }
                    if path.Ext(req.URL.Path)==".jpg" {
                        fmt.Printf("-----> Check file-ext on\n",req.URL.Path)
                    }
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


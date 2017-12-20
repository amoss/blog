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
      "rst"
      //"regexp"
)

var mimeTypes = map[string]string {
    ".jpg":  "image/jpg",
    ".webm": "video/webm",
    ".mov":  "video/quicktime",
    ".svg":  "image/svg+xml",
}

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
             "/rightarrow.svg", "/closearrow.svg", "/settings.svg",
             "/fliparrow.jpg", "/flipbackarrow.jpg":
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
            mime,ok := mimeTypes[ path.Ext(req.URL.Path) ]
            if ok {
                filename := "data" + req.URL.Path
                cnt,err := ioutil.ReadFile(filename)
                if err!=nil && os.IsNotExist(err) {
                    out.WriteHeader(404)
                    fmt.Printf("%29s: Missing file %s\n",
                               "handler", filename)
                    out.Write( []byte("File not found") )
                } else {
                    out.Header().Set("Content-type", mime)
                    fmt.Printf("%29s: Recognised extension %s, served %s\n",
                               "handler", path.Ext(req.URL.Path), mime)
                    out.Write(cnt)
                }
            } else {
                switch( path.Ext(req.URL.Path) ) {
                    default:
                        out.WriteHeader(404)
                        fmt.Printf("%29s: Unknown file extension %s\n", "handler",
                                   path.Ext(req.URL.Path))
                        out.Write( []byte("Unknown file type") )
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
                            lines  := rst.LineScanner(filename)
                            if lines!=nil {
                                fmt.Printf("%29s: Path default - served from %s\n", "handler", filename)
                                blocks := rst.Parse(*lines)
                                out.Write( rst.RenderHtml(blocks) )
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
}

func main() {
    for _,arg := range os.Args[1:] {
        fmt.Println(arg)
        switch arg {
            case "--lines":  rst.LineScannerDbg  = true
            case "--parse":  rst.LineParserStDbg = true
            case "--blocks": rst.LineParserDbg   = true
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


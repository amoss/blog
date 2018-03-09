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
      "sort"
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

type Post struct {
    Title       []byte
    Filename    []byte
    Tags        []byte
    Date        time.Time
    Subtitle    []byte
}

func ScanPosts() []Post {
    files, err := ioutil.ReadDir("data")
    if err!=nil {
        return make([]Post,0,0)
    }
    posts := make([]Post, 0, len(files))
    for _,entry := range files {
        if path.Ext(entry.Name())==".rst" {
            lines  := rst.LineScanner("data/"+entry.Name())
            if lines!=nil {
                blocks := rst.Parse(*lines)
                headBlock := <-blocks
                for x := range blocks { x = x}
                pTime,err := time.Parse("2006-01-02",string(headBlock.Date))
                if err!=nil{
                    fmt.Printf("Cannot parse %s date: %s\n",entry.Name(), headBlock.Date)
                }else{
                    bName := strings.TrimSuffix( entry.Name(), path.Ext(entry.Name()) )
                    linkName := []byte( bName + "/index.html" )
                    posts = append(posts, Post{Filename:linkName,
                                               Title:headBlock.Title,
                                               Date:pTime,
                                               Tags:headBlock.Tags,
                                               Subtitle:headBlock.Subtitle})
                }
            }
        }
    }
    return posts
}

func renderIndex(posts []Post) []byte {
    result := make([]byte, 0, 16384)
    dates :=  func(i,j int) bool {
        // Sort is in reverse order so newest posts are first
        return posts[i].Date.After(posts[j].Date)
    }
    sort.Slice(posts,dates)
    result = append(result, MakePageHeader(0)...)
    result = append(result, []byte(`
<div class="wblock">
    <div style="color:white; opacity:1; margin-top:1rem; margin-bottom:1rem">
    <h1>Avoiding The Needless Multiplication Of Forms</h1>
    </div>
</div>
<div class="wblock">
    <h2>Posts By Date</h2>
</div>
<div class="pblock"><div class="pinner">
<table>
`)...)
    for _,p := range posts {
        result = append(result, []byte("<tr><td>")...)
        result = append(result, p.Date.Format("2006 Jan _2 Mon")...)
        result = append(result, []byte("</td>")...)
        result = append(result, []byte("<td>")...)
        result = append(result, p.Title...)
        result = append(result, []byte("</td>")...)
        result = append(result, []byte("<td><a href=\"")...)
        result = append(result, p.Filename...)
        result = append(result, []byte("\">")...)
        result = append(result, p.Subtitle...)
        result = append(result, []byte("</a></td></tr>")...)
    }
    result = append(result, []byte(`
</table>
</div></div>
<div class="wblock">
    <h2>Posts By Series</h2>
</div>
<div class="pblock"><div class="pinner">
`)...)
    bySeries := make( map[ string ] []Post )
    for _,p := range posts {
        pStr := string(p.Title)
        if bySeries[pStr] == nil {
            bySeries[pStr] = make([]Post,0)
        }
        bySeries[pStr] = append( bySeries[pStr], p)
    }
    keys := make( []string, len(bySeries) )
    i := 0
    for k := range bySeries {
        keys[i] = k
        i++
    }
    byStrings := func(i,j int) bool {
        return keys[i] < keys[j]
    }
    sort.Slice(keys,byStrings)
    for _,k := range keys {
        result = append(result, []byte("<h3>")...)
        result = append(result, []byte(k)...)
        result = append(result, []byte("</h3><table>")...)
        l := len(bySeries[k])-1
        for i,_ := range bySeries[k] {
            p := bySeries[k][l-i]
            result = append(result, []byte("<tr><td>")...)
            result = append(result, p.Date.Format("2006 Jan _2 Mon")...)
            result = append(result, []byte("</td>")...)
            result = append(result, []byte("<td><a href=\"")...)
            result = append(result, p.Filename...)
            result = append(result, []byte("\">")...)
            result = append(result, p.Subtitle...)
            result = append(result, []byte("</a></td></tr>")...)
        }
        result = append(result, []byte("</table>")...)
    }

    result = append(result, []byte(`</div></body></html>`)...)
    return result
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
        case "/index.html":
            posts := ScanPosts()
            out.Write( renderIndex(posts) )
        case "/styles.css", "/graymaster2.jpg",
             "/Basic-Regular.ttf", "/Inconsolata-Regular.ttf",
             "/SourceSansPro-Regular.otf",
             "/ArbutusSlab-Regular.ttf",  "/Rasa-Medium.ttf",        "/Yrsa-Medium.ttf",
             "/FanwoodText-Regular.ttf",  "/SpectralSC-Medium.ttf":
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
                                out.Write( RenderHtml(blocks) )
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
    http.HandleFunc("/", handler)
    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        fmt.Printf("Error creating server: %s\n", err)
    }
}

package main

import (
    "fmt"
    "os"
    "path"
    "strings"
    "time"
    "io/ioutil"

    "rst"
)

type Post struct {
    Title       []byte
    URL         string
    Tags        []byte
    Date        time.Time
    FileMod     time.Time
    FileSize    int64
    Subtitle    []byte
    Draft       bool
    Body        []byte
    Comments    []Comment
}

var cache map[string]*Post

func ScanPosts() {
    t1 := time.Now()
    files, err := ioutil.ReadDir("data")
    if err!=nil {
        fmt.Printf("Error scanning data: %s\n", err.Error())
        return
    }
    for _,entry := range files {
        if path.Ext(entry.Name())==".rst" {
            name := strings.TrimSuffix(entry.Name(),".rst")
            Lookup(name)
         }
    }
    t2 := time.Now()
    fmt.Printf("%29s: Request serviced in %.1fms\n","ScanPosts",float64(t2.Sub(t1))/float64(time.Millisecond))
}


func Lookup(name string) (*Post,error) {
    filename  := "data/"+name+".rst"
    mdata,err := os.Stat(filename)
    if err!=nil {
        fmt.Printf("Error, cannot stat %s: %s\n", filename, err.Error())
        return nil,err
    }

    post,present := cache[name]
    // Check if the post is cached, if not create a placeholder.
    if !present {
        post = &Post{URL:"/awmblog/"+name+"/index.html", Draft:true, Comments:make([]Comment,0,4)}
        cache[name] = post
    }

    // Check if the post in the cache is up-to-date, rescan if not.
    if post.FileMod!=mdata.ModTime() || post.FileSize!=mdata.Size() {
        fmt.Printf("Cache detected change: reparsing %s\n",filename)
        lines  := rst.LineScannerPath(filename)
        if lines!=nil {
            blocks := rst.Parse(*lines)
            headBlock := <-blocks
            pTime,err := time.Parse("2006-01-02",string(headBlock.Date))
            if err!=nil {
                pTime = time.Now()       // Push "Draft" posts to top
            } else {
                post.Draft = false
            }
            post.Title    = headBlock.Title
            post.Date     = pTime
            post.Tags     = headBlock.Tags
            post.Subtitle = headBlock.Subtitle
            post.Body     = make([]byte,0,16384)
            post.Body     = append(post.Body, renderHeading(headBlock)...)
            post.Body     = append(post.Body, renderHtml(blocks)...)
            fmt.Printf("Cached %d bytes in %s body\n", len(post.Body), filename)
        } else {
            fmt.Printf("Error in the line scanner?\n")
        }
    }

    // Regardless of path to this point, mark cached data as up-to-date.
    post.FileMod  = mdata.ModTime()
    post.FileSize = mdata.Size()
    return post, nil
}

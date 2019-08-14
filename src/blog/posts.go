package main

import (
    "fmt"
    "encoding/json"
    "os"
    "path/filepath"
    "sort"
    "strconv"
    "strings"
    "sync"
    "time"
    "io/ioutil"

    "rst"
)

type Post struct {
    Title       []byte
    URL         string
    Key         string
    Tags        []byte
    Date        time.Time
    FileMod     time.Time
    FileSize    int64
    Subtitle    []byte
    Draft       bool
    Body        []byte
    Lock        sync.RWMutex
    Comments    []Comment
}

var cache map[string]*Post
var cacheLock sync.RWMutex

func ScanPosts() {
    t1 := time.Now()
    files, err := ioutil.ReadDir("data")
    if err!=nil {
        fmt.Printf("Error scanning data: %s\n", err.Error())
        return
    }
    for _,entry := range files {
        if filepath.Ext(entry.Name())==".rst" {
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

    cacheLock.RLock()
    post,present := cache[name]
    cacheLock.RUnlock()
    // Check if the post is cached, if not create a placeholder. We do not put it into the
    // cache index until after we have updated it so that we don't need a lock for the 
    // index.
    if !present {
        post = &Post{URL:"/awmblog/"+name+"/index.html", Key:name, Draft:true, Comments:make([]Comment,0,4)}
    }

    // Check if the post in the cache is up-to-date, rescan if not.
    if post.FileMod!=mdata.ModTime() || post.FileSize!=mdata.Size() {
        fmt.Printf("Cache detected change: reparsing %s\n",filename)
        post.Lock.Lock()
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
        post.Lock.Unlock()
    }

    // Check if there are comment stores for the post and load the most recent
    files, err := ioutil.ReadDir("var/run/blog")
    if err==nil  {
        post.Lock.Lock()
        indices := make([]int,0,128)
        for _, entry := range files {
            ext      := filepath.Ext(entry.Name())
            base     := strings.TrimSuffix(entry.Name(),ext)
            idx, err := strconv.Atoi(ext[1:])
            if base==name && err==nil { indices = append(indices, idx) }
        }
        sort.Ints(indices)
        fmt.Printf("Comment backing indices: %v\n",indices)
        valid := len(indices)-1
        for valid>=0 {
            commentFn := fmt.Sprintf("var/run/blog/%s.%d", name, indices[valid] )
            js,err    := ioutil.ReadFile(commentFn)
            err = json.Unmarshal(js,&post.Comments)
            if err!=nil {
                fmt.Printf("Can't load comments: %s\n", err.Error())
            } else {
                break
            }
        }
        for old:=0; old<valid; old+=1 {
            commentFn := fmt.Sprintf("var/run/blog/%s.%d", name, indices[old] )
            fmt.Printf("Removing stale %s\n",commentFn)
            os.Remove(commentFn)
        }
        post.Lock.Unlock()
    } else {
        fmt.Printf("Can't locate comment backing files: %s\n", err.Error() )
    }

    // Regardless of path to this point, mark cached data as up-to-date.
    post.FileMod  = mdata.ModTime()
    post.FileSize = mdata.Size()
    if !present {               // If this is a temporary placeholder then insert it
        cacheLock.Lock()
        cache[name] = post
        cacheLock.Unlock()
    }

    return post, nil
}

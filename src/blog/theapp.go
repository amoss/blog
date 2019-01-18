package main

import (
    "bytes"
    "context"
    "crypto/sha256"
    "crypto/hmac"
    "crypto/rand"
    "encoding/base64"
    "encoding/json"
    "errors"
    "hash"
    "fmt"
    "io/ioutil"
    "os"
    "net/http"
    "net/url"
    "path"
    "runtime/debug"
    "sort"
    "strings"
    "time"

    "rst"
    "golang.org/x/oauth2"
)


var mimeTypes = map[string]string {
    ".jpg":  "image/jpg",
    ".webm": "video/webm",
    ".mov":  "video/quicktime",
    ".svg":  "image/svg+xml",
    ".png":  "image/png",
}

type Post struct {
    Title       []byte
    Filename    []byte
    Tags        []byte
    Date        time.Time
    FileMod     time.Time
    FileSize    int64
    Subtitle    []byte
    Draft       bool
}

var cache map[string]Post

func ScanPosts() {
    t1 := time.Now()
    files, err := ioutil.ReadDir("data")
    if err!=nil {
        fmt.Printf("Error scanning data: %s\n", err.Error())
        return
    }
    for _,entry := range files {
        if path.Ext(entry.Name())==".rst" {
            mdata,err := os.Stat("data/"+entry.Name())
            if err!=nil {
                fmt.Printf("Error calling stat: %s\n", err.Error())
            } else {
                post,present := cache[entry.Name()]
                // Check if the post is cached, if not create a placeholder.
                if !present {
                    bName := strings.TrimSuffix( entry.Name(), path.Ext(entry.Name()) )
                    linkName := []byte( bName + "/index.html" )
                    post = Post{Filename:linkName, Draft:true}
                    cache[entry.Name()] = post
                }
                // Check if the post in the cache is up-to-date, rescan if not.
                if post.FileMod!=mdata.ModTime() || post.FileSize!=mdata.Size() {
                    fmt.Printf("Cache detected change: reparsing %s\n",entry.Name())
                    lines  := rst.LineScanner("data/"+entry.Name())
                    if lines!=nil {
                        blocks := rst.Parse(*lines)
                        headBlock := <-blocks
                        for x := range blocks { x = x}
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
                    } else {
                        fmt.Printf("Error in the line scanner?\n")
                    }
                }
                // Regardless of path to this point, mark cached data as up-to-date.
                post.FileMod  = mdata.ModTime()
                post.FileSize = mdata.Size()
                cache[entry.Name()] = post
            }
        }
    }
    t2 := time.Now()
    fmt.Printf("%29s: Request serviced in %.1fms\n","ScanPosts",float64(t2.Sub(t1))/float64(time.Millisecond))
}

func renderIndex(posts []Post, levelsDeep int, showDrafts bool) []byte {
    result := make([]byte, 0, 16384)
    dates :=  func(i,j int) bool {
        // Sort is in reverse order so newest posts are first
        return posts[i].Date.After(posts[j].Date)
    }
    sort.Slice(posts,dates)
    result = append(result, MakePageHeader(levelsDeep)...)
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

func privateHandler(out http.ResponseWriter, req *http.Request) {
  if req.Header["Eppn"][0] != "awm@bth.se" {
      http.Error(out, errors.New("There is a charm about the forbidden that makes it unspeakably desirable - Mark Twain.").Error(),
                 http.StatusForbidden)
      return
  }
  if req.URL.Path=="/private/index.html" {
      ScanPosts()
      posts := make([]Post,0,len(cache))
      for _,p := range cache {
          posts = append(posts,p)
      }
      out.Write( renderIndex(posts,1,true) )
      return
  }
  commonHandler(out,req,true)
}

func publicHandler(out http.ResponseWriter, req *http.Request) {
  if req.URL.Path=="/index.html" {
      ScanPosts()
      posts := make([]Post,0,len(cache))
      for _,p := range cache {
          if !p.Draft {
              posts = append(posts,p)
          }
      }
      out.Write( renderIndex(posts,0,false) )
      return
  }
  commonHandler(out,req,false)
}

func commonHandler(out http.ResponseWriter, req *http.Request, showDrafts bool) {
    if req.URL.Path[ len(req.URL.Path)-1 ] == '/' {
        http.Redirect(out,req,req.Header["X-Forwarded-Url"][0]+"index.html",302)
        return
    }

var reqPath string
    if showDrafts {
      reqPath = req.URL.Path[8:]     // Eat "/private/" -> "/"
    } else {
      reqPath = req.URL.Path
    }

    mime,ok := mimeTypes[ path.Ext(reqPath) ]
    if ok {
        filename := "data" + reqPath
        cnt,err := ioutil.ReadFile(filename)
        if err!=nil && os.IsNotExist(err) {
            http.Error(out, errors.New("File not found").Error(), http.StatusNotFound)
            fmt.Printf("%29s: Missing file %s\n", "handler", filename)
        } else {
            out.Header().Set("Content-type", mime)
            fmt.Printf("%29s: Recognised extension %s, served %s\n",
                       "handler", path.Ext(req.URL.Path), mime)
            out.Write(cnt)
        }

    } else if path.Ext(reqPath)==".html" {

        if strings.HasSuffix(req.URL.Path,"index.html") {
            dir := path.Dir(reqPath)
            filename := "data" + dir + ".rst"
            outsideFI, outsideErr := os.Stat(filename)
            //fmt.Printf("%s: %s, %s\n", outside, outsideFI, outsideErr)
            if outsideFI==nil {
                out.WriteHeader(404)
                fmt.Printf("%29s: Can't resolve %s because %s\n", "handler", filename, outsideErr)
                out.Write( []byte("File not found") )
                return
            }
            lines  := rst.LineScanner(filename)
            if lines!=nil {
                fmt.Printf("%29s: Path default - served from %s\n", "handler", filename)
                blocks := rst.Parse(*lines)
                out.Write( RenderHtml(blocks,showDrafts) )
            } else {
                fmt.Printf("%29s: File not found AFTER check! %s\n", "handler", filename)
                http.Error(out, errors.New("File not found").Error(), http.StatusNotFound)
            }
        }

    } else {

        http.Error(out, errors.New("File not found").Error(), http.StatusNotFound)
        fmt.Printf("%29s: File with unknown extension %s\n", "handler", reqPath)
    }
}

func wrapper(handler http.Handler) http.Handler {
  return http.HandlerFunc( func(out http.ResponseWriter, req *http.Request) {
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
    handler.ServeHTTP(out,req)
    t2 := time.Now()
    fmt.Printf("%29s: Request serviced in %.1fms\n","wrapper",float64(t2.Sub(t))/float64(time.Millisecond))
  })
}


func staticHandler(out http.ResponseWriter, req *http.Request) {
  target := "data" + req.URL.Path
  fmt.Printf("%29s: Path whitelisted - served from %s\n", "handler", target)
  cnt,_ := ioutil.ReadFile(target)
  /* No svg in the blog whitelist
  switch path.Ext(req.URL.Path) {
      case ".svg": out.Header().Set("Content-type", "image/svg+xml")
  } */
  out.Write(cnt)
}

func secureHandler(out http.ResponseWriter, req *http.Request) {
    token,err := req.Cookie("login")
    if err==http.ErrNoCookie {
        target := fmt.Sprintf("../login.html?from=%s",req.URL.Path[1:])  // Lead leading slash
        // The golang Redirect is based off an obsolete RFC so it forces the URL to absolute
        out.Header().Set("Location",target)
        out.WriteHeader(http.StatusFound)
        //http.Redirect(out, req, target, http.StatusFound)
        return
    } else if err!=nil {
        http.Error(out, errors.New("Something went wrong :(").Error(),
                                   http.StatusInternalServerError)
        return
    }

    loginKey,ok := checkMac(token.Value)
    s,found := sessions[string(loginKey)]

    if !ok || !found {
        fmt.Printf("Bad session cookie: %s -> %s (%s,%s)\n", token.Value, loginKey, ok, found)
        target := fmt.Sprintf("../login.html?from=%s",req.URL.Path[1:])  // Lead leading slash
        // The golang Redirect is based off an obsolete RFC so it forces the URL to absolute
        out.Header().Set("Location",target)
        out.WriteHeader(http.StatusFound)
        //http.Redirect(out, req, target, http.StatusFound)
        return

    } else {
        fmt.Fprintf(out, "In you are, %s from %s\n",s.user.Name,s.provider)
    }

}


var providers = map[string]oauth2.Config{
    "google": oauth2.Config{
            ClientID:     "CENSORED",
            ClientSecret: "CENSORED",
            Endpoint:    oauth2.Endpoint{AuthURL:"https://accounts.google.com/o/oauth2/v2/auth",
                                         TokenURL:"https://oauth2.googleapis.com/token"},
            RedirectURL:  "https://mechani.se/awmblog/callback",
            Scopes:       []string{"openid", "profile", "email" }}}

var userInfos = map[string]string {
    "google": "https://openidconnect.googleapis.com/v1/userinfo" }

func authHandler(out http.ResponseWriter, req *http.Request) {
    provName := req.URL.Query().Get("provider")
    config,found := providers[provName]
    if !found {
        http.Error(out, "Who the fuck is that?!?", http.StatusInternalServerError)
        return
    }
    referer := req.Header.Get("Referer")
    refUrl,err := url.Parse(referer)
    if len(referer)==0  ||  err!=nil {
        http.Error(out, "Referer was made of hairy bollocks", http.StatusInternalServerError)
        return
    }
    original := refUrl.Query().Get("from")
    stateData := fmt.Sprintf("%s|%s",provName,original)
    encState := msgMac(stateData)
    http.Redirect(out, req, config.AuthCodeURL(encState), http.StatusFound)
}


func msgMac(msg string) string {
    stateHmac.Reset()
    stateHmac.Write([]byte(msg))
    mac := stateHmac.Sum(nil)
    state := fmt.Sprintf("%s|%s",msg,mac)
    return base64.StdEncoding.EncodeToString([]byte(state))
}

func checkMac(mac string) ([]byte, bool) {
    fmt.Printf("checkMac: %s\n",mac)
    raw,err := base64.StdEncoding.DecodeString(mac)
    if err!=nil {
        return nil, false
    }
    split   := bytes.LastIndexByte(raw,'|')
    msg     := raw[:split]
    oldSig  := raw[split+1:]

    fmt.Printf("msg: %s\n",msg)

    stateHmac.Reset()
    stateHmac.Write([]byte(msg))
    newSig := stateHmac.Sum(nil)

    return msg, hmac.Equal(oldSig,newSig)
}

func doRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
        client := http.DefaultClient
        if c, ok := ctx.Value(oauth2.HTTPClient).(*http.Client); ok {
                client = c
        }
        return client.Do(req.WithContext(ctx))
}

type UserInfo struct {
    Sub     string
    Profile string
    Email   string
    Name    string
}

type Session struct {
    user     UserInfo
    token   *oauth2.Token
    provider string
}

var sessions map[string] Session = make(map[string]Session)
var hmacKey []byte
var stateHmac hash.Hash

func callbackHandler( out http.ResponseWriter, req *http.Request) {
    var provider, original string
    ctx := context.Background()
    state,ok := checkMac( req.URL.Query().Get("state") )
    if ok {
        decodeParts := bytes.Split(state,[]byte("|"))
        provider = string(decodeParts[0])
        original = string(decodeParts[1])
    } else {
        http.Error(out, "State is always the problem %s", http.StatusBadRequest)
        return
    }
    config := providers[provider]
    fmt.Printf("Config: %s\n",config)
    fmt.Printf("Code: %s\n",req.URL.Query().Get("code"))
    oauth2Token,err := config.Exchange(ctx, req.URL.Query().Get("code"))
    if err!=nil {
        http.Error(out, fmt.Sprintf("Provider cannot exchange code: %s",err.Error()), 
                        http.StatusBadRequest)
        return
    }
    tokenSrc := oauth2.StaticTokenSource(oauth2Token)
    token,err := tokenSrc.Token()
    if err!=nil {
        http.Error(out, "Token problem", http.StatusInternalServerError)
        return
    }
    uiReq,err := http.NewRequest("GET", userInfos[provider],nil)
    if err!=nil {
        http.Error(out, "No request for userinfo", http.StatusInternalServerError)
        return
    }
    token.SetAuthHeader(uiReq)
    fmt.Printf("Req: %s\n",uiReq)
    resp, err := doRequest(ctx, uiReq)
    if err!=nil {
        http.Error(out, fmt.Sprintf("Token request failed: %s",err.Error()),
                        http.StatusInternalServerError)
        return
    }
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    if err!=nil {
        http.Error(out, "Can't read Token request body", http.StatusInternalServerError)
        return
    }
    if resp.StatusCode != http.StatusOK {
        http.Error(out, fmt.Sprintf("Token exchange failed %s: %s", resp.Status, body),
                        http.StatusInternalServerError)
    }

    fmt.Printf("Third leg response: %s\n", body)

    var userInfo UserInfo
    if err := json.Unmarshal(body, &userInfo); err != nil {
        http.Error(out, fmt.Sprintf("Userinfo decode failed: %v", err),
                        http.StatusInternalServerError)
        return
    }

    loginKey := fmt.Sprintf("%s|%s", provider, userInfo.Sub)
    encLogin := msgMac(loginKey)

    http.SetCookie(out, &http.Cookie{Name:"login",
                                     Value:encLogin,
                                     Expires:time.Now().Add(time.Minute*10)})

    sessions[loginKey] = Session{user:userInfo,token:oauth2Token,provider:provider}
    http.Redirect(out, req, original, http.StatusFound)
}

//func loginHandler(out http.ResponseWriter, req *http.Request) {
//}

var whitelist = []string{ "/styles.css", "/graymaster2.jpg", "/Basic-Regular.ttf",
                "/Inconsolata-Regular.ttf", "/SourceSansPro-Regular.otf",
                "/ArbutusSlab-Regular.ttf", "/Rasa-Medium.ttf", "/Yrsa-Medium.ttf",
                "/FanwoodText-Regular.ttf",  "/SpectralSC-Medium.ttf", "/Rasa-Regular.ttf",
                "/login.html"}
func main() {
    cache = make(map[string]Post)
    hmacKey = make([]byte,32)
    _,err := rand.Read(hmacKey)
    if err!=nil {
        fmt.Printf("Can't initialise the random hmac key! %s\n", err.Error())
        return
    }
    stateHmac = hmac.New(sha256.New,hmacKey)

    http.Handle("/",         wrapper(http.HandlerFunc(publicHandler)))
    http.Handle("/private/", wrapper(http.HandlerFunc(privateHandler)))
    http.Handle("/secure/",  wrapper(http.HandlerFunc(secureHandler)))
    http.Handle("/auth",     wrapper(http.HandlerFunc(authHandler)))
    http.Handle("/callback", wrapper(http.HandlerFunc(callbackHandler)))
//    http.Handle("/login", http.HandlerFunc(loginHandler))
    for _,p := range whitelist {
      http.Handle(p, wrapper(http.HandlerFunc(staticHandler)))
    }
    err = http.ListenAndServe(":8080", nil)
    if err != nil {
        fmt.Printf("Error creating server: %s\n", err)
    }
}

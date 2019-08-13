package main

import (
    "bytes"
    "context"
    "crypto/sha256"
    "crypto/hmac"
    "crypto/rand"
    //"encoding/base64"
    "encoding/json"
    "errors"
    "hash"
    "fmt"
    "io/ioutil"
    "os"
    "net/http"
    "net/url"
    "path"
    "path/filepath"
    "runtime/debug"
    "sort"
    "strings"
    "time"

    "rst"
    "github.com/gorilla/websocket"
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
                    lines  := rst.LineScannerPath("data/"+entry.Name())
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

func renderIndex(posts []Post, levelsDeep int, showDrafts bool, sessionBar []byte) []byte {
    result := make([]byte, 0, 16384)
    dates :=  func(i,j int) bool {
        // Sort is in reverse order so newest posts are first
        return posts[i].Date.After(posts[j].Date)
    }
    sort.Slice(posts,dates)
    result = append(result, PageHeader...)
    result = append(result, sessionBar...)
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


func publicHandler(out http.ResponseWriter, req *http.Request) {
    session, sessionBar := Find(req)

    showDrafts := false
    fmt.Printf("Session: %s / %s\n", session.Name, session.provider)
    if session.provider=="local" && session.Name=="amoss" { showDrafts = true }

    if req.URL.Path=="/awmblog/index.html" {
        ScanPosts()
        posts := make([]Post,0,len(cache))
        for _,p := range cache {
            if showDrafts || !p.Draft {
                posts = append(posts,p)
            }
        }
        out.Write( renderIndex(posts,0,showDrafts,sessionBar) )
        return
    }
    commonHandler(out,req,showDrafts,session,sessionBar)
}

func commonHandler(out http.ResponseWriter, req *http.Request, showDrafts bool, session *Session, sessionBar []byte) {
    if req.URL.Path[ len(req.URL.Path)-1 ] == '/' {
        http.Redirect(out,req,req.URL.Path+"index.html",302)
        return
    }

var reqPath string
    reqPath = req.URL.Path
    if strings.HasPrefix(reqPath,"/awmblog") {
        reqPath = reqPath[8:]
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
            lines  := rst.LineScannerPath(filename)
            if lines!=nil {
                fmt.Printf("%29s: Path default - served from %s\n", "handler", filename)
                blocks := rst.Parse(*lines)
                out.Write( RenderPage(blocks,showDrafts,sessionBar))
                out.Write( []byte(`<div class="wblock"><h1>Comments</h1></div>`) )
                out.Write( CommentDemo )
                if session.provider=="none" {
                    out.Write([]byte(`<div class="wblock"><p>Sign in at the top of the page to leave a comment</p></div>`))
                } else {
                    out.Write( CommentEditor(session) )
                }
                out.Write( PageFooter )
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


/* Logging request, profiling handler time and catching panics.
*/
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


/* Whitelisting to prevent attacks that escape the data/ directory.
*/
func staticHandler(out http.ResponseWriter, req *http.Request) {
  target := "data/" + filepath.Base(req.URL.Path)
  fmt.Printf("%29s: Path whitelisted - served from %s\n", "handler", target)
  cnt,_ := ioutil.ReadFile(target)
  switch path.Ext(req.URL.Path) {
      case ".css": out.Header().Set("Content-type", "text/css")
      case ".js" : out.Header().Set("Content-type", "application/javascript")
  }
  out.Write(cnt)
}


func cuteDate(d time.Time) string {
    now := time.Now()
    dt  := now.Sub(d)
    if dt.Hours() < 20 {
        return d.Format("15:04")
    } else if dt.Hours() < 6*24 {
        return d.Format("Mon 15:04")
    } else if now.Year() == d.Year() {
        return d.Format("Mon Jan 2 (15:04)")
    } else {
        return d.Format("2006 Jan 2 (15:04)")
    }
}

type WsMsg struct {
    Action string
    Body   string
}

func reader(conn *websocket.Conn) {
    _, p, err := conn.ReadMessage()
    fmt.Printf("Initial message: %s\n", p)
    if err != nil {
            fmt.Println(err)
            return
        }
    cookie := string(p)
    if cookie[:6]!="login=" {
        fmt.Printf("ws did not authenticate! %s\n", cookie[:6])
        conn.Close(); 
        return 
    }
    loginKey,ok := checkMac( cookie[6:] )
    if !ok {
        fmt.Printf("MAC check failed\n")
        conn.Close()
        return
    }

    s,ok := sessions[ string(loginKey) ]
    if !ok {
        fmt.Printf("Session not found? %s\n", loginKey);
        conn.Close();
        return
    }

    for {
    // read in a message
        _, p, err := conn.ReadMessage()
        if err != nil {
            fmt.Println(err)
            return
        }
        msg := WsMsg{}
        json.Unmarshal(p,&msg)
        fmt.Printf("%s-->%s\n",p,msg)
        if msg.Action=="preview" {
            lines  := rst.LineScannerBytes([]byte(msg.Body))
            if lines!=nil {
                blocks := rst.Parse(*lines)
                html := make([]byte,4096)
                html = append(html, []byte(fmt.Sprintf(`<div class="wblock"><h2>%s/%s at %s commented:</h2></div>`, s.Name, s.provider, cuteDate(time.Now())))...)
                html = append(html,renderHtml(blocks)...)
                //fmt.Printf("%s: %s\n", messageType, html)
                conn.WriteMessage(1, html)
            }
        }
    }
}



func wsHandler(out http.ResponseWriter, req *http.Request) {
    fmt.Printf("Incoming ws\n")
    upgrader := websocket.Upgrader{ ReadBufferSize: 1024, WriteBufferSize: 1024 }
    upgrader.CheckOrigin = func(r *http.Request) bool { return true }
    ws, err  := upgrader.Upgrade(out, req, nil)
    if err!=nil {
        fmt.Printf("Error on ws upgrade: %s\n", err.Error())
        return
    }
    reader(ws)

}



var providers = map[string]*oauth2.Config{
    "google": &oauth2.Config{
            ClientID:     "CENSORED",
            ClientSecret: "CENSORED",
            Endpoint:     oauth2.Endpoint{
                            AuthURL:  "https://accounts.google.com/o/oauth2/auth",
                            TokenURL: "https://accounts.google.com/o/oauth2/token"},
            RedirectURL:  "https://mechani.se/awmblog/callback",
            Scopes:       []string{"openid", "profile", "email" }},
    "twitter": &oauth2.Config{
            
    }}


var userInfos = map[string]string {
    "google": "https://openidconnect.googleapis.com/v1/userinfo" }

func authHandler(out http.ResponseWriter, req *http.Request) {
    provName := req.URL.Query().Get("provider")
    if provName=="local" {
        if err := req.ParseForm(); err!=nil {
            http.Error(out, "Quit fucking around", http.StatusInternalServerError)
            return
        }
        user := req.FormValue("user")
        pass := req.FormValue("password")
        original := req.FormValue("from")
        if user=="amoss" && pass=="CENSORED" {
            loginKey := fmt.Sprintf("local|1")
            encLogin := msgMac(loginKey)

            http.SetCookie(out, &http.Cookie{Name:"login",
                                             Value:encLogin,
                                             Expires:time.Now().Add(time.Minute*60)})

            sessions[loginKey] = &Session{Name:user,Profile:"",Email:"",Sub:"1",provider:"local"}
            fmt.Println("Create local session: %s",sessions[loginKey])
            http.Redirect(out, req, original, http.StatusFound)
        }
    }
    config,found := providers[provName]        // Do they hide state in here?
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
    original := refUrl.Path
    stateData := fmt.Sprintf("%s|%s",provName,original)
    encState := msgMac(stateData)
    http.Redirect(out, req, config.AuthCodeURL(encState), http.StatusFound)
}

func doRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
        client := http.DefaultClient
        if c, ok := ctx.Value(oauth2.HTTPClient).(*http.Client); ok {
                client = c
        }
        return client.Do(req.WithContext(ctx))
}

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
        http.Error(out, "State is always the problem", http.StatusBadRequest)
        return
    }
    config := providers[provider]      // Do they hide state in here?
    fmt.Printf("State: %s\n",state)
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
                                     Expires:time.Now().Add(time.Minute*60)})

    sessions[loginKey] = &Session{Name:userInfo.Name,Profile:userInfo.Profile,Email:userInfo.Email,Sub:userInfo.Sub,token:oauth2Token,provider:provider}
    fmt.Println("Create session: %s",sessions[loginKey])
    http.Redirect(out, req, original, http.StatusFound)
}

func logoutHandler(out http.ResponseWriter, req *http.Request) {
    http.SetCookie(out, &http.Cookie{Name:"login",
                                     Value:"",
                                     Expires:time.Now().Add(-time.Minute*24*60)})
    referer := req.Header.Get("Referer")
    refUrl,err := url.Parse(referer)
    if len(referer)==0  ||  err!=nil {    
        http.Error(out, "Referer was made of hairy bollocks", http.StatusInternalServerError)
        return
    }
    original := refUrl.Path
    http.Redirect(out, req, original, http.StatusFound)
}

var whitelist = []string{ "/awmblog/styles.css",    "/awmblog/graymaster2.jpg", "/awmblog/Basic-Regular.ttf",
                "/awmblog/Inconsolata-Regular.ttf", "/awmblog/SourceSansPro-Regular.otf",
                "/awmblog/ArbutusSlab-Regular.ttf", "/awmblog/Rasa-Medium.ttf", "/awmblog/Yrsa-Medium.ttf",
                "/awmblog/FanwoodText-Regular.ttf", "/awmblog/SpectralSC-Medium.ttf", "/awmblog/Rasa-Regular.ttf",
                "/awmblog/login.html",              "/awmblog/comments.js" }
func main() {
    cache = make(map[string]Post)
    hmacKey = make([]byte,32)
    _,err := rand.Read(hmacKey)
    if err!=nil {
        fmt.Printf("Can't initialise the random hmac key! %s\n", err.Error())
        return
    }
    stateHmac = hmac.New(sha256.New,hmacKey)

    http.Handle("/awmblog/",           wrapper(http.HandlerFunc(publicHandler)))
    http.Handle("/awmblog/auth",       wrapper(http.HandlerFunc(authHandler)))
    http.Handle("/awmblog/callback",   wrapper(http.HandlerFunc(callbackHandler)))
    http.Handle("/awmblog/logout",     wrapper(http.HandlerFunc(logoutHandler)))
    http.Handle("/awmblog/local.html", wrapper(http.HandlerFunc(LocalHandler)))
    http.HandleFunc("/awmblog/preview",  wsHandler)

    for _,p := range whitelist {
      http.Handle(p, wrapper(http.HandlerFunc(staticHandler)))
    }
    err = http.ListenAndServe(":8080", nil)
    if err != nil {
        fmt.Printf("Error creating server: %s\n", err)
    }
}

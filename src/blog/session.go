package main

import (
    "bytes"
    "crypto/hmac"
    "encoding/base64"
    "fmt"
    "net/http"
    "net/url"
    "golang.org/x/oauth2"
)

type UserInfo struct {
    Sub     string
    Profile string
    Email   string
    Name    string
}

type Session struct {
    Sub     string
    Profile string
    Email   string
    Name    string
    token   *oauth2.Token
    provider string
}

var sessions map[string] *Session = make(map[string]*Session)

func (self *Session) GenerateBar() []byte {
    var bar []byte
    if self.Name=="guest" && self.provider=="none" {
        bar = []byte(fmt.Sprintf(`<div class="session">Login with: 
<a href="/awmblog/auth?provider=google">Google</a>
Twitter  Facebook <a href="/awmblog/local.html">Local</a></div>`))
    } else {
        bar = []byte(fmt.Sprintf("<div class=\"session\"> Logged as %s. <a href=\"/awmblog/logout\">Log out</a></div>", self.Name))
    }
    return bar
}


/* Check if the req is associated with a current session (logged in user).
   Generate the div for the session bar and return the current session.
   If the user is not logged in then generate a dummy session to mark them
   as a guest.
*/
func Find(req *http.Request) (*Session,[]byte) {
    var session *Session = nil
    token,err := req.Cookie("login")
    if err==nil {
        loginKey,ok := checkMac(token.Value)
        if ok {
            session, _ = sessions[string(loginKey)]
            if session==nil { fmt.Printf("Error? %s %s\n", session, sessions ) }
        } else {
            fmt.Printf("Login %v failed mac check\n",token.Value)
        }
    }
    if session==nil {
        session = &Session{Name:"guest",provider:"none"}
    }
    sessionBar := session.GenerateBar()
    return session, sessionBar
}


func LocalHandler(out http.ResponseWriter, req *http.Request) {
    referer := req.Header.Get("Referer")
    refUrl,err := url.Parse(referer)
    if len(referer)==0  ||  err!=nil {
        http.Error(out, "Referer was made of hairy bollocks", http.StatusInternalServerError)
        return
    }
    original := refUrl.Path
    out.Write([]byte(fmt.Sprintf(
`<html>
<body>
<h1>Local site account</h1>

<form action="auth?provider=local" method="post">
    <p>Username: <input type="text" name="user" /></p>
    <p>Password: <input type="password" name="password" /></p>
    <input type="hidden" name="from" value="%s"/>
    <input type="submit" name="submit" value="Submit" />
</form>

</body>
</html>`, original)))
}


func msgMac(msg string) string {
    stateHmac.Reset()
    stateHmac.Write([]byte(msg))
    mac := stateHmac.Sum(nil)
    state := fmt.Sprintf("%s|%s",msg,mac)
    return base64.StdEncoding.EncodeToString([]byte(state))
}

func checkMac(mac string) ([]byte, bool) {
    raw,err := base64.StdEncoding.DecodeString(mac)
    fmt.Printf("checkMac: mac %d bytes -> %d  %v/%v\n", len(mac), len(raw), mac, raw)
    if err!=nil {
        fmt.Println("checkMac failed to base64 decode state")
        return nil, false
    }
    split   := bytes.LastIndexByte(raw,'|')
    msg     := raw[:split]
    oldSig  := raw[split+1:]

    stateHmac.Reset()
    stateHmac.Write([]byte(msg))
    newSig := stateHmac.Sum(nil)

    match := hmac.Equal(oldSig,newSig)
    if !match { fmt.Printf("checkMac failed to match sig %v vs %v",oldSig,newSig) }
    return msg, match
}

package main

import (
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
    if self==nil {
        bar = []byte(fmt.Sprintf(`<div class="session">Login with: 
<a href="/awmblog/auth?provider=google">Google</a>
Twitter  Facebook <a href="/awmblog/local.html">Local</a></div>`))
    } else {
        bar = []byte(fmt.Sprintf("<div class=\"session\"> Logged as %s. <a href=\"/awmblog/logout\">Log out</a></div>", self.Name))
    }
    return bar
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

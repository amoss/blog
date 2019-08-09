package main

import (
    "fmt"
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

func (self *Session) GenerateBar(page string) []byte {
    var bar []byte
    if self==nil {
        bar = []byte(fmt.Sprintf(`<div class="session">Login with: 
<a href="/awmblog/auth?provider=google&from=%s">Google</a>
Twitter  Facebook Local</a></div>`, page))
    } else {
        bar = []byte(fmt.Sprintf("<div class=\"session\"> Logged as %s. <a href=\"\">Log out</a></div>", self.Name))
    }
    return bar
}

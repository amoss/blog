package main

import (
    "errors"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
    "net/url"
    "path/filepath"
    "strings"
    "time"

    "rst"
)

var CommentDemo = []byte(`<div class="wblock"><a href="javascript:showDemo()">Markdown syntax...</a></div>
                          <div class="wblock" id="commentDemo" style="visibility: hidden; display: none">
                <div style="display: inline-block; width:45%; float:left; background: #cccccc; white-space:pre; color: #444444; font-family: 'monospace'">
Markdown syntax is based on RST: 
*italics* **bold** are inline styles. 
Paragraphs wrap

until blank lines separate. Other inline 
styles are ` +
":code:`x=y+z`, :shell:`ls` and " + `
` + ":math:`x^n=y^n+z^n`. Links are " + `
` + "`here <example.com>`_ " + `and blocks 
require blank lines

.. code::

  x <- f(y)
  // Until indent change

* Bullets are not indented
* Listed between blank lines</div><div class="comment" style="display: inline-block; width:45%; margin-left:8%; float;right; border:1px solid #666666"><div class="pblock"><div class="pinner">Markdown syntax is based on RST: <i>italics</i> <b>bold</b> are inline styles. Paragraphs wrap</div></div><div class="pblock"><div class="pinner">until blank lines separate. Other inline styles are <span class="code">x=y+z</span> <span class="shell">ls</span>and \\((x^n=y^n+z^n\\)) Links are <a href="example.com">here</a> and blocks require blank lines</div></div><div class="rblock"><div class="code">x &lt;- f(y)
// Until indent change</div></div><div class="pblock"><div class="pinner"><ul><li>Bullets are not indented</li><li>Listed between blank lines</li></ul></div></div></div><div style="clear:both"></div></div>`)


func CommentEditor(session *Session) []byte {
    return []byte(`<div class="wblock" style="height:auto">
                <div style="display: inline-block; width:45%; float:left">
                <textarea id="comment" style="height:100%; width:100%; resize:none" oninput='commentUpdate(this)'></textarea>
                <form action="javascript:submitComment()">
                    <input type="submit" value="Add Comment" id="submitButton" />
                </form>
            </div>
            
            <div id="comPreview" class="comment" style="display: inline-block; width:45%; margin-left:8%; margin-right:0; float;right; border:1px solid #666666">
            </div><div style="clear:both"></div>
</div>`)
}
                /*<form>
                <input type="textarea" name="comment" rows="10" style="height:100%"/>
                </form>*/

type Comment struct {
    Name     string
    Provider string
    Sub      string
    Email    string
    Body     string
    When     time.Time
    Html     []byte
}

func PostComment(ustr string, s *Session, body string) error {
    u,err := url.Parse(ustr)
    fmt.Println(u)
    if err!=nil { return err }
    if !u.IsAbs() || filepath.Base(u.Path)!="index.html" || !strings.HasPrefix(u.Path,"/awmblog") {
        return errors.New("Url is absolute garbage")
    }
    postPath := u.Path[8:]                           // e.g. "awmblog/blah/index.html" -> "/blah/index.html"
    fmt.Printf("URL Path is %s\n",u.Path)
    parent := filepath.Base(filepath.Dir(postPath))  // e.g. ... -> "blah"
    filename := "data/" + parent + ".rst"             // e.g. ... -> "data/blah.rst"
    fmt.Printf("Checking %s\n", filename)
    outsideFI, _ := os.Stat(filename)
    if outsideFI==nil { return errors.New("Url is garbage") }

    newComm := Comment{ Name:s.Name, Provider:s.provider, Sub:s.Sub, Email:s.Email, Body:body, When:time.Now() }
    lines  := rst.LineScannerBytes([]byte(body))
    if lines!=nil {
        blocks := rst.Parse(*lines)
        newComm.Html = make([]byte,4096)
        newComm.Html = append(newComm.Html, []byte(fmt.Sprintf(`<div class="wblock"><h2>%s/%s at %s commented:</h2></div>`, s.Name, s.provider, cuteDate(newComm.When)))...)
        newComm.Html = append(newComm.Html,renderHtml(blocks)...)
    }

    fmt.Printf("New comment: %s\n", newComm)
    post,ok := cache[parent]
    if ok {
        fmt.Printf("Located post: %s\n", post)
    } else {
        for k := range cache {
            fmt.Printf("Not in cache: %s\n", k)
        }
    }
    post.Comments = append(post.Comments,newComm)
    err = post.Save()
    return err
}

func (self *Post) Save() error {
    files, err := ioutil.ReadDir("var/run/blog")
    if err!=nil  { return err }
    count := 0
    for _, entry := range files {
        base := strings.TrimSuffix(entry.Name(),filepath.Ext(entry.Name()))
        if base==self.Key { count += 1 }
    }
    base := fmt.Sprintf("var/run/blog/%s.%d", self.Key, count)
    fmt.Printf("New backing: %s\n",base)
    f, err := os.Create(base)
    if err!=nil  { return err }
    defer f.Close()
    encoded,err := json.Marshal(self.Comments)
    if err!=nil  { return err }
    _, err = f.Write( encoded )
    return nil
}

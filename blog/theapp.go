package main

import (
      "net/http"
      "fmt"
      "io/ioutil"
      "strings"
)

type FragStyle int
const (
  FsNone FragStyle = iota
  FsEmpty
  FsStrong
  FsDefhead
  FsDefitem
  FsLink
  FsLiteral
  FsTopicTitle
)

type DocFrag struct {
    style FragStyle
    cnt   string
}

type BlockStyle int
const (
  BsNone BlockStyle = iota
  BsBulleted
  BsNumbered
  BsQuote
  BsBeginTopic
  BsEndTopic
  BsBibItem
  BsSubsection
)

type DocBlock struct {
    style BlockStyle
    frags []DocFrag
}

type DocSection struct {
    title    string
    blocks []DocBlock
}

type Document struct {
    sections []DocSection
}

func (doc *Document) newBlock(style BlockStyle) {
  if len(doc.sections)==0 {
    doc.sections = make([]DocSection,1)
  }
  curSection := &doc.sections[ len(doc.sections)-1 ]
  curSection.blocks = append(curSection.blocks, DocBlock{style:style})
}

func (doc *Document) newFragment(style FragStyle, content string) {
    if len(doc.sections)==0 {
        doc.newBlock(BsNone)
    }
    curSection := &doc.sections[ len(doc.sections)-1 ]
    if len(curSection.blocks)==0 {
        doc.newBlock(BsNone)
    }
    curBlock := &curSection.blocks[ len(curSection.blocks)-1 ]
    curBlock.frags = append( curBlock.frags, DocFrag{style:style, cnt:content} )
    fmt.Println(curBlock,len(curBlock.frags))
}

func (doc *Document) renderHtml() {
    for s, curSection := range doc.sections {
        fmt.Println("Section",s,curSection.title)
        for b, curBlock := range(curSection.blocks) {
            fmt.Println(b,curBlock.style,len(curBlock.frags))
            for f, frag := range(curBlock.frags) {
                fmt.Println("F:", f, frag.style, frag.cnt)
            }
        }
    }
}

type LineClass int
const (
      Blank LineClass = iota
      Indented
      SectionHeading
      SubsectionHeading
      Directive
      Other
)

// Simple abstract domain for the types of lines in an .rst
func classifyLine(line string) LineClass {
  if len(line)==0 {
    return Blank
  }
  // If you are not me then don't use tabs, haha only serious
  expanded := strings.Replace(line,"\t","  ",-1)
  text := strings.TrimLeft( expanded," " )
  indent := len(expanded) - len(text)
  if indent>0 {
    return Indented
  }

  if len(strings.Trim(expanded,"="))==0 {
    return SectionHeading
  }
  if len(strings.Trim(expanded,"-"))==0 {
    return SubsectionHeading
  }
  if line[0:2]==".." {
    return Directive
  }
  return Other
}

type ParseState int
const (
        Default ParseState = iota
        TitleBlock
        InPara
        InDirective
)

func parseRst(src string) {
  var doc Document
  lines := strings.Split(src,"\n")
  state := Default
  cur := ""
  for i:= 0; i<len(lines); i++ {
    switch state {
      case Default:
        switch classifyLine(lines[i]) {
          case SectionHeading:
            state = TitleBlock
          case Blank:
            // Ignore
          case Other:
            state = InPara
            cur   = lines[i]
            doc.newFragment(FsNone, cur)
          case Directive:
            state = InDirective
            cur = lines[i][3:]
        }
      case InDirective:
        switch classifyLine(lines[i]) {
          case Indented:
            cur += "\n" + lines[i]    // Do a tab-expansion and TrimLeft
          case Blank:
            fmt.Println("Found a directive", cur)
            state = Default
            cur   = ""
        }

      case TitleBlock:
        switch classifyLine(lines[i]) {
          case Other:
            cur += "\n" + lines[i]
          case SectionHeading:
            fmt.Println("Found a title block", cur)
            state = Default
        }
      case InPara:
        switch classifyLine(lines[i]) {
          case Other:
            cur += "\n" + lines[i]
            doc.newFragment( FsNone, lines[i] )
          case Blank:
            fmt.Println("Found a para", cur)
            state = Default
            cur   = ""
          case SectionHeading:
            fmt.Println("Found a section heading", cur)
            ns := DocSection{title:cur}
            doc.sections = append( doc.sections, ns)
            cur = ""
            state = Default
          case SubsectionHeading:
            doc.newBlock(BsSubsection)
            fmt.Println("Found a subsection heading", cur)
            cur = ""
            state = Default
        }
    }
    fmt.Println(i, classifyLine(lines[i]), lines[i])
  }
  doc.renderHtml()
}

func handler(out http.ResponseWriter, req *http.Request) {
  fmt.Println("Req:", req.URL)
  filename := "data" + req.URL.Path + ".rst"
  cnt, err := ioutil.ReadFile(filename)
  if err==nil {
    parseRst( string(cnt) )
  } else {
    fmt.Println(err)
  }
}

func main() {
    http.HandleFunc("/", handler)
    http.ListenAndServe(":8080", nil)
}

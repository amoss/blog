package main

import (
      "net/http"
      "fmt"
      "io/ioutil"
      "strings"
      "regexp"
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

func (doc *Document) lastFragment() *DocFrag {
  block := doc.lastBlock()
  return &block.frags[ len(block.frags)-1 ]
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
  BsLiteral
)

type DocBlock struct {
    style           BlockStyle
    litStyle,title  string
    frags           []DocFrag
}

type DocSection struct {
    title    string
    blocks []DocBlock
}

type DocKind int
const (
    DocPage DocKind = iota
    DocSlides
)

type Document struct {
    title,author,date,courseCode,courseName    string
    kind  DocKind
    sections []DocSection
}

func (doc *Document) lastBlock() *DocBlock {
    if len(doc.sections)==0 {
        doc.newBlock(BsNone)
    }
    curSection := &doc.sections[ len(doc.sections)-1 ]
    if len(curSection.blocks)==0 {
        doc.newBlock(BsNone)
    }
    return &curSection.blocks[ len(curSection.blocks)-1 ]
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
}

// Transition function
func forceEnv(out http.ResponseWriter, current string, next string) string {
    if current==next { return next }
    if current!="none" {
        out.Write( []byte("</") )
        out.Write( []byte(current) )
        out.Write( []byte(">") )
    }
    if next!="none" {
        out.Write( []byte("<") )
        out.Write( []byte(next) )
        out.Write( []byte(">") )
    }
    return next
}

func (doc *Document) renderHtml(out http.ResponseWriter) {
    out.Write( []byte("<html><head><title>") )
    out.Write( []byte(doc.title) )
    out.Write( []byte("</title><script src=\"http://cdn.mathjax.org/mathjax/latest/MathJax.js?config=TeX-AMS-MML_HTMLorMML\" type=\"text/javascript\"></script><link href=\"styles.css\" type=\"text/css\" rel=\"stylesheet\"></link></head><body>"))
    opEnv := "none"
    for s, curSection := range doc.sections {
        // Document can start with implied (empty) section
        if len(curSection.title)>0 {
          out.Write( []byte("<h1>") )
          out.Write( []byte(curSection.title) )
          out.Write( []byte("</h1>") )
        }
        fmt.Println("Section",s,curSection.title)
        for b, curBlock := range(curSection.blocks) {
            switch curBlock.style {
                case BsNone:
                    opEnv = forceEnv(out,opEnv,"none")
                    out.Write( []byte("\n<p>") )
                    for _, frag := range(curBlock.frags) {
                        out.Write( []byte(frag.cnt) )
                        out.Write( []byte(" ") )
                    }
                    out.Write( []byte("</p>\n") )
                case BsSubsection:
                    opEnv = forceEnv(out,opEnv,"none")
                    out.Write( []byte("<h2>") )
                    for _, frag := range(curBlock.frags) {
                        out.Write( []byte(frag.cnt) )
                        out.Write( []byte(" ") )
                    }
                    out.Write( []byte("</h2>") )
                case BsLiteral:
                    opEnv = forceEnv(out,opEnv,"none")
                    out.Write( []byte("<div class=\"") )
                    out.Write( []byte(curBlock.litStyle) )
                    out.Write( []byte("\">") )
                    for _, frag := range(curBlock.frags) {
                        out.Write( []byte(frag.cnt) )
                    }
                    out.Write( []byte("</div>") )
                case BsBulleted:
                    opEnv = forceEnv(out,opEnv,"ul")
                    out.Write( []byte("<li>") )
                    for _, frag := range(curBlock.frags) {
                        out.Write( []byte(frag.cnt) )
                        out.Write( []byte(" ") )
                    }
                default:
                    fmt.Println(b,curBlock.style,len(curBlock.frags))
                    for f, frag := range(curBlock.frags) {
                        fmt.Println("F:", f, frag.style, frag.cnt)
                    }
            }
        }
    }
    out.Write( []byte("</body></html>") )
}

type LineClass int
const (
      Blank LineClass = iota
      Indented
      SectionHeading
      SubsectionHeading
      Directive
      Bulleted
      Numbered
      Attribute
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
  if line[0:2]=="* " {
    return Bulleted
  }
  slices := strings.Split(line,":")
  if len(slices)==3  &&  
     len(slices[0])==0  &&  
     !strings.Contains(slices[1]," ") {
    return Attribute
  }
  return Other
}

type ParseState int
const (
        Default ParseState = iota
        TitleBlock
        InPara
        InDirective
        InBullet
        InTopic
        InTopic2
)

var state ParseState
func parseDirective(doc *Document, name string, extra string) {
    switch name {
        case "shell":
            doc.newBlock(BsLiteral)
            doc.lastBlock().litStyle = "shell"
            state = InDirective
        case "code":
            doc.newBlock(BsLiteral)
            doc.lastBlock().litStyle = "code"
            state = InDirective
        case "topic":
            doc.newBlock(BsBeginTopic)
            doc.lastBlock().title = extra
            state = InTopic
        case "image":
            fmt.Println("Do shit with image", extra)
        case "epigraph":
            fmt.Println("Dropping some quote shit", extra)
        default:
            panic("Unknown directive "+name)
    }
}

func (b *DocBlock) processInlines() {
  // TODO: Initial block should be single fragment
  //      Find inline markers and split into multi-frags...
}

func parseRst(src string) Document {
  var doc Document
  lines := strings.Split(src,"\n")
  state = Default
  for i:= 0; i<len(lines); i++ {
    switch state {
      case Default:
        switch classifyLine(lines[i]) {
          case SectionHeading:
            state = TitleBlock
          case Blank:
            // Ignore
          case Other:
            doc.newBlock(BsNone)
            state = InPara
            doc.newFragment(FsNone, lines[i])
          case Directive:
            re := regexp.MustCompile(".. *([A-Za-z]+):: *(.*)")
            m := re.FindStringSubmatch(lines[i])
            parseDirective(&doc, m[1], m[2])
          case Bulleted:
            doc.newBlock(BsBulleted)
            doc.newFragment(FsNone, lines[i][2:])
            state = InBullet
          case Attribute:
            if len(doc.sections)==0 {
                slices := strings.Split(lines[i],":")
                switch slices[1] {
                  case "Author":     doc.author     = slices[2]
                  case "Date":       doc.date       = slices[2]
                  case "CourseCode": doc.courseCode = slices[2]
                  case "CourseName": doc.courseCode = slices[2]
                  case "Style":
                      switch strings.TrimLeft(slices[2]," ") {
                          case "Slides": doc.kind = DocSlides
                          case "Page":   doc.kind = DocPage
                          default: panic("Unknown document style:"+slices[2])
                      }
                  default:
                    panic("Unknown document attribute:"+slices[1])
                }
            }
          default:
            fmt.Println("Dropping in ",state, lines[i])
        }
      case InTopic:
        switch classifyLine(lines[i]) {
          case Blank:
            state = InTopic2
          case Indented:
            doc.newFragment(FsNone, lines[i])
          case Other:
            panic("Need a blank after topic block")
          default:
        }
      case InTopic2:
        switch classifyLine(lines[i]) {
          case Blank:
          case Indented:
            doc.newBlock(BsNone)
            doc.newFragment(FsNone, lines[i])
            state = InTopic
          case Other:
            doc.newBlock(BsEndTopic)
            doc.newBlock(BsNone)
            doc.newFragment(FsNone, lines[i])
            state = Default
          default:
        }
      case InBullet:
        switch classifyLine(lines[i]) {
            case Blank:
                state = Default
            case Indented:
                doc.newFragment(FsNone, lines[i][2:])
            case Bulleted:
                doc.newBlock(BsBulleted)
                doc.newFragment(FsNone, lines[i][2:])
            default:
                fmt.Println("Dropping in InBullet", lines[i])
        }
      case InDirective:
        switch classifyLine(lines[i]) {
          case Indented:
            if len(doc.lastBlock().frags) == 0 {
              doc.newFragment(FsNone, lines[i])
            } else {
              doc.lastFragment().cnt = doc.lastFragment().cnt + "\n" + lines[i]
            }
          case Other:
            doc.newBlock(BsNone)
            doc.newFragment(FsNone,lines[i])
            state = InPara
          case Bulleted:
            doc.newBlock(BsBulleted)
            doc.newFragment(FsNone, lines[i][2:])
            state = InBullet
          case Blank:   // Drop
          /*case Blank:
            fmt.Println("Found a directive", cur)
            state = Default
            cur   = ""*/
          default:
            fmt.Println("Dropping in ",state, lines[i])
        }

      case TitleBlock:
        switch classifyLine(lines[i]) {
          case Other:
            doc.title = doc.title + lines[i]
          case SectionHeading:
            state = Default
          default:
            fmt.Println("Dropping in ",state, lines[i])
        }
      case InPara:
        switch classifyLine(lines[i]) {
          case Other:
            doc.lastFragment().cnt = doc.lastFragment().cnt + lines[i]
          case Blank:
            state = Default
          case SectionHeading:
            oldBlock := doc.popBlock()
            ns := DocSection{title:oldBlock.flatten()}
            doc.sections = append( doc.sections, ns)
            state = Default
          case SubsectionHeading:
            oldBlock := doc.popBlock()
            doc.newBlock(BsSubsection)
            doc.newFragment(FsNone, oldBlock.flatten())
            state = Default
          default:
            fmt.Println("Dropping in ",state, lines[i])
        }
      default:
        fmt.Println("Stuck in ",state)
    }
    //fmt.Println(i, classifyLine(lines[i]), lines[i])
  }
  return doc
}

func (doc *Document) popBlock() DocBlock {
  s := len(doc.sections)
  if s<1 || len(doc.sections[s-1].blocks)<1 {
      panic("Bad shit...")
  }
  lastSect := &doc.sections[s-1]
  b := len(lastSect.blocks)
  var bl DocBlock
  lastSect.blocks,bl = lastSect.blocks[:b-1], lastSect.blocks[b-1]
  if len(lastSect.blocks)==0 {
    doc.sections = doc.sections[:s-1]
  }
  return bl
}

func (b *DocBlock) flatten() string {
  var res []byte
  for _,f := range(b.frags) {
    if f.style!=FsNone { panic("Cannot flatten styled text!") }
    res = append(res, []byte(f.cnt)... )
  }
  return string(res)
}

func handler(out http.ResponseWriter, req *http.Request) {
  fmt.Println("Req:", req.URL)
  switch req.URL.Path {
      case "/styles.css":
        cnt,_ := ioutil.ReadFile("data/styles.css")
        out.Write(cnt)
      default:
        filename := "data" + req.URL.Path + ".rst"
        cnt, err := ioutil.ReadFile(filename)
        if err==nil {
          doc := parseRst( string(cnt) )
          doc.renderHtml(out)
        } else {
          fmt.Println(err)
        }
  }
}

func main() {
    http.HandleFunc("/", handler)
    http.ListenAndServe(":8080", nil)
}

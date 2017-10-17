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

func toString(bs BlockStyle) string {
    switch bs {
        case BsNone: return "BsNone"
        case BsBulleted: return "BsBulleted"
        case BsNumbered: return "BsNumbered"
        case BsQuote: return "BsQuote"
        case BsBeginTopic: return "BsBeginTopic"
        case BsEndTopic: return "BsEndTopic"
        case BsBibItem: return "BsBibItem"
        case BsSubsection: return "BsSubsection"
        case BsLiteral: return "BsLiteral"
        default: panic("Unknown BlockStyle")
    }
}

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
    curBlock.frags = append(curBlock.frags, DocFrag{style:style, cnt:content})
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
                    fmt.Println("renderHtml/unknown block",b,curBlock.style,len(curBlock.frags))
                    for f, frag := range(curBlock.frags) {
                        fmt.Println("F:", f, frag.style, frag.cnt)
                    }
            }
        }
    }
    out.Write( []byte("</body></html>") )
}

type LineClassE int
const (
      Blank LineClassE = iota
      SectionHeading
      SubsectionHeading
      Directive
      Bulleted
      Numbered
      Attribute
      Other
)
type LineClass struct {
    indent int
    kind   LineClassE
}

// Simple abstract domain for the types of lines in an .rst
func classifyLine(line string) LineClass {
    if len(line)==0 {
        return LineClass{0,Blank}
    }
    // Tabs are defined in the .rst "spec"
    expanded := strings.Replace(line,"\t","        ",-1)
    if len(strings.Trim(expanded,"="))==0 {
        return LineClass{0,SectionHeading}
    }
    if len(strings.Trim(expanded,"-"))==0 {
        return LineClass{0,SubsectionHeading}
    }

    text := strings.TrimLeft( expanded," " )
    indent := len(expanded) - len(text)

    if text[0:2]==".." {
      return LineClass{indent,Directive}
    }
    if text[0:2]=="* " {
      return LineClass{indent,Bulleted}
    }
    if regexp.MustCompile("[1-9][0-9]*[.] ").MatchString(text) {
      return LineClass{indent,Numbered}
    }

    slices := strings.Split(line,":")
    if len(slices)==3     &&
       len(slices[0])==0  &&
       !strings.Contains(slices[1]," ") {
      return LineClass{indent,Attribute}
    }

    return LineClass{indent,Other}
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
  if len(b.frags)==0 { return }     // Topic Begin/End markers
  if len(b.frags)!=1 { b.dump(); panic("Blocks should star with one fragment") }
  // TODO: Initial block should be single fragment
  //      Find inline markers and split into multi-frags...
  fmt.Println("Initial block size: ",len(b.frags))
  inline := regexp.MustCompile("`[^`]+`_")
  splits := inline.Split(b.frags[0].cnt,-1)
  fmt.Println(splits)
}

func (doc *Document) processInlines() {
    for _,s := range(doc.sections) {
        for _,b := range(s.blocks) {
          b.processInlines()
        }
    }
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
            panic("Unexpected line type in InTopic")
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
          case Bulleted:
            doc.newBlock(BsEndTopic)
            doc.newBlock(BsBulleted)
            doc.newFragment(FsNone, lines[i][2:])
            state = InBullet
          default:
            panic("Unexpected line type in InTopic2")
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
            case Directive:
                re := regexp.MustCompile(".. *([A-Za-z]+):: *(.*)")
                m := re.FindStringSubmatch(lines[i])
                parseDirective(&doc, m[1], m[2])
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

func (block *DocBlock) dump() {
    fmt.Println("Block",len(block.frags),toString(block.style))
    for f,frag := range(block.frags) {
        fmt.Println(" Frag",f,frag.style,frag.cnt)
    }
}

func (doc *Document) dump() {
    for s,section := range(doc.sections) {
        fmt.Println("Section",s,section.title)
        for _,block := range(section.blocks) {
            block.dump()
        }
    }
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
          doc.dump()
          doc.processInlines()
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

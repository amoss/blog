package main

import (
    //"bytes"
    //"regexp"
    //"bufio"
    "os"
    "fmt"
    "runtime"
)
var lineParserDbg = false
var lineParserStDbg = false

func main() {
    for _,arg := range os.Args[2:] {
        switch arg {
            case "--lines":  lineScannerDbg  = true
            case "--parse":  lineParserStDbg = true
            case "--blocks": lineParserDbg   = true
            default: panic("Unrecognised arg "+arg)
        }
    }
    lines := LineScanner(os.Args[1])
    parse(*lines)
}

func processMetadata(key []byte, value []byte) {
    fmt.Println("Dropping:",string(key),"=",string(value))
}

type ParseSt struct {
    input       chan LineClass
    cur         LineClass
    indent      int
    topicIndent int
    pos         int
    body        []byte
    directive   []byte
    //output    chan Block
}
type StateFn func(*ParseSt) StateFn
func (st *ParseSt) String() string {
    return fmt.Sprintf("PSt{Line %d Ind %d}", st.pos, st.indent)
}

func (p *ParseSt) next() {
    p.cur = <-p.input
    p.pos++
}

func dbg(st *ParseSt) {
    pc,_,_,_ := runtime.Caller(1)
    callingName := runtime.FuncForPC(pc).Name()
    if lineParserStDbg {
        fmt.Printf("LineParser: %s %s %s\n", callingName, st, st.cur)
    }
}

func ParseSt_Init(st *ParseSt) StateFn {
    dbg(st)
    if st.topicIndent>=0  &&
       st.cur.kind!=Blank &&
       st.cur.indent<st.topicIndent {
        fmt.Println("Emit: TopicEnd()")
        st.topicIndent=-1
    }
    switch st.cur.kind {
        case SectionHeading:
            st.next()
            return ParseSt_InHeading
        case Other:
            return ParseSt_InPara
        case Blank:
            st.next()
            return ParseSt_Init
        case Directive:
            //mlen   := len(st.cur.marker)
            st.body      = st.cur.body
            st.directive = st.cur.marker
            st.indent    = -1
            st.next()
            return ParseSt_InDirective
        case Numbered:
            st.indent = len(st.cur.marker)
            st.body   = st.cur.body
            st.next()
            return ParseSt_Numbered
        case Bulleted:
            st.indent = len(st.cur.marker)
            st.body   = st.cur.body
            st.next()
            return ParseSt_Bulleted
        case EOF:
            return nil
        default:
            panic("Don't know how to parse "+st.cur.String())
    }
    return nil
}


func ParseSt_Numbered(st *ParseSt) StateFn {
    dbg(st)
    switch st.cur.kind {
        case Numbered:
            fmt.Printf("Emit: Numbered(%s)\n", st.body)
            st.body   = st.cur.body
            st.indent = len(st.cur.marker)
            st.next()
            return ParseSt_Numbered
        case Other:
            st.body = append(st.body, []byte(" ")...)
            st.body = append(st.body, st.cur.body...)
            st.next()
            return ParseSt_Numbered
        case Blank:
            fmt.Printf("Emit: Numbered(%s)\n", st.body)
            st.next()
            st.indent = -1
            return ParseSt_Init
        default: panic("Can't continue numbers with "+string(st.cur.kind))
    }
}


func ParseSt_Bulleted(st *ParseSt) StateFn {
    dbg(st)
    switch st.cur.kind {
        case Bulleted:
            fmt.Printf("Emit: Bullet(%s)\n", st.body)
            st.body   = st.cur.body
            st.indent = len(st.cur.marker)
            st.next()
            return ParseSt_Bulleted
        case Other:
            st.body = append(st.body, []byte(" ")...)
            st.body = append(st.body, st.cur.body...)
            st.next()
            return ParseSt_Bulleted
        case Blank:
            fmt.Printf("Emit: Bullet(%s)\n", st.body)
            st.next()
            st.indent = -1
            return ParseSt_Init
        default: panic("Can't continue bullets with "+string(st.cur.kind))
    }
}


func ParseSt_InDirective(st *ParseSt) StateFn {
    dbg(st)
    dirName := st.directive[3:len(st.directive)-2]
    switch string(dirName) {
        case "image":
            fmt.Println("Emit: image block "+string(st.body))
            return ParseSt_Init
        case "video":
            fmt.Println("Emit: video block "+string(st.body))
            return ParseSt_Init
        case "shell","code":
            if st.cur.kind!=Blank {
                panic("Must put shell lit in separate block")
            }
            st.next()
            st.indent = st.cur.indent
            fmt.Println("Scanning literal with indent of",st.indent)
            st.body   = make( []byte, 0, 1024)
            for st.cur.indent>=st.indent  ||  st.cur.kind==Blank {
                if len(st.body)>0 { st.body = append(st.body, byte('\n')) }
                st.body = append(st.body, st.cur.body...)
                st.next()
            }
            if st.cur.kind==Blank { st.next() }
            // Leave implicit newline from last Blank intact
            // Missing terminating blank will not be detected - is it worth it?
            switch string(dirName) {
                case "shell":
                    fmt.Println("Emit: Shell, Literal block...")
                    fmt.Println(string(st.body))
                case "code":
                    fmt.Println("Emit: Code, Literal block...")
                    fmt.Println(string(st.body))
            }
            st.indent = -1
            return ParseSt_Init
        case "topic":
            fmt.Println("Emit: TopicBegin("+string(st.body)+")")
            if st.cur.kind!=Blank { panic("Topic requires a blank") }
            st.next()
            if st.topicIndent!=-1 { panic("Cannot nest topics") }
            st.topicIndent = st.cur.indent
            return ParseSt_Init
        case "reference":
            return ParseSt_Reference
        case "quote":
            if st.cur.kind!=Blank { panic("Need a blank after quote") }
            st.next()
            body := []byte("")
            for st.cur.kind==Other && st.cur.indent>0 {
                if len(body)>0 { body = append(body, []byte(" ")...) }
                body = append(body, st.cur.body...)
                st.next()
            }
            fmt.Printf("Emit: Quote(attributation=%s body=%s)\n",
                       st.body, body)
            return ParseSt_Init
        default:
            panic("Unrecognised directive "+string(st.directive))
    }
    panic("Missing implementation")
}


func ParseSt_Reference(st *ParseSt) StateFn {
    title  := []byte("")
    author := []byte("")
    url    := []byte("")
    for st.cur.kind==Attribute && st.cur.indent>0 {
        switch string(st.cur.marker) {
            case ":title:":  title = st.cur.body
            case ":author:": author = st.cur.body
            case ":url:":    url = st.cur.body
            default:       panic("Unknown refererence attribute "+string(st.cur.marker))
        }
        st.next()
    }
    fmt.Printf("Emit: Reference(title=%s author=%s url=%s)\n",
                title, author, url)
    return ParseSt_Init
}


func ParseSt_InPara(st *ParseSt) StateFn {
    dbg(st)
    st.indent = st.cur.indent
    body := make( []byte,0,1024 )
    body = append(body, st.cur.body...)
    first := st.pos

    st.next()
    // Not a regular para - make a definition list
    if st.cur.indent > st.indent {
        body   := st.cur.body
        indent := st.cur.indent
        st.next()
        for st.cur.kind==Other && st.cur.indent==indent {
            body = append(body, []byte(" ")...)
            body = append(body, st.cur.body...)
            st.next()
        }
        fmt.Printf("Emit: DefList(def=%s body=%s)\n", st.body, body)
        st.indent = -1
        return ParseSt_Init
    }
    for st.cur.kind==Other && st.cur.indent==st.indent {
        body = append(body, byte(' '))
        body = append(body, st.cur.body...)
        st.next()
    }
    switch st.cur.kind {
        case Blank:
            fmt.Printf("Emit: Paragraph(%s)\n", body)
            st.next()
            st.indent = -1
            return ParseSt_Init
        case SectionHeading:
            if st.pos-first > 1 {
                fmt.Println("Ambiguous para/section heading: use blank")
            } else {
                fmt.Printf("Emit: MediumHeading(%s)\n", body)
            }
            st.next()  // eat it
            st.indent = -1
            return ParseSt_Init
        case SubsectionHeading:
            if st.pos-first > 1 {
                fmt.Println("Ambiguous para/section heading: use blank")
            } else {
                fmt.Printf("Emit: SmallHeading(%s)\n", body)
            }
            st.next()  // eat it
            st.indent = -1
            return ParseSt_Init
        default:
            panic("Can't end a paragraph with "+string(st.cur.kind))
    }
}


func ParseSt_InHeading(st *ParseSt) StateFn {
    dbg(st)
    body := make( []byte,0,1024 )
    body = append(body,st.cur.body...)

    st.next()
    for st.cur.kind==Other && st.cur.indent==0 {
        body = append(body, byte(' '))
        body = append(body, st.cur.body...)
        st.next()
    }

    if st.cur.kind!=SectionHeading {
      panic("Unterminated heading")
    }
    st.next()
    fmt.Printf("Emit: BigSection(%s)\n", body)
    for st.cur.kind==Attribute {
        processMetadata(st.cur.marker[1:len(st.cur.marker)-1],
                        st.cur.body)
        st.next()
    }
    return ParseSt_Init
}

func parse(input chan LineClass) {
  state := &ParseSt{input:input}
  state.indent = -1
  state.topicIndent = -1
  for stateFn := ParseSt_Init; stateFn != nil; {
      stateFn = stateFn(state)
  }

}

package main

import (
    //"bytes"
    //"regexp"
    //"bufio"
    "os"
    "fmt"
)
func main() {
    lines := LineScanner(os.Args[1])
    parse(*lines)
}

type ParseSt struct {
    input     chan LineClass
    cur       LineClass
    indent    int
    pos       int
    body      []byte
    directive []byte
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


func ParseSt_Init(st *ParseSt) StateFn {
    fmt.Println("ParseInit:", st, st.cur)
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
/*
            if st.cur.kind == Other {
                st.indent = st.cur.indent
                st.body   = append(st.body, st.cur.body...)
                return ParseSt_InDirective
            } else if st.cur.kind == Blank {
                st.indent = -1
                st.next()
                return ParseSt_InLiteral    // Not always...
            } else {
                panic("Can't follow directive with "+string(st.cur.kind))
            }
            // Have not handled quoting...
*/
        default:
            panic("Don't know how to parse "+st.cur.String())
    }
    return nil
}


func ParseSt_InDirective(st *ParseSt) StateFn {
    fmt.Println("ParseInDirective:", st, st.cur)
    switch string(st.directive[3:len(st.directive)-2]) {
        case "image":
            fmt.Println("Emit: image block "+string(st.body))
            return ParseSt_Init
        case "shell":
            if st.cur.kind!=Blank { 
                panic("Must put shell lit in separate block") 
            }
            st.next()
            st.indent = st.cur.indent
            st.body   = make( []byte, 0, 1024)
            for st.cur.indent>=st.indent  ||  st.cur.kind==Blank {
                if len(st.body)>0 { st.body = append(st.body, byte('\n')) }
                st.body = append(st.body, st.cur.body...)
                st.next()
            }
            // Leave implicit newline from last Blank intact
            st.next()
            fmt.Println("Emit: Shell, Literal block...")
            fmt.Println(string(st.body))
            return ParseSt_Init
        case "code":
        case "topic":
            fmt.Println("Emit: TopicBegin("+string(st.body)+")")
            st.indent = -1
            return ParseSt_InTopic
        case "reference":
        case "quote":
        case "video":
        default:
            panic("Unrecognised directive "+string(st.directive))
    }
    panic("Missing implementation")
}


func ParseSt_InTopic(st *ParseSt) StateFn {
    if st.cur.kind==Blank  &&  st.indent==-1 {
        st.next()
        st.indent = st.cur.indent
    }
    if st.cur.indent<st.indent {
        // Eat trailing blank from body - it was a separator
        fmt.Println("Emit: TopicEnd()")
        return ParseSt_Init
    }
    return nil
}


func ParseSt_InPara(st *ParseSt) StateFn {
    fmt.Println("ParseInPara:", st, st.cur)
    st.indent = st.cur.indent
    body := make( []byte,0,1024 )
    body = append(body, st.cur.body...)
    first := st.pos

    st.next()
    for st.cur.kind==Other && st.cur.indent==st.indent {
        body = append(body, byte(' '))
        body = append(body, st.cur.body...)
        st.next()
    }
    switch st.cur.kind {
        case Blank:
            fmt.Printf("Emit: Paragraph(%s)\n", body)
            st.next()
            return ParseSt_Init
        case SectionHeading:
            if st.pos-first > 1 {
                fmt.Println("Ambiguous para/section heading: use blank")
            } else {
                fmt.Printf("Emit: MediumHeading(%s)\n", body)
            }
            st.next()  // eat it
            return ParseSt_Init
        case SubsectionHeading:
            if st.pos-first > 1 {
                fmt.Println("Ambiguous para/section heading: use blank")
            } else {
                fmt.Printf("Emit: SmallHeading(%s)\n", body)
            }
            st.next()  // eat it
            return ParseSt_Init
        default:
            panic("Can't end a paragraph with "+string(st.cur.kind))
    }
}


func ParseSt_InHeading(st *ParseSt) StateFn {
    fmt.Println("ParseInHeading:", st, st.cur)
    body := make( []byte,0,1024 )
    body = append(body,st.cur.body...)

    st.next()
    for st.cur.kind==Other && st.cur.indent==st.indent {
        body = append(body, byte(' '))
        body = append(body, st.cur.body...)
        st.next()
    }

    if st.cur.kind!=SectionHeading {
      panic("Unterminated heading")
    }
    st.next()
    fmt.Printf("Emit: BigSection(%s)\n", body)
    return ParseSt_Init
}

func parse(input chan LineClass) {
  state := &ParseSt{input:input}
  for stateFn := ParseSt_Init; stateFn != nil; {
      stateFn = stateFn(state)
  }

}

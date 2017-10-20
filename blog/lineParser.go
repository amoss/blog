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
            mlen := len(st.cur.marker)
            st.next()
            if st.cur.kind == Blank {
                st.next()
                st.indent = st.cur.indent
                return ParseSt_InDirective
            } else {
                st.indent = mlen
                return ParseSt_InDirective
            }
        default:
            panic("Don't know how to parse "+st.cur.String())
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

func ParseSt_InDirective(st *ParseSt) StateFn {
    return nil
}

func parse(input chan LineClass) {
  state := &ParseSt{input:input}
  for stateFn := ParseSt_Init; stateFn != nil; {
      stateFn = stateFn(state)
  }

}

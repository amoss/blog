package rst

import (
    "runtime/debug"
    "bytes"
    "strings"
    "fmt"
    "runtime"
)
var LineParserDbg = false
var LineParserStDbg = false

func processMetadata(key []byte, value []byte) {
    fmt.Println("Dropping:",string(key),"=",string(value))
}

type BlockE int
const (
    BlkParagraph BlockE = iota
    BlkBulleted
    BlkNumbered
    BlkBigHeading
    BlkMediumHeading
    BlkSmallHeading
    BlkDefList
    BlkTopicBegin
    BlkTopicEnd
    BlkBeginLongform
    BlkEndLongform
    BlkImage
    BlkVideo
    BlkShell
    BlkCode
    BlkQuote
    BlkReference
    BlkTableRow
    BlkTableCell
    BlkError
)
type Block struct {
    Kind     BlockE
    Body     []byte
    Author   []byte
    Title    []byte
    Url      []byte
    Heading  []byte
    Date     []byte
    Style    []byte
    Detail   []byte
    Position []byte
    CourseCode   []byte
    CourseName   []byte
    Location     []byte
    Subtitle     []byte
    Tags         []byte
}

type ParseSt struct {
    input       chan LineClass
    cur         LineClass
    indent      int
    topicIndent int
    pos         int
    tabColumns  []int
    body        []byte
    directive   []byte
    output      chan Block
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
    if LineParserStDbg {
        fmt.Printf("%29s: %s %s %s\n", "parser", callingName, st, st.cur)
    }
}

func dbgForce(st *ParseSt) {
    pc,_,_,_ := runtime.Caller(1)
    callingName := runtime.FuncForPC(pc).Name()
    fmt.Printf("%29s: %s %s %s\n", "parser", callingName, st, st.cur)
}

func ParseSt_Init(st *ParseSt) StateFn {
    dbg(st)
    if st.topicIndent>=0  &&
       st.cur.kind!=Blank &&
       st.cur.indent<st.topicIndent {
        st.output <- Block{Kind:BlkTopicEnd}
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
        case TableSeparator:
            colWidths := bytes.Split(st.cur.body,[]byte("+"))
            // Split contains 0-len at beginning at end that we trim
            st.tabColumns = make([]int,len(colWidths)-2)
            for i:=0; i<len(colWidths)-2; i++ {
                st.tabColumns[i] = len(colWidths[i+1])
            }
            st.next()
            return ParseSt_Table
        case Comment:
            st.next()
            return ParseSt_Init
        case BeginLongform:
            st.next()
            st.output <- Block{Kind:BlkBeginLongform}
            return ParseSt_Init
        case EndLongform:
            st.next()
            st.output <- Block{Kind:BlkEndLongform}
            return ParseSt_Init
        default:
            panic("Don't know how to parse "+st.cur.String())
    }
    return nil
}


func ParseSt_Table(st *ParseSt) StateFn {
    dbg(st)
    fmt.Println("Table state", st.tabColumns)
    rowCells := make( [][]byte, len(st.tabColumns))
    for i:=0;i<len(rowCells);i++ {
        rowCells[i] = make( []byte,0,1024 )
    }
    for {
        switch st.cur.kind {
            case Blank:
                st.next()
                return ParseSt_Init
            case TableSeparator:
                st.output <- Block{Kind:BlkTableRow}
                offset := 1
                for i:=0; i<len(st.tabColumns); i++ {
                    if st.cur.body[offset+st.tabColumns[i]]!='+' {
                        msg := fmt.Sprintf("Table cells not aligned properly %d / %s \"%s\"!",
                                           i,st.tabColumns, st.cur.body[:offset+st.tabColumns[i]+1])
                        panic(msg)
                    }
                    st.output <- Block{Kind:BlkTableCell,Body:rowCells[i]}
                    offset += st.tabColumns[i]+1
                }
                for i:=0;i<len(rowCells);i++ {
                    rowCells[i] = make( []byte,0,1024 )
                }
                st.next()
            case TableRow:
                offset := 1
                for i:=0; i<len(st.tabColumns); i++ {
                    cell := st.cur.body[offset:offset+st.tabColumns[i]]
                    rowCells[i] = append( rowCells[i], cell...)
                    if st.cur.body[offset+st.tabColumns[i]]!='|' {
                        panic("Table cells not aligned properly!")
                    }
                    offset += st.tabColumns[i]+1
                }
                st.next()
            default:
                fmt.Printf("State %d\n",st.cur.kind);
                panic("Unexpected line in table")
        }
    }
}


func ParseSt_Numbered(st *ParseSt) StateFn {
    dbg(st)
    switch st.cur.kind {
        case Numbered:
            st.output <- Block{Kind:BlkNumbered,Body:st.body}
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
            st.output <- Block{Kind:BlkNumbered,Body:st.body}
            st.next()
            st.indent = -1
            return ParseSt_Init
        case EndLongform:
            st.next()
            st.output <- Block{Kind:BlkEndLongform}
            return ParseSt_Init
        case EOF:
            st.output <- Block{Kind:BlkNumbered,Body:st.body}
            return nil
        default: panic("Can't continue numbers with "+string(st.cur.kind))
    }
}


func ParseSt_Bulleted(st *ParseSt) StateFn {
    dbg(st)
    switch st.cur.kind {
        case Bulleted:
            st.output <- Block{Kind:BlkBulleted,Body:st.body}
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
            st.output <- Block{Kind:BlkBulleted,Body:st.body}
            st.next()
            st.indent = -1
            return ParseSt_Init
        case EndLongform:
            st.next()
            st.output <- Block{Kind:BlkEndLongform}
            return ParseSt_Init
        case EOF:
            st.output <- Block{Kind:BlkBulleted,Body:st.body}
            return nil
        default: panic("Can't continue bullets with "+string(st.cur.kind))
    }
}


func ParseSt_InDirective(st *ParseSt) StateFn {
    dbg(st)
    dirName := st.directive[3:len(st.directive)-2]
    switch string(dirName) {
        case "image":
            st.output <- Block{Kind:BlkImage,Body:st.body}
            return ParseSt_Init
        case "video":
            st.output <- Block{Kind:BlkVideo,Body:st.body}
            return ParseSt_Init
        case "shell","code":
            position := []byte("default")
            title := st.body
            style := []byte("")
            if st.cur.kind==Attribute {
                switch string( bytes.ToLower(st.cur.marker) ) {
                    case "position":
                        position = bytes.ToLower(st.cur.body)
                        st.next()
                    case "style":
                        style = st.cur.body
                        st.next()
                    default:
                        panic("Unknown position attribute "+string(st.cur.marker))
                }
            }
            if st.cur.kind!=Blank {
                panic("Must put shell lit in separate block")
            }
            st.next()
            st.indent = st.cur.indent
            if st.indent==0 {
                panic("Only one blank line after directive!")
            }
            st.body   = make( []byte, 0, 1024)
            for st.cur.indent>=st.indent  ||  st.cur.kind==Blank {
                if len(st.body)>0 { st.body = append(st.body, byte('\n')) }
                if st.cur.indent-st.indent>0 {
                    reIndent := bytes.Repeat([]byte(" "),st.cur.indent-st.indent)
                    st.body = append(st.body, reIndent...)
                }
                st.body = append(st.body, st.cur.body...)
                st.next()
            }
            if st.cur.kind==Blank { st.next() }
            // Kill the newline from the final Blank.
            if st.body[ len(st.body)-1 ] == '\n' {
              st.body = st.body[:len(st.body)-1]
            }

            // Missing terminating blank will not be detected - is it worth it?
            switch string(dirName) {
                case "shell":
                    st.output <- Block{Kind:BlkShell,Body:st.body,Position:position,Style:style}
                case "code":
                    st.output <- Block{Kind:BlkCode,Body:st.body,Position:position,Style:style,Title:title}
            }
            st.indent = -1
            return ParseSt_Init
        case "topic":
            st.output <- Block{Kind:BlkTopicBegin,Body:st.body}
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
            st.output <- Block{Kind:BlkQuote,Body:body,Author:st.body}
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
    detail    := []byte("")
    for st.cur.kind==Attribute && st.cur.indent>0 {
        switch string(st.cur.marker) {
            case "title":  title = st.cur.body
            case "author": author = st.cur.body
            case "url":    url = st.cur.body
            case "detail": detail = st.cur.body
            default:       panic("Unknown refererence attribute "+string(st.cur.marker))
        }
        st.next()
    }
    st.output <- Block{Kind:BlkReference,Title:title,Author:author,Url:url,Detail:detail}
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
        heading := body
        body := make( []byte,0,1024 )
        body  = append(body, st.cur.body...)
        indent  := st.cur.indent
        st.next()
        for st.cur.kind==Other && st.cur.indent==indent {
            body = append(body, []byte(" ")...)
            body = append(body, st.cur.body...)
            st.next()
        }
        st.output <- Block{Kind:BlkDefList,Body:body,Heading:heading}
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
            st.output <- Block{Kind:BlkParagraph,Body:body}
            st.next()
            st.indent = -1
            return ParseSt_Init
        case SectionHeading:
            if st.pos-first > 1 {
                fmt.Println("Ambiguous para/section heading: use blank")
                st.next()  // eat it
            } else {
                st.next()  // eat it
                layout := []byte("single")
                for st.cur.kind==Attribute {
                    if strings.ToLower(string(st.cur.marker))=="layout" {
                        layout = bytes.ToLower(st.cur.body)
                    }
                    st.next()
                }
                st.output <- Block{Kind:BlkMediumHeading,Body:body,Style:layout}
            }
            st.indent = -1
            return ParseSt_Init
        case SubsectionHeading:
            if st.pos-first > 1 {
                fmt.Println("Ambiguous para/section heading: use blank")
                st.next()  // eat it
            } else {
                st.next()  // eat it
                layout := []byte("single")
                for st.cur.kind==Attribute {
                    if strings.ToLower(string(st.cur.marker))=="layout" {
                        layout = bytes.ToLower(st.cur.body)
                    }
                    st.next()
                }
                st.output <- Block{Kind:BlkMediumHeading,Body:body,Style:layout}
            }
            st.indent = -1
            return ParseSt_Init
        case EndLongform:
            st.next()
            st.output <- Block{Kind:BlkEndLongform}
            return ParseSt_Init
        case EOF:
            st.output <- Block{Kind:BlkParagraph,Body:body}
            return nil
        default:
            dbgForce(st)
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
    metadata := map[string] []byte {
        "author" : []byte("No author specified"),
        "date"   : []byte("No date specified"),
    }
    for st.cur.kind==Attribute {
        atName := string(st.cur.marker)
        metadata[ strings.ToLower(atName) ] = bytes.TrimLeft(st.cur.body," ")
        st.next()
    }
    st.output <- Block{Kind:BlkBigHeading, Title:body,
                       Author:metadata["author"], CourseCode:metadata["coursecode"],
                       CourseName:metadata["coursename"], Location:metadata["location"],
                       Date:metadata["date"], Style:bytes.ToLower(metadata["style"]),
                       Tags:metadata["tags"], Subtitle:metadata["subtitle"]}
    return ParseSt_Init
}

func Parse(input chan LineClass) chan Block {
    state := &ParseSt{input:input}
    state.indent = -1
    state.topicIndent = -1
    state.output = make(chan Block)
    go func() {
        defer close(state.output)
        defer func(){
            if r:= recover(); r!=nil {
                errMsg := []byte( fmt.Sprintf("Panic during parse! %s %s\n", r, debug.Stack() ) )
                state.output <- Block{Kind:BlkError, Body:errMsg}
            }
        }()
        for stateFn := ParseSt_Init; stateFn != nil; {
            stateFn = stateFn(state)
        }
  }()
  return state.output

}

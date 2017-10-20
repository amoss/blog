package main
import (
    "bytes"
    "regexp"
    "bufio"
    "os"
    "fmt"
)
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
    marker []byte
    body   []byte
    kind   LineClassE
}

func (lc LineClass) String() string {
  return fmt.Sprintf("{%v %s \"%s\" \"%s\"}",lc.indent,lc.kind,lc.marker,lc.body)
}

func classify(line []byte) LineClass {
    if len(line)==0 { return LineClass{} }      // Blank
    expanded := bytes.Replace(line, []byte("\t"), []byte("        "),-1)

    if len(bytes.Trim(expanded," "))==0 {
        return LineClass{}                      // Blank
    }
    if len(bytes.Trim(expanded,"="))==0 {
        return LineClass{0,nil,nil,SectionHeading}
    }
    if len(bytes.Trim(expanded,"-"))==0 {
        return LineClass{0,nil,nil,SubsectionHeading}
    }

    text := bytes.TrimLeft( expanded," " )
    indent := len(expanded) - len(text)

    dir := regexp.MustCompile("^[.][.] [A-Za-z]+::")
    d := dir.FindSubmatch(text)
    if d!=nil {
        return LineClass{indent,d[0],text[indent+len(d[0]):],Directive}
    }
    if bytes.Compare(text[0:2],[]byte("* "))==0 {
      return LineClass{indent, []byte("* "), text[2:], Bulleted}
    }
    nums := regexp.MustCompile("^[1-9][0-9]*[.] ")
    m := nums.FindSubmatch(text)
    if m!=nil {
      return LineClass{indent,m[0],text[indent+len(m[0]):],Numbered}
    }

    slices := bytes.Split(text,[]byte(":"))
    if len(slices)==3     &&
       len(slices[0])==0  &&
       !bytes.Contains(slices[1],[]byte(" ")) {
      return LineClass{indent, text[0:2+len(slices[1])],
                       text[2+len(slices[1]):],Attribute}
    }

    return LineClass{indent,nil,text,Other}

}

func main() {
  fd,_    := os.Open(os.Args[1])
  scanner := bufio.NewScanner(fd)
  lines   := make( []LineClass, 0, 1024 )
  for scanner.Scan() {
    lines = append( lines, classify( scanner.Bytes() ) )
  }
  fmt.Println(lines[0])
  parse(lines)
}

type ParseSt struct {
    input     []LineClass
    indent    int
    pos       int
    //output    chan Block
}
type StateFn func(*ParseSt) StateFn
func (st *ParseSt) String() string {
    return fmt.Sprintf("PSt{Line %d/%d Ind %d}", st.pos, len(st.input), st.indent)
}


func (st *ParseSt) peek() *LineClass{
    if st.pos == len(st.input) { return nil }
    res := &st.input[ st.pos ]
    return res
}


func ParseSt_Init(st *ParseSt) StateFn {
    fmt.Println("ParseInit:", st, st.peek())
    cur := st.peek()
    switch cur.kind {
        case SectionHeading:
            st.pos++
            return ParseSt_InHeading
        case Other:
            st.indent = cur.indent
            return ParseSt_InPara
        case Blank:
            st.pos++
            return ParseSt_Init
        case Directive:
            st.pos++;
            if st.input[st.pos].kind == Blank {
                st.pos++;
                st.indent = st.input[st.pos].indent
                return ParseSt_InDirective
            } else {
                st.indent = len(st.input[st.pos-1].marker)
                return ParseSt_InDirective
            }
        default:
            panic("Don't know how to parse "+cur.String())
    }
    return nil
}


func ParseSt_InPara(st *ParseSt) StateFn {
    fmt.Println("ParseInPara:", st, st.peek())
    end  := st.pos
    size := 0
    indent := st.indent
    for st.input[end].kind==Other && end<len(st.input) &&
        st.input[end].indent==indent {
        size += len(st.input[end].body)+1
        end++
    }
    body := make( []byte,0,size )
    for i := st.pos; i<end; i++ {
        if i!=st.pos { body = append(body, byte(' ')) }
        body = append(body, st.input[i].body...)
    }
    fmt.Println("ParseInPara:", end, st.input[end])
    switch st.input[end].kind {
        case Blank:
            fmt.Printf("Emit: Paragraph(%s)\n", body)
            st.pos = end
            st.pos++;
            return ParseSt_Init
        case SectionHeading:
            if end-st.pos > 1 {
                fmt.Println("Ambiguous para/section heading: use blank")
                st.pos = end
                st.pos++
                return ParseSt_Init
            }
            fmt.Printf("Emit: MediumHeading(%s)\n", body)
            st.pos = end
            st.pos++;
            return ParseSt_Init
        case SubsectionHeading:
            if end-st.pos > 1 {
                fmt.Println("Ambiguous para/section subheading: use blank")
                st.pos = end
                st.pos++
                return ParseSt_Init
            }
            fmt.Printf("Emit: SmallHeading(%s)\n", body)
            st.pos = end
            st.pos++;
            return ParseSt_Init
        default:
            panic("Can't end a paragraph with "+string(st.input[st.pos].kind))
    }
}


func ParseSt_InHeading(st *ParseSt) StateFn {
    fmt.Println("ParseInHeading:", st, st.peek())
    end  := st.pos
    size := 0
    for st.input[end].kind==Other && end<len(st.input) {
        size += len(st.input[end].body)+1
        end++
    }
    title := make( []byte,0,size )
    for i := st.pos; i<end; i++ {
        if i!=st.pos { title = append(title, byte(' ')) }
        title = append(title, st.input[i].body...)
    }
    st.pos = end
    if end==len(st.input) || st.input[end].kind!=SectionHeading {
      panic("Unterminated heading")
    }
    st.pos++
    fmt.Printf("Emit: BigSection(%s)\n", title)
    return ParseSt_Init
}

func parse(input []LineClass) {
  state := &ParseSt{input:input}
  for stateFn := ParseSt_Init; stateFn != nil; {
      stateFn = stateFn(state)
  }

}

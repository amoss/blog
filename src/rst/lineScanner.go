package rst
import (
    "bufio"
    "bytes"
    "io"
    "fmt"
    "os"
    "regexp"
    "runtime/debug"
)
var LineScannerDbg bool = false
type LineClassE int
const (
      Blank LineClassE = iota
      SectionHeading
      SubsectionHeading
      Directive
      Bulleted
      Numbered
      Attribute
      TableSeparator
      TableRow
      Other
      Comment
      BeginLongform
      EndLongform
      EOF
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
    if string(expanded[:7])=="/*{{{*/" {
        return LineClass{0,nil,nil,BeginLongform}
    }
    if string(expanded[:7])=="/*}}}*/" {
        return LineClass{0,nil,nil,EndLongform}
    }

    text := bytes.TrimLeft( expanded," " )
    indent := len(expanded) - len(text)

    if len(bytes.Trim(text,"-+"))==0 {
        return LineClass{indent,nil,text,TableSeparator}
    }
    if text[0]=='|' && text[ len(text)-1 ]=='|' {
        return LineClass{indent,nil,text,TableRow}
    }
    dir := regexp.MustCompile("^[.][.] [A-Za-z]+::")
    d := dir.FindSubmatch(text)
    if d!=nil {
        return LineClass{indent,d[0],text[len(d[0]):],Directive}
    }
    // Overlaps with directives, check if did not match
    com := regexp.MustCompile("^[.][.] ")
    c := com.FindSubmatch(text)
    if c!=nil {
        return LineClass{indent,nil,text,Comment}
    }
    if len(text)==1 {
      return LineClass{indent,nil,text,Other}
    }
    if bytes.Compare(text[0:2],[]byte("* "))==0 {
      return LineClass{indent, []byte("* "), text[2:], Bulleted}
    }
    nums := regexp.MustCompile("^[1-9][0-9]*[.] ")
    m := nums.FindSubmatch(text)
    if m!=nil {
      return LineClass{indent,m[0],text[len(m[0]):],Numbered}
    }

    attribute := regexp.MustCompile("^:([A-Za-z]+): ")
    a := attribute.FindSubmatch(text)
    if a!=nil {
        return LineClass{indent,a[1],bytes.Trim(text[len(a[0]):]," "),Attribute}
    }

    return LineClass{indent,nil,text,Other}

}


func LineScannerBytes(lines []byte) *chan LineClass {
    return LineScanner( bytes.NewReader(lines) )
}


func LineScannerPath(path string) *chan LineClass {
    fd,err := os.Open(path)
    if err!=nil { return nil }
    return LineScanner(fd)
}


func LineScanner(reader io.Reader) *chan LineClass {

    output := make(chan LineClass)

    scanner := bufio.NewScanner(reader)
    go func() {
        defer func() {
            output <- LineClass{kind:EOF}
            close(output)
            r := recover()
            if r!=nil {
                fmt.Printf("Scanner panic! %s\n %s\n",r,debug.Stack() )
            }
        }()

        for scanner.Scan() {
            line := classify( scanner.Bytes() )
            if LineScannerDbg { fmt.Println("Lex:",line) }
            output <- line
        }
        output <- LineClass{kind:EOF}
    }()
    return &output
}

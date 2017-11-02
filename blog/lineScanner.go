package main
import (
    "os"
    "bufio"
    "regexp"
    "bytes"
    "fmt"
)
var lineScannerDbg bool = false
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
        return LineClass{indent,a[1],bytes.TrimLeft(text[len(a[0]):]," "),Attribute}
    }

    return LineClass{indent,nil,text,Other}

}

func LineScanner(path string) *chan LineClass {
    output := make(chan LineClass)

    fd,err := os.Open(path)
    if err!=nil { return nil }
    scanner := bufio.NewScanner(fd)
    go func() {
        defer func() {
            close(output)
            r := recover()
            if r!=nil {
                fmt.Println("Scanner panic!",r)
            }
        }()

        for scanner.Scan() {
            line := classify( scanner.Bytes() )
            if lineScannerDbg { fmt.Println("Lex:",line) }
            output <- line
        }
        output <- LineClass{kind:EOF}
    }()
    return &output
}

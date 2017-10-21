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

func LineScanner(path string) *chan LineClass {
    output := make(chan LineClass)

    fd,err := os.Open(path)
    if err!=nil { return nil }
    scanner := bufio.NewScanner(fd)
    go func() {
        for scanner.Scan() {
            line := classify( scanner.Bytes() )
            if lineScannerDbg { fmt.Println("Lex:",line) }
            output <- line
        }
    }()
    return &output
}

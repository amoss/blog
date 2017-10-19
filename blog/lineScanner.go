package main
import (
    "bytes"
    "regexp"
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

func classify(line []byte) LineClass {
    if len(line)==0 { return LineClass{} }
    expanded := bytes.Replace(line, []byte("\t"), []byte("        "),-1)

    if len(bytes.Trim(expanded,"="))==0 {
        return LineClass{0,nil,nil,SectionHeading}
    }
    if len(bytes.Trim(expanded,"-"))==0 {
        return LineClass{0,nil,nil,SubsectionHeading}
    }

    text := bytes.TrimLeft( expanded," " )
    indent := len(expanded) - len(text)

    dir := regexp.MustCompile("^[.][.][A-Za-z]+::")
    d := dir.FindSubmatch(text)
    if d!=nil {
        return LineClass{indent,d[0],text[indent+len(d):],Directive}
    }
    if bytes.Compare(text[0:2],[]byte("* "))==0 {
      return LineClass{indent, []byte("* "), text[2:], Bulleted}
    }
    nums := regexp.MustCompile("^[1-9][0-9]*[.] ")
    m := nums.FindSubmatch(text)
    if m!=nil {
      return LineClass{indent,m[0],text[indent+len(m):],Numbered}
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

https://stackoverflow.com/questions/46143974/golang-convert-slice-of-string-input-from-console-to-slice-of-numbers

func main() {
  cnt,_ := ioutil.ReadFile("data/part1.rst")


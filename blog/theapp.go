package main

import (
      "net/http"
      "fmt"
      "io/ioutil"
      "strings"
)

type LineClass int
const (
      Blank LineClass = iota
      Indented
      SectionHeading
      SubsectionHeading
      Directive
      Other
)

// Simple abstract domain for the types of lines in an .rst
func classifyLine(line string) LineClass {
  if len(line)==0 {
    return Blank
  }
  // If you are not me then don't use tabs, haha only serious
  expanded := strings.Replace(line,"\t","  ",-1)
  text := strings.TrimLeft( expanded," " )
  indent := len(expanded) - len(text)
  if indent>0 {
    return Indented
  }

  if len(strings.Trim(expanded,"="))==0 {
    return SectionHeading
  }
  if len(strings.Trim(expanded,"-"))==0 {
    return SubsectionHeading
  }
  if line[0:2]==".." {
    return Directive
  }
  return Other
}

type ParseState int
const (
        Default ParseState = iota
        TitleBlock
        InPara
        InDirective
)

func parseRst(src string) {
  lines := strings.Split(src,"\n")
  state := Default
  cur := ""
  for i:= 0; i<len(lines); i++ {
    switch state {
      case Default:
        switch classifyLine(lines[i]) {
          case SectionHeading:
            state = TitleBlock
          case Blank:
            // Ignore
          case Other:
            state = InPara
            cur   = lines[i]
          case Directive:
            state = InDirective
            cur = lines[i][3:]
        }
      case InDirective:
        switch classifyLine(lines[i]) {
          case Indented:
            cur += "\n" + lines[i]    // Do a tab-expansion and TrimLeft
          case Blank:
            fmt.Println("Found a directive", cur)
            state = Default
            cur   = ""
        }

      case TitleBlock:
        switch classifyLine(lines[i]) {
          case Other:
            cur += "\n" + lines[i]
          case SectionHeading:
            fmt.Println("Found a title block", cur)
            state = Default
        }
      case InPara:
        switch classifyLine(lines[i]) {
          case Other:
            cur += "\n" + lines[i]
          case Blank:
            fmt.Println("Found a para", cur)
            state = Default
            cur   = ""
          case SectionHeading:
            fmt.Println("Found a section heading", cur)
            cur = ""
            state = Default
          case SubsectionHeading:
            fmt.Println("Found a subsection heading", cur)
            cur = ""
            state = Default
        }
    }
    fmt.Println(i, classifyLine(lines[i]), lines[i])
  }
}

func handler(out http.ResponseWriter, req *http.Request) {
  fmt.Println("Req:", req.URL)
  filename := "data" + req.URL.Path + ".rst"
  cnt, err := ioutil.ReadFile(filename)
  if err==nil {
    parseRst( string(cnt) )
  } else {
    fmt.Println(err)
  }
}

func main() {
    http.HandleFunc("/", handler)
    http.ListenAndServe(":8080", nil)
}

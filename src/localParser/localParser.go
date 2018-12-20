package main

import (
  "os"
  "rst"
  "fmt"
)

func main() {
    output := false
    for _,arg := range os.Args[2:] {
        switch arg {
            case "--lines":  rst.LineScannerDbg  = true
            case "--parse":  rst.LineParserStDbg = true
            case "--blocks": rst.LineParserDbg   = true
            case "--html":   output = true
            default: panic("Unrecognised arg "+arg)
        }
    }
    lines  := rst.LineScanner(os.Args[1])
    if lines!=nil {
        blocks := rst.Parse(*lines)
        res := rst.RenderHtml(blocks,false)
        if output { fmt.Println(string(res)) }
    }
}

package rst

import (
    "fmt"
    "regexp"
    "bytes"
    "html"
)

func makePageHeader(extra string) []byte {
    result := make([]byte,0,1024)
    result = append(result,[]byte(`
<html>
<head>
<script src="http://cdn.mathjax.org/mathjax/latest/MathJax.js?config=TeX-AMS-MML_HTMLorMML" type="text/javascript"></script>
<link href="/styles.css" type="text/css" rel="stylesheet"></link>
`)...)
    if extra != "" {
        result = append(result, []byte("<link href=\"")...)
        result = append(result, []byte("/"+extra)...)
        result = append(result, []byte(".css\" type=\"text/css\" rel=\"stylesheet\"></link>\n")...)
    }
    if extra=="slides" {
        result = append(result, []byte("<script src=\"/slides.js\" type=\"text/javascript\"></script>")...)
    }
    result = append(result, []byte(`</head>
<body>
`)...)
    return result
}

var pageFooter = []byte(`
</body>
</html>
`)

func inlineStyles(input []byte) []byte {
  links  := regexp.MustCompile("`([^`]+) <([^>]+)>`_")
  input   = links.ReplaceAll(input, []byte("<a href=\"$2\">$1</a>"))
  strong := regexp.MustCompile("\\*\\*([^*]+)\\*\\*")
  input   = strong.ReplaceAll(input, []byte("<b>$1</b>"))
  emp    := regexp.MustCompile(" \\*([^*]+)\\*")
  input   = emp.ReplaceAll(input, []byte(" <i>$1</i>"))
  shell  := regexp.MustCompile(":shell:`([^`]+)`")
  input   = shell.ReplaceAll(input, []byte("<span class=\"shell\">$1</span>"))
  code   := regexp.MustCompile(":code:`([^`]+)`")
  input   = code.ReplaceAll(input, []byte("<span class=\"code\">$1</span>"))
  math   := regexp.MustCompile(":math:`([^`]+)`")
  input   = math.ReplaceAll(input, []byte(`\($1\)`))
  return input
}

var tagNames = map[BlockE]string {
    BlkBulleted: "ul",
    BlkNumbered: "ol",
    BlkDefList:  "dl",
    BlkTableRow: "table",
    BlkTableCell: "table" }



func renderHtmlPage(headBlock Block, input chan Block) []byte {
    result := make([]byte, 0, 16384)
    result = append(result, makePageHeader(string(headBlock.style))...)
    result = append(result, []byte("<div style=\"width:100%; background-color:#dddddd; padding:1rem\">")... )
    result = append(result, []byte("<h1>")... )
    result = append(result, inlineStyles(headBlock.title)... )
    result = append(result, []byte("</h1>")... )
    result = append(result, []byte("<i>")... )
    result = append(result, headBlock.author... )
    result = append(result, []byte("</i>")... )
    result = append(result, []byte("<p>")... )
    result = append(result, headBlock.date... )
    result = append(result, []byte("</p>")... )
    result = append(result, []byte("</div>")... )
    lastKind := BlkParagraph
    for blk := range input {
        if tagNames[lastKind]!="" && blk.kind!=lastKind {
            result = append(result, []byte("</")... )
            result = append(result, []byte(tagNames[lastKind])... )
            result = append(result, []byte(">")... )
        }
        if tagNames[blk.kind]!="" && blk.kind!=lastKind {
            result = append(result, []byte("<")... )
            result = append(result, []byte(tagNames[blk.kind])... )
            result = append(result, []byte(">")... )
        }
        switch blk.kind {
            case BlkParagraph:
                result = append(result, []byte("<p>")... )
                result = append(result, inlineStyles(blk.body)... )
                result = append(result, []byte("</p>")... )
            case BlkNumbered, BlkBulleted:
                result = append(result, []byte("<li>")... )
                result = append(result, inlineStyles(blk.body)... )
                result = append(result, []byte("</li>")... )
            case BlkMediumHeading:
                result = append(result, []byte("<h1>")... )
                result = append(result, inlineStyles(blk.body)... )
                result = append(result, []byte("</h1>")... )
            case BlkSmallHeading:
                result = append(result, []byte("<h2>")... )
                result = append(result, inlineStyles(blk.body)... )
                result = append(result, []byte("</h2>")... )
            case BlkShell:
                result = append(result, []byte("<div class=\"shell\">")... )
                result = append(result, []byte(html.EscapeString(string(blk.body)))... )
                result = append(result, []byte("</div>")... )
            case BlkCode:
                escaped := html.EscapeString(string(blk.body))
                fmt.Println(escaped)
                result = append(result, []byte("<div class=\"code\">")... )
                result = append(result, []byte(escaped)... )
                result = append(result, []byte("</div>")... )
            case BlkTopicBegin:
                result = append(result, []byte("<div class=\"Scallo\"><div class=\"ScalloHd\">")... )
                result = append(result, inlineStyles(blk.body)... )
                result = append(result, []byte("</div>")...)
            case BlkTopicEnd:
                result = append(result, []byte("</div>")... )
            case BlkQuote:
                result = append(result, []byte("<div class=\"quoteinside\"><div class=\"quotebegin\">&#8220;</div>")... )
                result = append(result, inlineStyles(blk.body)... )
                if len(blk.author)>0 {
                    result = append(result, []byte("<br/>--- ")... )
                    result = append(result, blk.author... )
                }
                result = append(result, []byte("<div class=\"quoteend\">&#8221;</div></div>\n")... )
            case BlkImage:
                result = append(result, []byte("<img src=\"")...)
                result = append(result, blk.body... )
                result = append(result, []byte("\" style=\"width:100%; max-height:100%; object-fit:contain\"/>")...)
            case BlkVideo:
                result = append(result, []byte("<video width=\"100%%\" style=\"max-width:100%% max-height:95%%\" controls>\n")... )
                result = append(result, []byte("<source src=\"")... )
                result = append(result, blk.body... )
                result = append(result, []byte(".webm\" type=\"video/webm;\">")...)
                result = append(result, []byte("<source src=\"")... )
                result = append(result, blk.body... )
                result = append(result, []byte(".mov\" type=\"video/quicktime;\">")...)
                result = append(result, []byte("</video>")...)
            case BlkReference:
                result = append(result, []byte("<div class=bibitem><table style=\"width=100%%\">\n<tr><td rowspan=\"3\"><img style=\"width:2rem;height:2rem\" src=\"/book-icon.png\"/></td>\n<td><a href=\"")...)
                result = append(result, blk.url... )
                result = append(result, []byte("\">")...)
                result = append(result, blk.title... )
                result = append(result, []byte("</a></td></tr><tr><td><i>")...)
                result = append(result, blk.author... )
                result = append(result, []byte("</i></td></tr>")...)
                if blk.detail!=nil {
                    result = append(result, []byte("<tr><td>")...)
                    result = append(result, blk.detail... )
                    result = append(result, []byte("</td></tr>")...)
                }
                result = append(result, []byte("</table></div>")...)
            case BlkDefList:
                result = append(result, []byte("<dt>")...)
                result = append(result, blk.heading... )
                result = append(result, []byte("</dt><dd>")...)
                result = append(result, inlineStyles(blk.body)... )
                result = append(result, []byte("</dd>\n")...)
            default:
                fmt.Println("Block:", blk)
        }
        lastKind = blk.kind
    }
    result = append(result, pageFooter...)
    return result
}

type MultiChanSlide struct {
    primary, secondary, longform []byte
    title  []byte
    active string       // primary, secondary, longform
    layout string       // single, rows or columns
}

func (self *MultiChanSlide) extendB( data []byte  ) {
    switch self.active {
        case "primary":
            self.primary = append(self.primary, data...)
        case "secondary":
            self.secondary = append(self.secondary, data...)
        case "longform":
            self.longform = append(self.longform, data...)
    }
}

func (self *MultiChanSlide) extendS( data string ) {
    switch self.active {
        case "primary":
            self.primary = append(self.primary, []byte(data)...)
        case "secondary":
            self.secondary = append(self.secondary, []byte(data)...)
        case "longform":
            self.longform = append(self.longform, []byte(data)...)
    }
}

func (self *MultiChanSlide) Printf( format string, args ...interface{} ) {
  formatted := fmt.Sprintf(format, args...)
    switch self.active {
        case "primary":
            self.primary = append(self.primary, []byte(formatted)...)
        case "secondary":
            self.secondary = append(self.secondary, []byte(formatted)...)
        case "longform":
            self.longform = append(self.longform, []byte(formatted)...)
    }
}

func (self *MultiChanSlide) PrintfHL( highlight bool, format string, args ...interface{} ) {
    formatted := fmt.Sprintf(format, args...)
    if !highlight {
        self.primary = append(self.primary, []byte(formatted)...)
    } else {
        switch self.layout {
            case "rows", "columns":
                self.secondary = append(self.secondary, []byte(formatted)...)
            case "single":
                self.primary = append(self.primary, []byte(formatted)...)
        }
        //self.longform = append(self.longform, []byte(formatted)...)
    }
}

func (self *MultiChanSlide) PrintfLong( format string, args ...interface{} ) {
    formatted := fmt.Sprintf(format, args...)
    self.longform = append(self.longform, []byte(formatted)...)
}

func makeMultiChanSlide(layout string, title []byte) MultiChanSlide{
  result := MultiChanSlide{layout:layout, title:title}
  result.primary   = make([]byte,0,16384)
  result.secondary = make([]byte,0,16384)
  result.longform  = make([]byte,0,16384)
  result.active    = "primary"
  return result
}

func (self *MultiChanSlide) finalise(buffer []byte, counter int) []byte {
    buffer = append(buffer, []byte("\n<div class=\"S169\" id=\"slide")... )
    buffer = append(buffer, []byte(fmt.Sprintf("%d",counter))... )
    buffer = append(buffer, []byte("\"><div class=\"Stitle169\"><h1>")... )
    pageNum := fmt.Sprintf("%d. ",counter)
    buffer = append(buffer, []byte(pageNum)... )
    buffer = append(buffer, inlineStyles(self.title)... )
    buffer = append(buffer, []byte(`</h1></div><div class="Slogo"><img src="/logo.svg"/></div><div class="Sin169">`)... )
    switch self.layout {
        case "single":
            buffer = append(buffer, self.primary...)
        case "rows":
            buffer = append(buffer, []byte("<div style=\"width:100%; height:49%; display:inline-block\">")... )
            buffer = append(buffer, self.primary...)
            buffer = append(buffer, []byte("</div><div style=\"width:100%; height:49%; display:inline-block\">")... )
            buffer = append(buffer, self.secondary...)
            buffer = append(buffer, []byte("</div>")...)
        case "cols", "columns":
            buffer = append(buffer, []byte("<div style=\"width:49%;height:100%;display:inline-block;vertical-align:top\">")... )
            buffer = append(buffer, self.primary...)
            buffer = append(buffer, []byte("</div><div style=\"width:49%;height:100%;display:inline-block;margin-left:1%\">")... )
            buffer = append(buffer, self.secondary...)
            buffer = append(buffer, []byte("</div>")...)
    }
    buffer = append(buffer, []byte("</div>\n")...)
    if len(self.longform)>0 {
        buffer = append(buffer, []byte("<div class=\"flipicon\"><a onclick=\"javascript:flipPage(")...)
        buffer = append(buffer, []byte(fmt.Sprintf("%d",counter))...)
        buffer = append(buffer, []byte(")\"><img src=\"/fliparrow.jpg\"></img></a>></div>\n")...)
        buffer = append(buffer, []byte("\n<div class=\"longform\" id=\"slide")...)
        buffer = append(buffer, []byte(fmt.Sprintf("%d",counter))...)
        buffer = append(buffer, []byte("_long\">")...)
        buffer = append(buffer, self.longform...)
        buffer = append(buffer, []byte("</div>")...)
    } else {
        buffer = append(buffer, []byte("</div>\n")...)
    }
    return buffer
}


func renderHtmlSlides(headBlock Block, input chan Block) []byte {
    counter := 1
    layout  := "single"
    var target MultiChanSlide
    result := make([]byte, 0, 16384)
    result = append(result, makePageHeader(string(headBlock.style))...)
    result = append(result, []byte(`<div id="navpanel"><a><img src="/leftarrow.svg" class="icon" onclick="javascript:leftButton()" id="navleft"></img></a><a><img src="/rightarrow.svg" class="icon" onclick="javascript:rightButton()" id="navright"></img></a><a><img src="/closearrow.svg" class="icon" onclick="javascript:navcloseButton()" id="navclose"></img></a><button onclick="javascript:flipMode()">flip mode</button><button onclick="javascript:flipAspect()">flip aspect</button></div><a class="settings" onclick="javascript:settingsButton()"><img src="/settings.svg" class="settings"></img></a>`)...)
    result = append(result, []byte(`<div id="slides">`)...)
    // Title slide
    result = append(result, []byte(`<div class="S169"><div class="Slogo"><img src="/logo.svg"/></div><div class="Sin169">`)...)
    result = append(result, []byte("<h1>")... )
    result = append(result, inlineStyles(headBlock.courseCode)... )
    result = append(result, []byte("</h1>")... )
    result = append(result, []byte("<h1 style=\"margin-bottom:1.5em\">")... )
    result = append(result, inlineStyles(headBlock.courseName)... )
    result = append(result, []byte("</h1>")... )
    result = append(result, []byte("<i>")... )
    result = append(result, headBlock.author... )
    result = append(result, []byte("</i>")... )
    result = append(result, []byte("<p>")... )
    result = append(result, headBlock.date... )
    result = append(result, []byte("</p>")... )
    result = append(result, []byte("<i>")... )
    result = append(result, headBlock.location... )
    result = append(result, []byte("</i>")... )
    result = append(result, []byte("<h2 style=\"margin-top:1.5em\">")... )
    result = append(result, inlineStyles(headBlock.title)... )
    result = append(result, []byte("</h2>")... )
    result = append(result, []byte("</div></div>")... )

    lastTag  := ""
    lastKind := BlkParagraph
    for blk := range input {
        if blk.kind!=BlkTableRow && blk.kind!=BlkTableCell && lastKind==BlkTableCell {
            target.extendB( []byte("</tr>") )
        }
        if lastTag!="" && tagNames[blk.kind]!=lastTag {
            target.Printf( "</%s>", lastTag)
        }
        if tagNames[blk.kind]!="" && tagNames[blk.kind]!=lastTag {
            target.Printf( "<%s", tagNames[blk.kind] )
            if blk.kind==BlkTableRow {
                target.extendB([]byte(" class=\"allborders\" width=\"100%\""))
            }
            target.extendB([]byte(">"))
        }
        switch blk.kind {
            case BlkParagraph:
                target.Printf("<p>%s</p>", inlineStyles(blk.body))
            case BlkNumbered, BlkBulleted:
                target.Printf("<li>%s</li>", inlineStyles(blk.body))
            case BlkSmallHeading, BlkMediumHeading:
                if target.primary!=nil && target.active=="longform" {
                    target.Printf("<h2>%s</h2>\n",blk.body)
                } else {
                    if target.primary!=nil {
                        result = target.finalise(result,counter)
                        counter++
                    }
                    layout = string(blk.style)
                    target = makeMultiChanSlide(layout,blk.body)
                }
            case BlkShell:
                target.PrintfHL(bytes.Compare(blk.position,[]byte("highlight"))==0,
                                "<div class=\"shell\">%s</div>",
                                html.EscapeString(string(blk.body)) )
            case BlkCode:
                target.PrintfHL(bytes.Compare(blk.position,[]byte("highlight"))==0,
                                "<div class=\"code\">%s</div>",
                                html.EscapeString(string(blk.body)) )
            case BlkTopicBegin:
                target.Printf("<div class=\"Scallo\"><div class=\"ScalloHd\">%s</div>",
                              inlineStyles(blk.body) )
            case BlkTopicEnd:
                target.extendB( []byte("</div>") )
            case BlkQuote:
                target.Printf("<div class=\"quoteinside\"><div class=\"quotebegin\">&#8220;</div>%s",
                              inlineStyles(blk.body) )
                if len(blk.author)>0 {
                    target.Printf("<br/>--- %s", blk.author)
                }
                target.extendB( []byte("<div class=\"quoteend\">&#8221;</div></div>\n") )
            case BlkImage:
                target.PrintfHL(true,
                                "<img src=\"%s\" style=\"width:100%%; max-height:100%%; object-fit:contain\"/>",
                                blk.body)
                target.PrintfLong("<img src=\"%s\" style=\"width:25rem; max-height:25rem%%; border: 1px solid black; margin:1rem; object-fit:contain\"/>",
                                blk.body)
            case BlkVideo:
                target.Printf("<video width=\"100%%\" style=\"max-width:100%%; max-height:95%%\" controls>\n" +
                              "<source src=\"%s.webm\" type=\"video/webm;\">" +
                              "<source src=\"%s.mov\" type=\"video/quicktime;\"></video>",
                              blk.body, blk.body)
            case BlkReference:
                target.Printf("<div class=bibitem><table style=\"width=100%%\">\n" +
                              "<tr><td rowspan=\"3\"><img src=\"/book-icon.png\"/></td>\n" +
                              "<td><a href=\"%s\"></a></td></tr><tr><td><i>%s</i></td></tr>",
                              blk.url, blk.title, blk.author)
                if blk.detail!=nil {
                    target.Printf("<tr><td>%s</td></tr>", blk.detail)
                }
                target.extendB( []byte("</table></div>") )
            case BlkDefList:
                target.Printf("<dt>%s</dt><dd>%s</dd>\n",
                              blk.heading, inlineStyles(blk.body))
            case BlkTableRow:
                if lastKind==BlkTableCell {
                    target.extendB( []byte("</tr>") )
                }
                target.extendB( []byte("<tr>") )
            case BlkTableCell:
                target.Printf("<td>%s</td>", inlineStyles(blk.body) )
            case BlkBeginLongform:
                target.active = "longform"
            case BlkEndLongform:
                target.active = "primary"
            default:
                fmt.Println("Unknown Block:", blk)
        }
        lastTag  = tagNames[blk.kind]
        lastKind = blk.kind
    }
    result = target.finalise(result,counter)
    result = append(result, []byte("</div></div>")...)
    result = append(result, pageFooter...)
    return result
}


func RenderHtml(input chan Block) []byte {
    headBlock := <-input
    if headBlock.kind!=BlkBigHeading {
        var bstr string
        fmt.Sprintf(bstr,"%s",headBlock)
        panic("Parser is not sending the BigHeading first! "+bstr)
    }
    switch string(headBlock.style) {
        case "page":
            return renderHtmlPage(headBlock,input)
        case "slides":
            return renderHtmlSlides(headBlock,input)
        default:
            panic("Unknown style to render! "+string(headBlock.style))
    }

}


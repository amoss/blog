package rst

import (
    "fmt"
    "regexp"
    "bytes"
    "html"
    "strings"
)

func makePageHeader(extra string, insert []byte) []byte {
    result := make([]byte,0,1024)
    result = append(result,[]byte(`
<html>
<head>
<script type="text/x-mathjax-config">
  MathJax.Hub.Config({
    extensions: ["tex2jax.js"],
    jax: ["input/TeX", "output/HTML-CSS"],
    tex2jax: {
      inlineMath: [ ['$math','math$'], ['\\\\((','\\\\))'] ],
      displayMath: [ ['$math$','$math$'], ["\\[[","\\]]"] ],
      processEscapes: true
    },
    "HTML-CSS": { availableFonts: ["TeX"] }
  });
</script>
<script src="/MathJax/MathJax.js?config=TeX-AMS-MML_HTMLorMML" type="text/javascript">
</script>
<link href="../styles.css" type="text/css" rel="stylesheet"></link>
`)...)
    if extra != "" {
        result = append(result, []byte("<link href=\"")...)
        result = append(result, []byte("../"+extra)...)
        result = append(result, []byte(".css\" type=\"text/css\" rel=\"stylesheet\"></link>\n")...)
    }
    if extra=="slides" {
        result = append(result, []byte("<script src=\"../slides.js\" type=\"text/javascript\"></script>")...)
    }
    result = append(result, insert...)
    result = append(result, []byte(`</head>
<body>
`)...)
    return result
}

var pageFooter = []byte(`
</body>
</html>
`)

/* We need to avoid applying the links / strong / emp styles inside the literal
   environments, hence this simple state machine. Partition the input into
   alternating non-literal / literal pieces and apply the two mappings 
   alternately.
*/
func inlineStyles(input []byte) []byte {
    result := make([]byte,0,4096)
    links  := regexp.MustCompile("`([^`]+) <([^>]+)>`_")
    strong := regexp.MustCompile("\\*\\*([^*]+)\\*\\*")
    emp    := regexp.MustCompile(" \\*([^*]+)\\*")

    litEnvs   := regexp.MustCompile(":(shell|code|math):`[^`]+`")
    partition := litEnvs.FindAllIndex(input,-1)
    pos := 0
    for _,pair := range partition {
        nonLit := links.ReplaceAll(input[pos:pair[0]], []byte("<a href=\"$2\">$1</a>"))
        nonLit  = strong.ReplaceAll(nonLit,[]byte("<b>$1</b>"))
        nonLit  = emp.ReplaceAll(nonLit, []byte(" <i>$1</i>"))
        result = append(result, nonLit...)
        //fmt.Printf("Inline: %s %s\n", pair,result)
        switch input[ pair[0]+1 ] {
            case 'm':
                result = append(result, []byte(`\\((`)... )
                result = append(result, input[ pair[0]+7 : pair[1]-1 ]... )
                result = append(result, []byte(`\\))`)... )
            case 'c':
                result = append(result, []byte(`<span class="code">`)... )
                result = append(result, html.EscapeString(string(input[ pair[0]+7 : pair[1]-1 ]))... )
                result = append(result, []byte(`</span>`)... )
            case 's':
                result = append(result, []byte(`<span class="shell">`)... )
                result = append(result, html.EscapeString(string(input[ pair[0]+8 : pair[1]-1 ]))... )
                result = append(result, []byte(`</span>`)... )
        }
        pos = pair[1]+1
        //fmt.Printf("Inline2: %s %s\n", pair,result)
    }
    if pos<len(input) {
        nonLit := links.ReplaceAll(input[pos:], []byte("<a href=\"$2\">$1</a>"))
        nonLit  = strong.ReplaceAll(nonLit,[]byte("<b>$1</b>"))
        nonLit  = emp.ReplaceAll(nonLit, []byte(" <i>$1</i>"))
        result = append(result, nonLit...)
    }
    return result
}


/*func inlineStyles(input []byte) []byte {
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
}*/

var tagNames = map[BlockE]string {
    BlkBulleted: "ul",
    BlkNumbered: "ol",
    BlkDefList:  "dl",
    BlkTableRow: "table",
    BlkTableCell: "table" }



func renderHtmlPage(headBlock Block, input chan Block) []byte {
    result := make([]byte, 0, 16384)
    result = append(result, makePageHeader(string(headBlock.Style),[]byte(""))...)
    result = append(result, []byte("<div style=\"width:100%; background-color:#a7a8aa; padding:1.5rem\">")... )
    result = append(result, []byte("<h1>")... )
    result = append(result, inlineStyles(headBlock.Title)... )
    result = append(result, []byte("</h1>")... )
    result = append(result, []byte("<i>")... )
    result = append(result, headBlock.Author... )
    result = append(result, []byte("</i>")... )
    result = append(result, []byte("<p>")... )
    result = append(result, headBlock.Date... )
    result = append(result, []byte("</p>")... )
    result = append(result, []byte("</div>")... )
    result = append(result, []byte("<div style=\"width:45rem;max-width:45rem;display:table;margin:0 auto;background-color:#ffffff;padding:1.5rem\">")... )
    lastTag  := ""
    lastKind := BlkParagraph
    for blk := range input {
        if lastTag!="" && tagNames[blk.Kind]!=lastTag {
            result = append(result, []byte("</")... )
            result = append(result, []byte(tagNames[lastKind])... )
            result = append(result, []byte(">")... )
        }
        if tagNames[blk.Kind]!="" && tagNames[blk.Kind]!=lastTag {
            result = append(result, []byte("<")... )
            result = append(result, []byte(tagNames[blk.Kind])... )
            if blk.Kind==BlkTableRow {
                result = append(result, []byte(" class=\"allborders\" width=\"100%\"")...)
            }
            result = append(result, []byte(">")... )
        }
        if blk.Kind!=BlkTableRow && blk.Kind!=BlkTableCell && lastKind==BlkTableCell {
            result = append(result, []byte("</tr>")... )
        }
        switch blk.Kind {
            case BlkParagraph:
                result = append(result, []byte("<p>")... )
                result = append(result, inlineStyles(blk.Body)... )
                result = append(result, []byte("</p>")... )
            case BlkNumbered, BlkBulleted:
                result = append(result, []byte("<li>")... )
                result = append(result, inlineStyles(blk.Body)... )
                result = append(result, []byte("</li>")... )
            case BlkMediumHeading:
                result = append(result, []byte("<h1>")... )
                result = append(result, inlineStyles(blk.Body)... )
                result = append(result, []byte("</h1>")... )
            case BlkSmallHeading:
                result = append(result, []byte("<h2>")... )
                result = append(result, inlineStyles(blk.Body)... )
                result = append(result, []byte("</h2>")... )
            case BlkShell:
                result = append(result, []byte("<pre class=\"shell\">")... )
                result = append(result, []byte(html.EscapeString(string(blk.Body)))... )
                result = append(result, []byte("</pre>")... )
            case BlkCode:
                result = append(result, []byte("<pre class=\"code\"><table style=\"width: 100%;border-collapse: collapse\">")... )
                fmt.Print(string(blk.Body))
                for _,line := range strings.Split(string(blk.Body),"\n") {
                    result = append(result, []byte("<tr><td class=\"lnum\"></td><td class=\"content\">")... )
                    if len(line)>0 {
                        result = append(result, []byte(html.EscapeString(line))... )
                    } else {
                        result = append(result, byte('\n'))
                    }
                    result = append(result, []byte("</td></tr>")... )
                }
                result = append(result, []byte("</table></pre>")... )
            case BlkTopicBegin:
                result = append(result, []byte("<div class=\"Scallo\"><div class=\"ScalloHd\">")... )
                result = append(result, inlineStyles(blk.Body)... )
                result = append(result, []byte("</div>")...)
            case BlkTopicEnd:
                result = append(result, []byte("</div>")... )
            case BlkQuote:
                result = append(result, []byte("<div class=\"quoteinside\"><div class=\"quotebegin\">&#8220;</div>")... )
                result = append(result, inlineStyles(blk.Body)... )
                if len(blk.Author)>0 {
                    result = append(result, []byte("<br/>--- ")... )
                    result = append(result, blk.Author... )
                }
                result = append(result, []byte("<div class=\"quoteend\">&#8221;</div></div>\n")... )
            case BlkImage:
                result = append(result, []byte("<img src=\"")...)
                result = append(result, blk.Body... )
                result = append(result, []byte("\" style=\"width:100%; max-height:100%; object-fit:contain\"/>")...)
            case BlkVideo:
                result = append(result, []byte("<video width=\"100%%\" style=\"max-width:100%% max-height:95%%\" controls>\n")... )
                result = append(result, []byte("<source src=\"")... )
                result = append(result, blk.Body... )
                result = append(result, []byte(".webm\" type=\"video/webm;\">")...)
                result = append(result, []byte("<source src=\"")... )
                result = append(result, blk.Body... )
                result = append(result, []byte(".mov\" type=\"video/quicktime;\">")...)
                result = append(result, []byte("</video>")...)
            case BlkReference:
                result = append(result, []byte("<div class=bibitem><table style=\"width=100%%\">\n<tr><td rowspan=\"3\"><img style=\"width:2rem;height:2rem\" src=\"../book-icon.png\"/></td>\n<td><a href=\"")...)
                result = append(result, blk.Url... )
                result = append(result, []byte("\">")...)
                result = append(result, blk.Title... )
                result = append(result, []byte("</a></td></tr><tr><td><i>")...)
                result = append(result, blk.Author... )
                result = append(result, []byte("</i></td></tr>")...)
                if blk.Detail!=nil {
                    result = append(result, []byte("<tr><td>")...)
                    result = append(result, blk.Detail... )
                    result = append(result, []byte("</td></tr>")...)
                }
                result = append(result, []byte("</table></div>")...)
            case BlkDefList:
                result = append(result, []byte("<dt>")...)
                result = append(result, blk.Heading... )
                result = append(result, []byte("</dt><dd>")...)
                result = append(result, inlineStyles(blk.Body)... )
                result = append(result, []byte("</dd>\n")...)
            case BlkTableRow:
                if lastKind==BlkTableCell {
                    result = append(result, []byte("</tr>")... )
                }
                result = append(result, []byte("<tr>")... )
            case BlkTableCell:
                result = append(result, []byte("<td>")... )
                result = append(result, inlineStyles(blk.Body)... )
                result = append(result, []byte("</td>")... )
            default:
                fmt.Println("Block:", blk)
        }
        lastTag  = tagNames[blk.Kind]
        lastKind = blk.Kind
    }
    result = append(result, []byte("</div>")... )
    result = append(result, pageFooter...)
    return result
}

type MultiChanSlide struct {
    primary, secondary, longform []byte
    title  []byte
    active string       // primary, secondary, longform
    layout string       // single, rows or columns
    usedLongform bool
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
            case "rows", "cols", "columns":
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

func makeMultiChanSlide(layout string, title []byte, pagenum int) MultiChanSlide{
  result := MultiChanSlide{layout:layout, title:title}
  result.primary   = make([]byte,0,16384)
  result.secondary = make([]byte,0,16384)
  result.longform  = make([]byte,0,16384)
  result.PrintfLong( "<h1><a><img src=\"../flipbackarrow.jpg\" onclick=\"javascript:flipBackPage(%d)\"></img></a>%s</h1>",
                     pagenum, title)
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
    buffer = append(buffer, []byte(`</h1></div><div class="Slogo"><img src="../logo.svg"/></div><div class="Sin169">`)... )
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
    if self.usedLongform {
        buffer = append(buffer, []byte("<div class=\"flipicon\"><a onclick=\"javascript:flipPage(")...)
        buffer = append(buffer, []byte(fmt.Sprintf("%d",counter))...)
        buffer = append(buffer, []byte(")\"><img src=\"../fliparrow.jpg\"></img></a></div>\n")...)
        buffer = append(buffer, []byte("\n</div><div class=\"longform\" id=\"slide")...)
        buffer = append(buffer, []byte(fmt.Sprintf("%d",counter))...)
        buffer = append(buffer, []byte("_long\">")...)
        buffer = append(buffer, self.longform...)
        buffer = append(buffer, []byte("</div>")...)
    } else {
        buffer = append(buffer, []byte("</div>\n")...)
    }
    return buffer
}


func renderHtmlSlides(headBlock Block, input chan Block, headerInsert []byte) []byte {
    counter := 1
    layout  := "single"
    var target MultiChanSlide
    result := make([]byte, 0, 16384)
    result = append(result, makePageHeader(string(headBlock.Style),headerInsert)...)
    result = append(result, []byte(`<div id="navpanel"><a><img src="../leftarrow.svg" class="icon" onclick="javascript:leftButton()" id="navleft"></img></a><a><img src="../rightarrow.svg" class="icon" onclick="javascript:rightButton()" id="navright"></img></a><a><img src="../closearrow.svg" class="icon" onclick="javascript:navcloseButton()" id="navclose"></img></a><button onclick="javascript:flipMode()">flip mode</button><button onclick="javascript:flipAspect()">flip aspect</button></div><a class="settings" onclick="javascript:settingsButton()"><img src="../settings.svg" class="settings"></img></a>`)...)
    result = append(result, []byte(`<div id="slides" style="margin-top:50%%; margin-bottom:50%%">`)...)
    // Title slide
    result = append(result, []byte(`<div class="S169"><div class="Slogo"><img src="../logo.svg"/></div><div class="Sin169">`)...)
    result = append(result, []byte("<h1>")... )
    result = append(result, inlineStyles(headBlock.CourseCode)... )
    result = append(result, []byte("</h1>")... )
    result = append(result, []byte("<h1 style=\"margin-bottom:1.5em\">")... )
    result = append(result, inlineStyles(headBlock.CourseName)... )
    result = append(result, []byte("</h1>")... )
    result = append(result, []byte("<i>")... )
    result = append(result, headBlock.Author... )
    result = append(result, []byte("</i>")... )
    result = append(result, []byte("<p>")... )
    result = append(result, headBlock.Date... )
    result = append(result, []byte("</p>")... )
    result = append(result, []byte("<i>")... )
    result = append(result, headBlock.Location... )
    result = append(result, []byte("</i>")... )
    result = append(result, []byte("<h2 style=\"margin-top:1.5em\">")... )
    result = append(result, inlineStyles(headBlock.Title)... )
    result = append(result, []byte("</h2>")... )
    result = append(result, []byte("</div></div>")... )

    lastTag  := ""
    lastKind := BlkParagraph
    for blk := range input {
        if LineParserDbg {
           fmt.Printf("%29s: %s\n", "block", blk)

        }
        if blk.Kind!=BlkTableRow && blk.Kind!=BlkTableCell && lastKind==BlkTableCell {
            target.extendB( []byte("</tr>") )
        }
        if lastTag!="" && tagNames[blk.Kind]!=lastTag {
            target.Printf( "</%s>", lastTag)
        }
        if tagNames[blk.Kind]!="" && tagNames[blk.Kind]!=lastTag {
            target.Printf( "<%s", tagNames[blk.Kind] )
            if blk.Kind==BlkTableRow {
                target.extendB([]byte(" class=\"allborders\" width=\"100%\""))
            }
            target.extendB([]byte(">"))
        }
        switch blk.Kind {
            case BlkParagraph:
                target.Printf("<p>%s</p>", inlineStyles(blk.Body))
            case BlkNumbered, BlkBulleted:
                target.Printf("<li>%s</li>", inlineStyles(blk.Body))
            case BlkSmallHeading, BlkMediumHeading:
                if target.primary!=nil && target.active=="longform" {
                    target.Printf("<h2>%s</h2>\n",blk.Body)
                } else {
                    if target.primary!=nil {
                        result = target.finalise(result,counter)
                        counter++
                    }
                    layout = string(blk.Style)
                    target = makeMultiChanSlide(layout,blk.Body,counter)
                }
            case BlkShell:
                var styles string
                if len(blk.Style)>0 {
                    styles = " style=\"" + string(blk.Style) + "\""
                } else {
                    styles = ""
                }
                target.PrintfHL(bytes.Compare(blk.Position,[]byte("highlight"))==0,
                                "<div class=\"shell\"%s>%s</div>", styles,
                                html.EscapeString(string(blk.Body)) )
            case BlkCode:
                if target.primary!=nil && target.active=="longform" {
                    target.Printf( "<div class=\"code\">%s</div>",
                                    html.EscapeString(string(blk.Body)) )
                } else {
                    target.PrintfHL(bytes.Compare(blk.Position,[]byte("highlight"))==0,
                                    "<div class=\"code\">%s</div>",
                                    html.EscapeString(string(blk.Body)) )
                }
            case BlkTopicBegin:
                target.Printf("<div class=\"Scallo\"><div class=\"ScalloHd\">%s</div>",
                              inlineStyles(blk.Body) )
            case BlkTopicEnd:
                target.extendB( []byte("</div>") )
            case BlkQuote:
                target.Printf("<div class=\"quoteinside\"><div class=\"quotebegin\">&#8220;</div>%s",
                              inlineStyles(blk.Body) )
                if len(blk.Author)>0 {
                    target.Printf("<br/>--- %s", blk.Author)
                }
                target.extendB( []byte("<div class=\"quoteend\">&#8221;</div></div>\n") )
            case BlkImage:
                target.PrintfHL(true,
                                "<img src=\"%s\" style=\"width:100%%; max-height:100%%; object-fit:contain\"/>",
                                blk.Body)
                target.PrintfLong("<img src=\"%s\" style=\"width:25rem; max-height:25rem%%; border: 1px solid black; margin:1rem; object-fit:contain\"/>",
                                blk.Body)
            case BlkVideo:
                target.Printf("<video width=\"100%%\" style=\"max-width:100%%; max-height:95%%\" controls>\n" +
                              "<source src=\"%s.webm\" type=\"video/webm;\">" +
                              "<source src=\"%s.mov\" type=\"video/quicktime;\"></video>",
                              blk.Body, blk.Body)
            case BlkReference:
                target.Printf("<div class=bibitem><table style=\"width=100%%\">\n" +
                              "<tr><td rowspan=\"3\"><img src=\"../book-icon.png\"/></td>\n" +
                              "<td><a href=\"%s\">%s</a></td></tr><tr><td><i>%s</i></td></tr>",
                              blk.Url, blk.Title, blk.Author)
                if blk.Detail!=nil {
                    target.Printf("<tr><td>%s</td></tr>", blk.Detail)
                }
                target.extendB( []byte("</table></div>") )
            case BlkDefList:
                target.Printf("<dt>%s</dt><dd>%s</dd>\n",
                              blk.Heading, inlineStyles(blk.Body))
            case BlkTableRow:
                if lastKind==BlkTableCell {
                    target.extendB( []byte("</tr>") )
                }
                target.extendB( []byte("<tr>") )
            case BlkTableCell:
                target.Printf("<td>%s</td>", inlineStyles(blk.Body) )
            case BlkBeginLongform:
                target.active = "longform"
                target.usedLongform = true
            case BlkEndLongform:
                target.active = "primary"
            default:
                fmt.Println("Unknown Block:", blk)
        }
        lastTag  = tagNames[blk.Kind]
        lastKind = blk.Kind
    }
    result = target.finalise(result,counter)
    result = append(result, []byte(`
</div><div id="vidarea" style="position:fixed; left:66%; width:34%; height:100%">
<div id="vidholder" style="max-width:100%; max-height:95%; margin-top:50%; margin-bottom:50%">
<video id="vid" width="100%" style="max-width:100%; max-height:95%" controls>\n
<source src="vid.ogv" type="video/ogg;">
<source src="vid.webm" type="video/webm;">
<source src="vid.mp4" type="video/mp4;"></video>
</div></div>`)...)
    result = append(result, pageFooter...)
    return result
}


func RenderHtml(input chan Block) []byte {
    headBlock := <-input
    if headBlock.Kind!=BlkBigHeading {
        var bstr string
        fmt.Sprintf(bstr,"%s",headBlock)
        panic("Parser is not sending the BigHeading first! "+bstr)
    }
    switch string(headBlock.Style) {
        case "page":
            return renderHtmlPage(headBlock,input)
        case "slides":
            return renderHtmlSlides(headBlock,input,[]byte(""))
        default:
            panic("Unknown style to render! "+string(headBlock.Style))
    }

}


func RenderHtmlWithHeader(input chan Block, headerInsert []byte) []byte {
    headBlock := <-input
    if headBlock.Kind!=BlkBigHeading {
        var bstr string
        fmt.Sprintf(bstr,"%s",headBlock)
        panic("Parser is not sending the BigHeading first! "+bstr)
    }
    switch string(headBlock.Style) {
        case "page":
            return renderHtmlPage(headBlock,input)
        case "slides":
            return renderHtmlSlides(headBlock,input,headerInsert)
        default:
            panic("Unknown style to render! "+string(headBlock.Style))
    }

}


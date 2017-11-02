package main

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
  emp    := regexp.MustCompile("\\*([^*]+)\\*")
  input   = emp.ReplaceAll(input, []byte("<i>$1</i>"))
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


func renderHtmlSlides(headBlock Block, input chan Block) []byte {
    counter := 1
    layout  := "single"
    var other []byte
    result := make([]byte, 0, 16384)
    result = append(result, makePageHeader(string(headBlock.style))...)
    result = append(result, []byte(`<div id="navpanel"><a><img src="/leftarrow.svg" class="icon" onclick="javascript:leftButton()" id="navleft"></img></a><a><img src="/rightarrow.svg" class="icon" onclick="javascript:rightButton()" id="navright"></img></a><a><img src="/closearrow.svg" class="icon" onclick="javascript:navcloseButton()" id="navclose"></img></a><button onclick="javascript:flipMode()">flip mode</button></div><a class="settings" onclick="javascript:settingsButton()"><img src="/settings.svg" class="settings"></img></a>`)...)
    result = append(result, []byte(`<div id="slides">`)...)
    // Title slide
    result = append(result, []byte(`<div class="S169"><div class="Sin169">`)...)
    result = append(result, []byte("<h1>")... )
    result = append(result, inlineStyles(headBlock.title)... )
    result = append(result, []byte("</h1>")... )
    result = append(result, []byte("<i>")... )
    result = append(result, headBlock.author... )
    result = append(result, []byte("</i>")... )
    result = append(result, []byte("<p>")... )
    result = append(result, headBlock.date... )
    result = append(result, []byte("</p>")... )
    result = append(result, []byte("</div></div>")... )

    lastTag  := ""
    lastKind := BlkParagraph
    for blk := range input {
        if blk.kind!=BlkTableRow && blk.kind!=BlkTableCell && lastKind==BlkTableCell {
            result = append(result, []byte("</tr>")... )
        }
        if lastTag!="" && tagNames[blk.kind]!=lastTag {
            result = append(result, []byte("</")... )
            result = append(result, []byte(lastTag)... )
            result = append(result, []byte(">")... )
        }
        if tagNames[blk.kind]!="" && tagNames[blk.kind]!=lastTag {
            result = append(result, []byte("<")... )
            result = append(result, []byte(tagNames[blk.kind])... )
            if blk.kind==BlkTableRow {
                result = append(result, []byte(" class=\"allborders\" width=\"100%\"")...)
            }
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
            case BlkSmallHeading, BlkMediumHeading:
                switch layout {
                    case "single":
                    case "rows":
                        result = append(result, []byte("</div><div style=\"width:100%; height:49%; display:inline-block\">")... )
                        result = append(result, other...)
                        result = append(result, []byte("</div>")...)
                    case "cols":
                        result = append(result, []byte("</div><div style=\"width:49%;height:100%;display:inline-block;margin-left:1%\">")... )
                        result = append(result, other...)
                        result = append(result, []byte("</div>")...)
                }
                result = append(result, []byte("</div></div>\n<div class=\"S169\"><div class=\"Stitle169\"><h1>")... )
                pageNum := fmt.Sprintf("%d. ",counter)
                counter++
                result = append(result, []byte(pageNum)... )
                result = append(result, inlineStyles(blk.body)... )
                result = append(result, []byte(`</h1></div><div class="Slogo"><img src="/logo.svg"/></div><div class="Sin169">`)... )
                layout = string(blk.style)
                switch layout {
                    case "single":
                    case "rows":
                        other = make([]byte, 0, 16384)
                        result = append(result, []byte("<div style=\"width:100%; height:49%; display:inline-block\">")... )
                    case "cols":
                        other = make([]byte, 0, 16384)
                        result = append(result, []byte("<div style=\"width:49%;height:100%;display:inline-block;vertical-align:top\">")... )
                }
            case BlkShell:
                if layout!="single" &&
                   bytes.Compare(blk.position,[]byte("highlight"))==0 {
                    other = append(other, []byte("<div class=\"shell\">")... )
                    other = append(other, []byte(html.EscapeString(string(blk.body)))... )
                    other = append(other, []byte("</div>")... )
                } else {
                    result = append(result, []byte("<div class=\"shell\">")... )
                    result = append(result, []byte(html.EscapeString(string(blk.body)))... )
                    result = append(result, []byte("</div>")... )
                }
            case BlkCode:
                if layout!="single" &&
                   bytes.Compare(blk.position,[]byte("highlight"))==0 {
                    other = append(other, []byte("<div class=\"code\">")... )
                    other = append(other, []byte(html.EscapeString(string(blk.body)))... )
                    other = append(other, []byte("</div>")... )
                } else {
                    result = append(result, []byte("<div class=\"code\">")... )
                    result = append(result, []byte(html.EscapeString(string(blk.body)))... )
                    result = append(result, []byte("</div>")... )
                }
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
                if layout=="single" {
                    result = append(result, []byte("<img src=\"")...)
                    result = append(result, blk.body... )
                    result = append(result, []byte("\" style=\"width:100%; max-height:100%; object-fit:contain\"/>")...)
                } else {
                    other = append(other, []byte("<img width=\"100%%\" src=\"")...)
                    other = append(other, blk.body... )
                    other = append(other, []byte("\" style=\"width:100%; max-height:100%; object-fit:contain\"/>")...)
                }
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
                result = append(result, []byte("<div class=bibitem><table style=\"width=100%%\">\n<tr><td rowspan=\"3\"><img src=\"/book-icon.png\"/></td>\n<td><a href=\"")...)
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
            case BlkTableRow:
                if lastKind==BlkTableCell {
                    result = append(result, []byte("</tr>")... )
                }
                result = append(result, []byte("<tr>")...)
            case BlkTableCell:
                result = append(result, []byte("<td>")...)
                result = append(result, inlineStyles(blk.body)...)
                result = append(result, []byte("</td>")...)
            default:
                fmt.Println("Unknown Block:", blk)
        }
        lastTag  = tagNames[blk.kind]
        lastKind = blk.kind
    }
    result = append(result, []byte("</div></div>")...)
    result = append(result, pageFooter...)
    return result
}


func renderHtml(input chan Block) []byte {
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


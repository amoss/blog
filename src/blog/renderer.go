package main

import (
    "fmt"
    "regexp"
    "html"
    "rst"
    "time"
)


var PageHeader = []byte(`<!DOCTYPE html>
<html lang="en"><head>
<link href="/awmblog/styles.css" type="text/css" rel="stylesheet"/>
</head>
<body> 
`)

var PageFooter = []byte(`
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
        fmt.Printf("Inline: %s %s\n", pair,result)
        switch input[ pair[0]+1 ] {
            case 'm':
                result = append(result, []byte(`\\((`)... )
                result = append(result, input[ pair[0]+7 : pair[1]-1 ]... )
                result = append(result, []byte(`\\))`)... )
            case 'c':
                result = append(result, []byte(`<span class="code">`)... )
                result = append(result, input[ pair[0]+7 : pair[1]-1 ]... )
                result = append(result, []byte(`</span>`)... )
            case 's':
                result = append(result, []byte(`<span class="shell">`)... )
                result = append(result, input[ pair[0]+8 : pair[1]-1 ]... )
                result = append(result, []byte(`</span>`)... )
        }
        pos = pair[1]+1
        fmt.Printf("Inline2: %s %s\n", pair,result)
    }
    if pos<len(input) {
        nonLit := links.ReplaceAll(input[pos:], []byte("<a href=\"$2\">$1</a>"))
        nonLit  = strong.ReplaceAll(nonLit,[]byte("<b>$1</b>"))
        nonLit  = emp.ReplaceAll(nonLit, []byte(" <i>$1</i>"))
        result = append(result, nonLit...)
    }
    return result
}


var tagNames = map[rst.BlockE]string {
    rst.BlkBulleted: "ul",
    rst.BlkNumbered: "ol",
    rst.BlkDefList:  "dl",
    rst.BlkTableRow: "table",
    rst.BlkTableCell: "table" }



func renderHtml(headBlock rst.Block, input chan rst.Block, showDrafts bool) []byte {
    result := make([]byte, 0, 16384)
    result = append(result, PageHeader...)
    result = append(result, []byte(`
<div class="wblock">
    <div style="color:white; opacity:1; margin-top:1rem; margin-bottom:1rem">
    <h1>Avoiding The Needless Multiplication Of Forms</h1>
    </div>
</div>
`)...)

    _,err := time.Parse("2006-01-02",string(headBlock.Date))
    if err!=nil && !showDrafts {
      return []byte("Good things comes to those who wait.")
    }

    result = append(result, []byte("<div class=\"rblock\" style=\"color:#c8c7ac;text-align:right\">")... )
    result = append(result, []byte("<h2 style=\"text-align:right\">")... )
    result = append(result, inlineStyles(headBlock.Title)... )
    result = append(result, []byte("</h2>")... )
    result = append(result, []byte("<i>")... )
    result = append(result, headBlock.Author... )
    result = append(result, []byte("</i>")... )
    result = append(result, []byte("<p>")... )
    result = append(result, headBlock.Date... )
    result = append(result, []byte("</p>")... )
    result = append(result, []byte("</div>")... )
    lastKind := rst.BlkParagraph
    for blk := range input {
        if tagNames[lastKind]!="" && tagNames[blk.Kind]!=tagNames[lastKind] {
            result = append(result, []byte("</")... )
            result = append(result, []byte(tagNames[lastKind])... )
            result = append(result, []byte("></div></div>")... )
        }
        if tagNames[blk.Kind]!="" && tagNames[blk.Kind]!=tagNames[lastKind] {
            result = append(result, []byte("<div class=\"pblock\"><div class=\"pinner\"><")... )
            result = append(result, []byte(tagNames[blk.Kind])... )
            result = append(result, []byte(">")... )
        }
        switch blk.Kind {
            case rst.BlkParagraph:
                result = append(result, []byte("<div class=\"pblock\"><div class=\"pinner\">")... )
                result = append(result, inlineStyles(blk.Body)... )
                result = append(result, []byte("</div></div>")... )
            case rst.BlkNumbered, rst.BlkBulleted:
                result = append(result, []byte("<li>")... )
                result = append(result, inlineStyles(blk.Body)... )
                result = append(result, []byte("</li>")... )
            case rst.BlkMediumHeading:
                result = append(result, []byte("<div class=\"pblock\"><div class=\"pinner\"><h2>")... )
                result = append(result, inlineStyles(blk.Body)... )
                result = append(result, []byte("</h2></div></div>")... )
            case rst.BlkSmallHeading:
                result = append(result, []byte("<div class=\"pblock\"><h3>")... )
                result = append(result, inlineStyles(blk.Body)... )
                result = append(result, []byte("</h3></div>")... )
            case rst.BlkShell:
                result = append(result, []byte("<div class=\"wblock\"><div class=\"shell\">")... )
                result = append(result, []byte(html.EscapeString(string(blk.Body)))... )
                result = append(result, []byte("</div></div>")... )
            case rst.BlkCode:
                escaped := html.EscapeString(string(blk.Body))
                result = append(result, []byte("<div class=\"rblock\"><div class=\"code\">")... )
                result = append(result, []byte(escaped)... )
                result = append(result, []byte("</div></div>")... )
            case rst.BlkTopicBegin:
                result = append(result, []byte("<div class=\"Scallo\"><div class=\"ScalloHd\">")... )
                result = append(result, inlineStyles(blk.Body)... )
                result = append(result, []byte("</div>")...)
            case rst.BlkTopicEnd:
                result = append(result, []byte("</div>")... )
            case rst.BlkQuote:
                result = append(result, []byte("<div class=\"rblock\"><div class=\"quoteinside\"><div class=\"quotebegin\">&#8220;</div>")... )
                result = append(result, inlineStyles(blk.Body)... )
                if len(blk.Author)>0 {
                    result = append(result, []byte("<br/>--- ")... )
                    result = append(result, blk.Author... )
                }
                result = append(result, []byte("<div class=\"quoteend\">&#8221;</div></div></div>\n")... )
            case rst.BlkImage:
                result = append(result, []byte("<div class=\"rblock\"><div class=\"pinner\"><img src=\"")...)
                result = append(result, blk.Body... )
                result = append(result, []byte("\" style=\"width:100%; max-height:100%; object-fit:contain\"/></div></div>")...)
            case rst.BlkVideo:
                result = append(result, []byte("<video width=\"100%%\" style=\"max-width:100%% max-height:95%%\" controls>\n")... )
                result = append(result, []byte("<source src=\"")... )
                result = append(result, blk.Body... )
                result = append(result, []byte(".webm\" type=\"video/webm;\">")...)
                result = append(result, []byte("<source src=\"")... )
                result = append(result, blk.Body... )
                result = append(result, []byte(".mov\" type=\"video/quicktime;\">")...)
                result = append(result, []byte("</video>")...)
            case rst.BlkReference:
                result = append(result, []byte("<div class=bibitem><table style=\"width=100%%\">\n<tr><td rowspan=\"3\"><img style=\"width:2rem;height:2rem\" src=\"/book-icon.png\"/></td>\n<td><a href=\"")...)
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
            case rst.BlkDefList:
                result = append(result, []byte("<dt>")...)
                result = append(result, blk.Heading... )
                result = append(result, []byte("</dt><dd>")...)
                result = append(result, inlineStyles(blk.Body)... )
                result = append(result, []byte("</dd>\n")...)
            case rst.BlkTableRow:
                if lastKind==rst.BlkTableCell {
                    result = append( result, []byte("</tr>")... )
                }
                result = append(result, []byte("<tr>")... )
            case rst.BlkTableCell:
                result = append(result, []byte("<td>")... )
                result = append(result, inlineStyles(blk.Body)... )
                result = append(result, []byte("</td>")... )
            default:
                fmt.Println("Block:", blk)
        }
        lastKind = blk.Kind
    }
    result = append(result, PageFooter...)
    return result
}

func RenderHtml(input chan rst.Block, showDrafts bool) []byte {
    headBlock := <-input
    if headBlock.Kind!=rst.BlkBigHeading {
        var bstr string
        fmt.Sprintf(bstr,"%s",headBlock)
        panic("Parser is not sending the BigHeading first! "+bstr)
    }
    return renderHtml(headBlock,input,showDrafts)
}


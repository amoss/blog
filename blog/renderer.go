package main

import (
    "fmt"
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
        result = append(result, []byte(extra)...)
        result = append(result, []byte(".css\" type=\"text/css\" rel=\"stylesheet\"></link>\n")...)
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

var tagNames = map[BlockE]string {
    BlkBulleted: "ul",
    BlkNumbered: "ol",
    BlkDefList:  "dl" }

func renderHtml(input chan Block) []byte {
    result := make([]byte, 0, 16384)
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
                result = append(result, blk.body... )
                result = append(result, []byte("</p>")... )
            case BlkNumbered, BlkBulleted:
                result = append(result, []byte("<li>")... )
                result = append(result, blk.body... )
                result = append(result, []byte("</li>")... )
            case BlkBigHeading:     // Assume this is first in stream
                result = append(result, makePageHeader(string(blk.style))...)
                result = append(result, []byte("<div style=\"width:100%; background-color:#dddddd; padding:1rem\">")... )
                result = append(result, []byte("<h1>")... )
                result = append(result, blk.title... )
                result = append(result, []byte("</h1>")... )
                result = append(result, []byte("<i>")... )
                result = append(result, blk.author... )
                result = append(result, []byte("</i>")... )
                result = append(result, []byte("<p>")... )
                result = append(result, blk.date... )
                result = append(result, []byte("</p>")... )
                result = append(result, []byte("</div>")... )
            case BlkMediumHeading:
                result = append(result, []byte("<h2>")... )
                result = append(result, blk.body... )
                result = append(result, []byte("</h2>")... )
            case BlkSmallHeading:
                result = append(result, []byte("<h3>")... )
                result = append(result, blk.body... )
                result = append(result, []byte("</h3>")... )
            case BlkShell:
                result = append(result, []byte("<div class=\"shell\">")... )
                result = append(result, blk.body... )
                result = append(result, []byte("</div>")... )
            case BlkCode:
                result = append(result, []byte("<div class=\"code\">")... )
                result = append(result, blk.body... )
                result = append(result, []byte("</div>")... )
            case BlkTopicBegin:
                result = append(result, []byte("<div class=\"Scallo\"><div class=\"ScalloHd\">")... )
                result = append(result, blk.body... )
                result = append(result, []byte("</div>")...)
            case BlkTopicEnd:
                result = append(result, []byte("</div>")... )
            case BlkQuote:
                result = append(result, []byte("<div class=\"quoteinside\"><div class=\"quotebegin\">&#8220;</div>")... )
                result = append(result, blk.body... )
                result = append(result, []byte(" - ")... )
                result = append(result, blk.author... )
                result = append(result, []byte("<div class=\"quoteend\">&#8221;</div></div>\n")... )
            case BlkImage:
                result = append(result, []byte("<img src=\"")...)
                result = append(result, blk.body... )
                result = append(result, []byte("\" />")...)
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
                result = append(result, []byte("<div class=bibitem><table style=\"width=100%%\"><tr><td rowspan=\"2\"><img src=\"book.icon.png\"/><a url=\"")...)
                result = append(result, blk.url... )
                result = append(result, []byte("\">")...)
                result = append(result, blk.title... )
                result = append(result, []byte("</a></td></tr>")...)
                result = append(result, blk.author... )
                result = append(result, []byte("</table></div>")...)
            case BlkDefList:
                result = append(result, []byte("<dt>")...)
                result = append(result, blk.heading... )
                result = append(result, []byte("</dt><dd>")...)
                result = append(result, blk.body... )
                result = append(result, []byte("</dd>\n")...)
            default:
                fmt.Println("Block:", blk)
        }
        lastKind = blk.kind
    }
    result = append(result, pageFooter...)
    return result
}


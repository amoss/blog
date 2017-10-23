package main

import (
    "fmt"
)

var tagNames = map[BlockE]string {
    BlkBulleted: "ul",
    BlkNumbered: "ol",
    BlkDefList:  "dl" }

func renderHtml(input chan Block) {
    lastKind := BlkParagraph
    for blk := range input {
        if tagNames[lastKind]!="" && blk.kind!=lastKind {
            fmt.Printf("</%s>", tagNames[lastKind])
        }
        if tagNames[blk.kind]!="" && blk.kind!=lastKind {
            fmt.Printf("<%s>", tagNames[blk.kind])
        }
        switch blk.kind {
            case BlkParagraph:
                fmt.Printf("<p>%s</p>\n", blk.body)
            case BlkNumbered, BlkBulleted:
                fmt.Printf("<li>%s</li>\n", blk.body)
            case BlkBigHeading:
                fmt.Printf("<h1>%s</h1>\n", blk.body)
            case BlkMediumHeading:
                fmt.Printf("<h2>%s</h2>\n", blk.body)
            case BlkShell:
                fmt.Printf("<div style=\"shell\">%s</div>", blk.body)
            case BlkCode:
                fmt.Printf("<div style=\"code\">%s</div>", blk.body)
            case BlkTopicBegin:
                fmt.Printf("<Scallo><div class=\"ScalloHd\">%s</div>", blk.body)
            case BlkTopicEnd:
                fmt.Printf("</Scallo>")
            case BlkQuote:
                fmt.Printf("<div class=\"quoteinside\"><div class=\"quotebegin\">&#8220;</div>%s - %s<div class=\"quoteend\">&#8221;</div></div>\n", blk.body, blk.author)
            case BlkImage:
                fmt.Printf("<img src=\"%s\" />", blk.body)
            case BlkVideo:
                fmt.Printf("<video width=\"100%%\" style=\"max-width:100%% max-height:95%%\" controls>\n")
                fmt.Printf("<source src=\"%s.webm\" type=\"video/webm;\">",blk.body)
                fmt.Printf("<source src=\"%s.mov\" type=\"video/quicktime;\">",blk.body)
                fmt.Printf("</video>")
            case BlkReference:
                fmt.Printf("<div class=bibitem><table style=\"width=100%%\"><tr><td rowspan=\"2\"><img src=\"book.icon.png\"/><a url=\"%s\">%s</a></td></tr>%s</table></div>", blk.url, blk.title, blk.author)
            case BlkSmallHeading:
                fmt.Printf("<h3>%s</h3>", blk.body)
            case BlkDefList:
                fmt.Printf("<dt>%s</dt><dd>%s</dd>\n", blk.heading, blk.body)
            default:
                fmt.Println("Block:", blk)
        }
        lastKind = blk.kind
    }
    _ = lastKind
}


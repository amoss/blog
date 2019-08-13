package main

var CommentDemo = []byte(`<div class="wblock"><a href="javascript:showDemo()">Markdown syntax...</a></div>
                          <div class="wblock" id="commentDemo" style="visibility: hidden; display: none">
                <div style="display: inline-block; width:45%; float:left; background: #cccccc; white-space:pre; color: #444444; font-family: 'monospace'">
Markdown syntax is based on RST: 
*italics* **bold** are inline styles. 
Paragraphs wrap

until blank lines separate. Other inline 
styles are ` +
":code:`x=y+z`, :shell:`ls` and " + `
` + ":math:`x^n=y^n+z^n`. Links are " + `
` + "`here <example.com>`_ " + `and blocks 
require blank lines

.. code::

  x <- f(y)
  // Until indent change

* Bullets are not indented
* Listed between blank lines</div><div class="comment" style="display: inline-block; width:45%; margin-left:8%; float;right; border:1px solid #666666"><div class="pblock"><div class="pinner">Markdown syntax is based on RST: <i>italics</i> <b>bold</b> are inline styles. Paragraphs wrap</div></div><div class="pblock"><div class="pinner">until blank lines separate. Other inline styles are <span class="code">x=y+z</span> <span class="shell">ls</span>and \\((x^n=y^n+z^n\\)) Links are <a href="example.com">here</a> and blocks require blank lines</div></div><div class="rblock"><div class="code">x &lt;- f(y)
// Until indent change</div></div><div class="pblock"><div class="pinner"><ul><li>Bullets are not indented</li><li>Listed between blank lines</li></ul></div></div></div><div style="clear:both"></div></div>`)


func CommentEditor(session *Session) []byte {
    return []byte(`<div class="wblock" style="height:auto">
                <div style="display: inline-block; width:45%; float:left">
                <textarea id="comment" style="height:100%; width:100%; resize:none" oninput='commentUpdate(this)'></textarea>
                <form action="javascript:submitComment()">
                    <input type="submit" value="Add Comment" id="submitButton" />
                </form>
            </div>
            
            <div id="comPreview" class="comment" style="display: inline-block; width:45%; margin-left:8%; float;right; border:1px solid #666666">
            </div><div style="clear:both"></div>
</div>`)
}
                /*<form>
                <input type="textarea" name="comment" rows="10" style="height:100%"/>
                </form>*/

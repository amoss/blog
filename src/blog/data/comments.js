let previewSocket = ""

function submitComment()
{
    var src = document.getElementById('comment');
    var but = document.getElementById('submitButton');
    if( previewSocket instanceof WebSocket && previewSocket.readyState===1) 
    {
        previewSocket.send( JSON.stringify({"action":"post","url":document.URL, "body":src.value}) );
        but.disabled = "disable";
        but.value    = "Submitting...";
    }
}


function errorEvent(ev)
{
    console.log("preview socket err: " + ev)
}


function openEvent(ev)
{
    console.log("preview socket open: " + ev)
    previewSocket.send(document.cookie)
    commentUpdate()
}


function readEvent(ev)
{
    document.getElementById('comPreview').innerHTML = ev.data;
}


function commentUpdate()
{
    src = document.getElementById('comment')
    src.style.height = "";
    src.style.height = src.scrollHeight + "px";
    if( !(previewSocket instanceof WebSocket) ) 
    {
        previewSocket = new WebSocket("ws://127.0.0.1:8080/awmblog/preview");
      //  previewSocket = new WebSocket("ws://mechani.se/awmblog/preview");
        previewSocket.onerror = errorEvent;
        previewSocket.onopen  = openEvent;
        previewSocket.onmessage = readEvent;
    }
    if( previewSocket instanceof WebSocket && previewSocket.readyState===1) 
    {
        previewSocket.send( JSON.stringify({"action":"preview","body":src.value}) );
    }
}

function showDemo()
{
    demo = document.getElementById("commentDemo")
    if( demo.style.visibility=='hidden' )
    {
        demo.style.display = 'block'
        demo.style.visibility = 'visible'
    }
    else
    {
        demo.style.display = 'none'
        demo.style.visibility = 'hidden'
    }
}

var numThumbs = 9;
var winSize = Math.floor(numThumbs/2);
var mainScale = 1.0;
var thumbs = []
var showSlides = false;
var currentSlide = undefined;


function calcAspect()
{
var sWidth  = slides[0].offsetWidth;
var sHeight = slides[0].offsetHeight
  console.log('Slide internal coords: ', sWidth, 'x', sHeight, ' ratio ',
              sWidth/sHeight);
  console.log('Window: ', window.innerWidth, 'x', window.innerHeight, ' ratio ',
              window.innerWidth/window.innerHeight);

  var wScale = window.innerWidth / (sWidth + 10.0);
  var hScale = window.innerHeight / (sHeight + 10.0);
  mainScale = Math.min(wScale,hScale);
  console.log('Scaling: ', mainScale);
}

function doScale(obj, s)
{
  obj.style.transform       = 'scale(' + s + ')';
  obj.style.webkitTransform = 'scale(' + s + ')';
}

function startSlides()
{
  showSlides = true;
  currentSlide = 0;
  for(var i=0; i<slides.length; i++)
  {
    doScale(slides[i], mainScale );
    slides[i].style.position = 'fixed';
    slides[i].style.top = i * 100 + '%';
    slides[i].style.transition = "top 0.5s";
  }
  document.documentElement.style.overflowY = 'hidden';
}

function flipAspect()
{
    s169 = document.getElementsByClassName("S169");
    if(s169.length!=0) {
        inner = document.getElementsByClassName("Sin169");
        while(inner.length>0) { inner[0].className = "Sin43"; }
        s169 = document.getElementsByClassName("S169");
        while(s169.length>0) { s169[0].className = "S43"; }
    }
    else
    {
        s43 = document.getElementsByClassName("S43");
        inner = document.getElementsByClassName("Sin43");
        while(inner.length>0) { inner[0].className = "Sin169"; }
        while(s43.length>0)   { s43[0].className = "S169"; }
    }  
    calcAspect();
    for (var i = 0; i < slides.length; i++) 
        doScale(slides[i], mainScale);
}

function stopSlides()
{
  showSlides = false;
  currentSlide = undefined;
  for(var i=0; i<slides.length; i++)
  {
    doScale(slides[i], 1 );
    slides[i].style.position = 'relative';
    slides[i].style.transition = '';
    slides[i].style.top = 0;
  }
  document.documentElement.style.overflowY = 'scroll';
}

window.onload   = function() 
{ 
  document.body.style.margin = "0 0 0 0";
  slides = document.getElementsByClassName('S43');
  if( slides.length == 0) 
    slides = document.getElementsByClassName('S169');
  slides = Array.prototype.slice.call(slides);
  calcAspect();

  nav = document.getElementById('navpanel')
  nav.style.visibility = "hidden";
  thumbSize  = .08;
  var gridSize = Math.floor(1 / thumbSize) - 1;
  // Navpanel is 80% of window, allow 20% padding in grid
  thumbScale = thumbSize * mainScale * 0.8 * 0.8;
  for(var i=0; i<slides.length; i++)
  {
    thumbs[i] = slides[i].cloneNode(true);
    thumbs[i].style.position = 'absolute';
    thumbs[i].style.transform = 'scale(' + thumbScale + ')';
    thumbs[i].style.webkitTransform = 'scale(' + thumbScale + ')';
    thumbs[i].style.left    = Math.floor(i/gridSize) * thumbSize*100 + 10 + '%';
    thumbs[i].style.top     = Math.floor(i%gridSize) * thumbSize*100 + 10 + '%';
    thumbs[i].style.zIndex   = 5;
    thumbs[i].onclick = Function('event', 'makeCurrent('+i+')');
    nav.appendChild(thumbs[i]);
  }

  if( window.location.search.includes("style=slideshow") )
    startSlides();
}


function makeCurrent(idx)
{
  if(idx >= slides.length || idx < 0)
    return;

  console.log('Switching to slide ',idx);
  currentSlide = idx;
  for(i=0; i<slides.length; i++)
    slides[i].style.top = (i-currentSlide)*100 + '%';
  return;
}

function flipMode()
{
  if(showSlides) stopSlides();
  else           startSlides();
}

window.onkeydown = function(kcode)
{
  if(!showSlides) return;
  switch(kcode.which)
  {
    case 37: //left
      break;
    case 38: // up
        makeCurrent(currentSlide-1);
      break;
    case 39: //right
      break;
    case 40: // down
        makeCurrent(currentSlide+1);
      break;
  }
}

window.onresize = function()
{
  calcAspect();
  if(!showSlides) return;
  for (var i = 0; i < slides.length; i++) 
    doScale(slides[i], mainScale);
}

function settingsButton()
{
  document.getElementById('navpanel').style.visibility = 'visible';
}

function navcloseButton()
{
  document.getElementById('navpanel').style.visibility = 'hidden';
}

function rightButton()
{
  makeCurrent(currentSlide+1);
}

function leftButton()
{
  makeCurrent(currentSlide-1);
}

function flipPage(num)
{
  var nav = document.getElementById('slide'+num+'_long');
  nav.style.visibility = 'visible';
  nav.style.display     = 'block';
  slides[num].style.visibility = 'hidden';
  slides[num].style.display     = 'none';
}

function flipBackPage(num)
{
  var nav = document.getElementById('slide'+num+'_long');
  nav.style.visibility = 'hidden';
  nav.style.display     = 'none';
  slides[num].style.visibility = 'visible';
  slides[num].style.display     = 'block';
}



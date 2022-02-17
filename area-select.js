

function _getQueryVariable(variable) {
    var query = window.location.search.substring(1);
    var vars = query.split('&');
    for (var i = 0; i < vars.length; i++) {
        var pair = vars[i].split('=');
        if (decodeURIComponent(pair[0]) == variable) {
            return decodeURIComponent(pair[1]);
        }
    }
    return null;
}

$("image").src = _getQueryVariable("tmpfile");
if (window.devicePixelRatio > 1) {
    $("image").addEventListener("load", function(event) {
        var img = $("image");
        img.style.width = (img.width / window.devicePixelRatio) + "px";
    });
}

var _top = $("top");
var _bottom = $("bottom");
var _left = $("left");
var _right = $("right");
var dragging = false;
var finished = false;
var rect = {x: 0, y: 0, width: 0, height: 0 };

document.addEventListener("mousedown", function(event) {
    if (finished) {
        return;
    }
    dragging = true;
    rect.x = event.pageX;
    rect.y = event.pageY;
    _top.style.height = rect.y + "px";
    _left.style.top = rect.y + "px";
    _left.style.left = "0px";
    _left.style.width = rect.x + "px";
    _right.style.top = rect.y + "px";
    return false;
});

document.addEventListener("mousemove", function(event) {
    if (finished || !dragging) {
        return;
    }

    rect.width = event.pageX - rect.x;
    rect.height = event.pageY - rect.y;
    // TODO: handle going from bottom-right to top-left as well.
    _right.style.left = (rect.x + rect.width) + "px";
    _left.style.height = rect.height + "px";
    _right.style.height = rect.height + "px";
    _right.style.right = "0";
    _bottom.style.top = (rect.y + rect.height) + "px";
    _bottom.style.bottom = "0";
});

document.addEventListener("mouseup", function() {
    if (finished) {
        return;
    }
    dragging = false;
    _top.style.display = "none";
    _left.style.display = "none";
    _right.style.display = "none";
    _bottom.style.display = "none";

    rect.x *= window.devicePixelRatio;
    rect.y *= window.devicePixelRatio;
    rect.width *= window.devicePixelRatio;
    rect.height *= window.devicePixelRatio;

    var canvas = document.createElement("canvas");
    canvas.width = rect.width;
    canvas.height = rect.height;
    var ctx = canvas.getContext("2d");
    ctx.drawImage(
        $("image"),
        rect.x, rect.y, rect.width, rect.height,
        0, 0, rect.width, rect.height);
    $("progress").style.display = "block";
    uploadImage(
        CaptureAPI.getBlob(canvas.toDataURL()),
        _getQueryVariable("filename"),
        function(percent) {
            document.querySelector(".meter > span").style.width = percent + "%";
        }, function(url) {
            chrome.tabs.update(null, {url: url});
        });
    finished = true;
});
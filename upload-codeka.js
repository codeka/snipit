

var BASE_URL = "https://codeka.com/";

var loaderHtml = "<div style=\"position: absolute; top: 33%; text-align: center; line-height: 40px;\">" +
    "Uploading..." +
  "</div>";

function uploadImage(blob, filename, progress, complete) {
  var xhr = new XMLHttpRequest();
  xhr.upload.addEventListener("progress", function(event) {
    if (event.lengthComputable) {
      var percentComplete = event.loaded / event.total;
      progress(percentComplete * 0.25 * 100);
    }
  });
  xhr.addEventListener("load", function() {
    var data = JSON.parse(this.responseText);
    complete(BASE_URL + "snip/" + data.slug);
  });
  xhr.open("POST", BASE_URL + "snip/upload?redirect=0");
  var form = new FormData();
  form.append("file", blob, filename);
  xhr.send(form);
}

setInterval(function(){
  var xhttp = new XMLHttpRequest();
  xhttp.onreadystatechange = function() {
    if (this.readyState == 4 && this.status == 200) {
      var data = JSON.parse(xhttp.responseText);
      postMessage(data["Band"]);
    }
  };
  xhttp.open("GET", "/update-band", true);
  xhttp.send();
}, 2000)

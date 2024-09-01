//logger page script

$(document).ready(function() {
	var title = $("title").text()
	if (title == "Log") {
	$("#search-button").click(function() {
    		$.getJSON("/callsearch?call="+$("#search-call").val())
    		.then(function(data){
    			$("#want-update").html(data["QRZMsg"])
    			$("#call").text(data["Call"])
    			$("#qrzname").text(data["Name"])
   			$("#born").text(data["Born"])
    			$("#addr1").text(data["Addr1"])
    			// $("#addr2").text(data["Addr2"])
    			// $("#qrzcountry").text(data["QRZCountry"])
   		 	$("#geolocation").text(data["GeoLoc"])
    			$("#class").text(data["Class"])
   			$("#timezone").text(data["TimeZone"])
    			$("#qsocount").text(data["QSOCount"])
   		});
  	});


	$("#call-sign").blur(function(){
  			$.getJSON("/getconn?call="+$("#call-sign").val())
  			.then(function(data){
    			$("#name").val(data["Name"])
    			$("#country").val(data["Country"])
    			// $("#band-select").val(data["Band"])
    			// $("#mode-select").val(data["Mode"])
    			// alert($("#master-mode").val());
  		});
	});
	};
});


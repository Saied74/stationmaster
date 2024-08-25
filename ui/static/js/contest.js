$(document).ready(function(){  
  	var title = $("title").text()
if (title == "Contest") {
	var err = false;
	$("#call-sign").on("keyup", function(){
		var callSign = $("#call-sign").val()
		var l = callSign.length
		var lastChar = "";
		if (l != 0) {
			lastChar = callSign.charAt(l-1)
		}
		var letterNumber = /\w/;
		var notLetterNumber = /\W/;
		var n = callSign.search(notLetterNumber)
		if (n != -1){
			$("#dupe-call").text("Error")
			err = true
		}
		if (n == -1) {
			$("#dupe-call").text("")
			err = false
		}

		if (l >= 3 && letterNumber.test(lastChar) && (err == false)) {
			$.getJSON("/check-dupe?call="+callSign)
  				.then (function(data){
					if (data["Isdupe"] == "Yes") {
						$("#dupe-call").text("DUPE")
					};
					if (data["Isdupe"] == "No") {
						$("#dupe-call").text("")
					};


			  	});
		};	
	});
	$("#field1").on("focusin", function() {
		if ($("#f1").text().startsWith("RS")) {
			$("#field1").val("599")
		}	
	});

	$("#field2").on("focusin", function() {
		if ($("#f2").text().startsWith("RS")) {
			$("#field2").val("599")
		}		
	});
	$("#field3").on("focusin", function() {
		if ($("#f3").text().startsWith("RS")) {
			$("#field3").val("599")
		}
	});
	$("#field4").on("focusin", function() {
		if ($("#f4").text().startsWith("RS")) {
			$("#field4").val("599")
		}
	});
	$("#field5").on("focusin", function() {
		if ($("#f5").text().startsWith("RS")) {
			$("#field5").val("599")
		}
			
	});


	$("#exChange").on("keyup", function(e) {
		if (e.which === 13) {
			var logdata = {
    				Call:     $("#call-sign").val(),
    				RST: 	  $("#RST").val(),
    				Exchange: $("#exChange").val(),
  			}

  		$.ajax({
            		url: "update-log",
            		type: 'post',
            		dataType: 'json',
            		contentType: 'application/json',
            		data: JSON.stringify(logdata),
        	}).done(function(data){
			$("#message").text(data["Message"])
		});
		$("#call-sign").val("")
		$("#RST").val("")
		$("#exChange").val("");
		$("#dupe-call").text("")
		$("#call-sign").focus()
		};

	});

}
});


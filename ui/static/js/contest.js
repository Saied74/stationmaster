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

	$("#field"+$("#fieldcount").text()).on("keyup", function(e) {

		if (e.which === 13) {
			var s =  $("#seq").text().split(" ")
			var n = parseInt(s[1])
			var logdata = {
    				Call:     $("#call-sign").val(),
				Seq:      s[1],
				Field1:	  $("#field1").val(),
				Field2:	  $("#field2").val(),
				Field3:	  $("#field3").val(),
				Field4:	  $("#field4").val(),
				Field5:    $("#field5").val(),
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
		$("#field1").val("")
		$("#field2").val("")
		$("#field3").val("")
		$("#field4").val("")
		$("#field5").val("")
		$("#dupe-call").text("")
		$("#seq").text("Sequence: " + (n+1))
		$("#call-sign").focus()
		};
	});

      $(document).on("keyup", function(e) {
	      var n = $("#seq").text().split(" ")[1]
	      const functionKeys = [112, 113, 114, 115, 116, 117, 118, 119, 120, 121]
		if (functionKeys.indexOf(e.which) != -1) {
			var keydata = {
    				Call:     $("#call-sign").val(),
				Field1:	  $("#field1").val(),
				Field2:	  $("#field2").val(),
				Field3:	  $("#field3").val(),
				Field4:	  $("#field4").val(),
				Field5:   $("#field5").val(),
				Seq:	  n,
				Key: e.which,
  			}
			$.ajax({
            		url: "update-key",
            		type: 'post',
            		dataType: 'json',
            		contentType: 'application/json',
            		data: JSON.stringify(keydata),
        	}).done(function(data){
			$("#message").text(data["Message"])
		});



		};
      });

    
}
});

	

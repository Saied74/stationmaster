$(document).ready(function(){  
  	var title = $("title").text()
	if (title == "Defaults") {
		
		for (var i = 0; i < 5; i++) {
			var lab1 = "#field" + (i + 1).toString()
			var lab2 = "#contestField" + (i + 1).toString()
			$(lab1).hide()
			$(lab2).hide()
		}

	$("#field-names").on("keyup", function(e) {
		if (e.which === 13) {
			var s = $("#field-names").val()
			var l = s.split(",")
			for (var i = 0; i < l.length; i++) {
				l[i] = l[i].trim()
			}
			var sel = $("#field-count").val()
			var k = parseInt(sel) + 2
			if (k != l.length) {
				$("#match-length").text("number of fields and field labels don't match up, try again")
				return
			} else {
				$("#match-length").text("")
			}
			for (var i = 0; i < k; i++) {
				var lab1 = "#field" + (i + 1).toString()
				$(lab1).text(l[i])
				var lab2 = "#contestField" + (i + 1).toString()
				$(lab1).show()
				$(lab2).show()
			}
		}
	});





	};
});

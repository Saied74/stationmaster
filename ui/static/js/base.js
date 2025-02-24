$(document).ready(function(){ 
	$("#qq").on("click", function() {
		var dummy = {}
		$.ajax({
            		url: "quit",
            		type: 'post',
            		dataType: 'json',
            		contentType: 'application/json',
            		data: JSON.stringify(dummy),
        	}).done(function(){
		});
		

	}
);

});

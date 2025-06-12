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

setInterval(function(){
       $.getJSON("/read-yaesu")
          .then(function(data){ 
      $("#contest-band").html("Band: " + data["Band"])
      $("#contest-mode").html("Mode: " + data["Mode"])
      $("#band-select").val(data["Band"])
      $("#mode-select").val(data["Mode"])
      
      
      });
  }, 2000);



});

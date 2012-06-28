window.setInterval(function() {
	$("#sidebar").load("/current/html");
	$("#status").load("/status");
}, 10000);

$(function() {
	var toggle = function() {
		$('a.toggles i').toggleClass('icon-chevron-left icon-chevron-right');
		$('#sidebar').animate({
			width: 'toggle'
		}, 0);
		$('#content').toggleClass('span12 span9 no-sidebar');
	};
	$('a.toggles').click(toggle);
	toggle();

	$("#sidebar").load("/current/html");
});

// Convert dates to user's timezone on load.
$(function() {
	$('.date').each(function() {
		var dateText = $(this).text() + ' UTC'
		var secs = Date.parse(dateText)
		var d = new Date(secs)
		$(this).text(d.toLocaleString())
	})
})

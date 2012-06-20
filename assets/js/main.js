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
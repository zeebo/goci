var convert = function() {
	var m = 'Jan Feb Mar Apr May Jun Jul Aug Sep Oct Nov Dec'.split(/ /);
	$('.date').each(function() {
		var dateText = $(this).text() + ' UTC'
		var secs = Date.parse(dateText)
		var d = new Date(secs);
		var hours = d.getHours()
		var minutes = d.getMinutes()
		var seconds = d.getSeconds()
		var meridiem = 'AM'
		if (hours > 12) {
			hours -= 12
			meridiem = 'PM'
		}
		if (minutes < 10) minutes = '0'+minutes
		if (seconds < 10) seconds = '0'+seconds
		// Jan 2, 2006 3:04:05 PM
		$(this).text(
			// Date
			m[d.getMonth()]+' '+d.getDate()+', '+d.getFullYear()
			+' '+
			// Time
			hours+':'+minutes+':'+seconds+' '+ meridiem
		)
	})
};

window.setInterval(function() {
	$("#sidebar").load("/current/html");
	$("#status").load("/status");
	convert();
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
	convert();

	$("#sidebar").load("/current/html");
});


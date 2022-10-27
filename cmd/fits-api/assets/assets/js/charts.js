$(document).ready(function($) {
    //start here: initChartParams: function (showMap){
    ldrChartClient.initChartParams(true);
    // Initialise Bootstrap 5 popover functionality
    // (copied from https://getbootstrap.com/docs/5.0/components/popovers/#example-enable-popovers-everywhere)
    var popoverTriggerList = [].slice.call(document.querySelectorAll('[data-bs-toggle="popover"]'));
    var popoverList = popoverTriggerList.map(function (popoverTriggerEl) {
        return new bootstrap.Popover(popoverTriggerEl,
            {delay: {show: "80", hide: "1000"}});
    });
});
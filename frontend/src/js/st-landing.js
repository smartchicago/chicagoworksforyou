var wardPaths = [];
var currWeekEnd = moment().day(-1).startOf('day');
var dateFormat = 'YYYY-MM-DD';
var currServiceType = "4fd3b167e750846744000005"; // Graffiti Removal
var numOfDays = 7;

function redrawChart() {
    var url = 'http://cwfy-api-staging.smartchicagoapps.org/requests/' + currServiceType + '/counts.json?end_date=' + currWeekEnd.format(dateFormat) + '&count=' + numOfDays + '&callback=?';
    $.getJSON(
        url,
        function(response) {
            var counts = _.pairs(response).slice(1,51);
            var sorted = _.sortBy(counts,function(pair) { return pair[1]; });

            var lowest = sorted[0];
            var highest = sorted[49];

            for (var path in wardPaths) {
                var wardNum = parseInt(path, 10) + 1;
                var fillOp = 0.1;
                var col = '#0873AD';
                if (wardNum == lowest[0]) {
                    fillOp = 1;
                } else if (wardNum == highest[0]) {
                    fillOp = 1;
                    col = 'black';
                }
                var poly = L.polygon(
                    wardPaths[path],
                    {
                        color: col,
                        opacity: 1,
                        weight: 2,
                        fillOpacity: fillOp
                    }
                ).addTo(window.map);
                poly.bindPopup('<a href="/wards/' + wardNum + '/">Ward ' + wardNum + '</a>');
            }
        }
    );
}

$(function () {
    for (var ward in WARDS) {
        var points = WARDS[ward].simple_shape[0][0];
        var wardPath = [];
        for (var p in points) {
            var latlong = [points[p][1], points[p][0]];
            wardPath.push(latlong);
        }
        wardPaths.push(wardPath);
    }

    window.map = L.map('map',{scrollWheelZoom: false}).setView([41.83, -87.81], 11);
    L.tileLayer('http://{s}.tile.cloudmade.com/{key}/997/256/{z}/{x}/{y}.png', {
        attribution: 'Map data &copy; <a href="http://openstreetmap.org">OpenStreetMap</a> contributors, <a href="http://creativecommons.org/licenses/by-sa/2.0/">CC-BY-SA</a>, Imagery © <a href="http://cloudmade.com">CloudMade</a>',
        key: '302C8A713FF3456987B21FAAE639A13B',
        maxZoom: 18
    }).addTo(map);
    map.zoomControl.setPosition('bottomright');

    redrawChart();
});

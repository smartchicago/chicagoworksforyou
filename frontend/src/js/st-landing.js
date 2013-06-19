var paths = [];
for (var ward in WARDS) {
    var points = WARDS[ward].simple_shape[0][0];
    var path = [];
    for (var p in points) {
        var latlong = [points[p][1], points[p][0]];
        path.push(latlong);
    }
    paths.push(path);
}

var map = L.map('map',{scrollWheelZoom: false}).setView([41.83, -87.81], 11);
L.tileLayer('http://{s}.tile.cloudmade.com/{key}/997/256/{z}/{x}/{y}.png', {
    attribution: 'Map data &copy; <a href="http://openstreetmap.org">OpenStreetMap</a> contributors, <a href="http://creativecommons.org/licenses/by-sa/2.0/">CC-BY-SA</a>, Imagery © <a href="http://cloudmade.com">CloudMade</a>',
    key: '302C8A713FF3456987B21FAAE639A13B',
    maxZoom: 18
}).addTo(map);
map.zoomControl.setPosition('bottomleft');

for (path in paths) {
    var wardNum = parseInt(path) + 1;
    var fillOp = .1;
    var col = '#0873AD';
    if (wardNum == 1) {
        fillOp = 1;
    } else if (wardNum == 2) {
        fillOp = 1;
        col = 'black';
    }
    var poly = L.polygon(
        paths[path],
        {
            color: col,
            opacity: 1,
            weight: 2,
            fillOpacity: fillOp
        }
    ).addTo(map);
    poly.bindPopup('<a href="/wards/' + wardNum + '/">Ward ' + wardNum + '</a>');
}

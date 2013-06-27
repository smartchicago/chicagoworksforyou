for (var ward in WARDS) {
    var points = WARDS[ward].simple_shape[0][0];
    var path = [];
    for (var p in points) {
        var latlong = [points[p][1], points[p][0]];
        path.push(latlong);
    }
    wardPaths.push(path);
}

var map = L.map('map',{scrollWheelZoom: false}).setView([41.83, -87.81], 11);
L.tileLayer('http://{s}.tile.cloudmade.com/{key}/997/256/{z}/{x}/{y}.png', {
    attribution: 'Map data &copy; <a href="http://openstreetmap.org">OpenStreetMap</a> contributors, <a href="http://creativecommons.org/licenses/by-sa/2.0/">CC-BY-SA</a>, Imagery © <a href="http://cloudmade.com">CloudMade</a>',
    key: '302C8A713FF3456987B21FAAE639A13B',
    maxZoom: 18
}).addTo(map);
map.zoomControl.setPosition('bottomleft');

for (path in wardPaths) {
    var wardNum = parseInt(path,10) + 1;
    var poly = L.polygon(
        wardPaths[path],
        {
            color: '#0873AD',
            opacity: 1,
            weight: 2,
            fillOpacity: (((wardNum % 5) + 2) / 10)
        }
    ).addTo(map);
    poly.bindPopup('<a href="/wards/' + wardNum + '/">Ward ' + wardNum + '</a>');
}

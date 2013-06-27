$(function () {
    buildWardPaths();
    drawChicagoMap();

    for (var path in wardPaths) {
        var wardNum = parseInt(path,10) + 1;
        var poly = L.polygon(
            wardPaths[path],
            {
                color: '#0873AD',
                opacity: 1,
                weight: 2,
                fillOpacity: (((wardNum % 5) + 2) / 10)
            }
        ).addTo(window.map);
        poly.bindPopup('<a href="/wards/' + wardNum + '/">Ward ' + wardNum + '</a>');
    }
});

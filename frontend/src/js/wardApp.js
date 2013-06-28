$(function () {
    var centroid = WARDS[wardNum].centroid;
    var wardCenter = [centroid[1], centroid[0]];
    var points = WARDS[wardNum].simple_shape[0][0];
    var path = [];
    for (var i=0; i<points.length; i++) {
        var latlong = [points[i][1], points[i][0]];
        path.push(latlong);
    }

    // WARD MAP

    var map = L.map('map', {scrollWheelZoom: false}).setView(wardCenter, 13);
    L.tileLayer('http://{s}.tile.cloudmade.com/{key}/997/256/{z}/{x}/{y}.png', {
        attribution: 'Map data &copy; <a href="http://openstreetmap.org">OpenStreetMap</a> contributors, <a href="http://creativecommons.org/licenses/by-sa/2.0/">CC-BY-SA</a>, Imagery © <a href="http://cloudmade.com">CloudMade</a>',
        key: '302C8A713FF3456987B21FAAE639A13B',
        maxZoom: 18
    }).addTo(map);
    map.zoomControl.setPosition('bottomleft');
    var polygon = L.polygon(path).addTo(map);

    // MAKE SUBNAV STICK

    $(".filter").affix({
        offset: { top: 510 }
    });
    $(".subnav-wrap").affix({
        offset: { top: 70 }
    });

    // ALDERMAN NAME

    $.getJSON(
        'http://data.cityofchicago.org/resource/htai-wnw4.json?ward=' + wardNum,
        function(response) {
            var wardInfo = response[0];
            $('.alderman').append('<a href="' + wardInfo.website.url + '"><i class="icon icon-user"></i> ' + wardInfo.alderman + '</a>');
        }
    );

});


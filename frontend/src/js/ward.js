$(function () {
    var ward = wardNum;
    var centroid = WARDS[ward].centroid;
    var wardCenter = [centroid[1], centroid[0]];
    var points = WARDS[ward].simple_shape[0][0];
    var path = [];
    for (var i=0; i<points.length; i++) {
        var latlong = [points[i][1], points[i][0]];
        path.push(latlong);
    }

    var week = 7 * 24 * 60 * 60 * 1000;
    var currServiceType = '';

    // MAKE SUBNAV STICK

    $(".filter").affix({
        offset: { top: 510 }
    });
    $(".subnav-wrap").affix({
        offset: { top: 70 }
    });

    // ALDERMAN NAME

    $.getJSON(
        'http://data.cityofchicago.org/resource/htai-wnw4.json?ward=' + ward,
        function(response) {
            wardInfo = response[0];
            $('.alderman').append('<a href="' + wardInfo.website.url + '"><i class="icon icon-user"></i> ' + wardInfo.alderman + '</a>');
        }
    );

    // WARD MAP

    var map = L.map('map', {scrollWheelZoom: false}).setView(wardCenter, 13);
    L.tileLayer('http://{s}.tile.cloudmade.com/{key}/997/256/{z}/{x}/{y}.png', {
        attribution: 'Map data &copy; <a href="http://openstreetmap.org">OpenStreetMap</a> contributors, <a href="http://creativecommons.org/licenses/by-sa/2.0/">CC-BY-SA</a>, Imagery © <a href="http://cloudmade.com">CloudMade</a>',
        key: '302C8A713FF3456987B21FAAE639A13B',
        maxZoom: 18
    }).addTo(map);
    map.zoomControl.setPosition('bottomleft');
    var polygon = L.polygon(path).addTo(map);

    // CHARTS

    function redrawChart(response) {
        var categories = [];
        var counts = [];
        for (var d in response) {
            categories.push(moment(d).format("MMM DD"));
            counts.push(response[d]);
        }
        countsChart.series[0].setData(counts);
        countsChart.xAxis[0].setCategories(categories);
        var currWeek = weekDuration.beforeMoment(currWeekEnd,true);
        $('.this-week a').text(currWeek.format({implicitYear: false}));
    }

    // CHARTS WEEKNAV

    $('.this-week a').click(function(evt) {
        evt.preventDefault();
    });

    $('.next-week a').click(function(evt) {
        evt.preventDefault();
        currWeekEnd.add('week',1);
        $.getJSON(
            window.apiDomain + 'wards/' + ward + '/counts.json?count=7&service_code=' + currServiceType + '&end_date=' + currWeekEnd.format(dateFormat) + '&callback=?',
            function(response) {redrawChart(response);}
        );
    });

    $('.prev-week a').click(function(evt) {
        evt.preventDefault();
        currWeekEnd.subtract('week',1);
        $.getJSON(
            window.apiDomain + 'wards/' + ward + '/counts.json?count=7&service_code=' + currServiceType + '&end_date=' + currWeekEnd.format(dateFormat) + '&callback=?',
            function(response) {redrawChart(response);}
        );
    });

    // SERVICE TYPE LIST

    $('.st-list').on('click', 'a', function(evt) {
        evt.preventDefault();
        currServiceType = $(this).attr('data-code');

        $.getJSON(
            window.apiDomain + 'wards/' + ward + '/counts.json?count=7&service_code=' + currServiceType + '&end_date=' + currWeekEnd.format(dateFormat) + '&callback=?',
            function(response) {redrawChart(response);}
        );

        $(this).closest('li').addClass('active').siblings().removeClass('active');
    });

    $('.st-list .initial a').trigger('click');

    Highcharts.setOptions({
        chart: {
            marginBottom: 80,
            type: 'column'
        },
        title: {
            text: ''
        },
        xAxis: {
            minPadding: 0.05,
            maxPadding: 0.05,
            tickmarkPlacement: 'between',
            labels: {
                style: {
                    fontFamily: 'Roboto, sans-serif',
                    fontSize: '13px'
                },
                y: 22
            }
        },
        yAxis: {
            title: {
                text: ''
            },
            minPadding: 0.1,
            labels: {
                style: {
                    fontFamily: 'Roboto, sans-serif',
                    fontWeight: 'bold'
                },
                align: 'left',
                x: 0,
                y: -2
            }
        },
        plotOptions: {
            column: {
                groupPadding: 0.1
            }
        },
        tooltip: {
            headerFormat: '',
            pointFormat: '<b>{point.y:,.0f}</b> tickets',
            shadow: false,
            style: {
                fontFamily: 'Roboto, sans-serif',
                fontSize: '15px'
            }
        },
        legend: {
            enabled: true,
            borderWidth: 0,
            backgroundColor: "#f7f7f7",
            padding: 10
        }
    });

    var countsChart = new Highcharts.Chart({
        chart: {
            renderTo: 'counts-chart'
        },
        series: [{
            name: "Ward " + ward
        },{
            name: "City average",
            data: [5, 6, 7, 8, 4, 3, 9],
            type: 'line',
            dashStyle: 'longdash'
        }]
    });
});

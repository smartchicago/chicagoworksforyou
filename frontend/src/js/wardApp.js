// JQUERY

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
});

// ANGULAR

var wardApp = angular.module('wardApp', []);

wardApp.config(function($routeProvider) {
    $routeProvider.
        when('/:serviceSlug', {
            controller: "wardCtrl",
            templateUrl: "/views/ward_charts.html"
        }).
        when('/:serviceSlug/:date', {
            controller: "wardCtrl",
            templateUrl: "/views/ward_charts.html"
        }).
        otherwise({
            redirectTo: '/graffiti_removal'
        });
});

// WARD DETAIL

wardApp.controller("serviceListCtrl", function ($scope, $http, $location) {
    $http.get('/data/services.json').success(function(data) {
        $scope.services = data;
    });
    $scope.orderProp = 'name';

    $scope.isActive = function(slug) {
        var currServiceSlug = $location.path().substr(1);
        return slug == currServiceSlug;
    };
});

wardApp.controller("wardCtrl", function ($scope, $location, $routeParams) {
    var serviceType = window.lookupSlug($routeParams.serviceSlug);
    var serviceCode = serviceType.code;

    // CHARTS WEEKNAV

    $('.this-week a').click(function(evt) {
        evt.preventDefault();
    });

    $('.next-week a').click(function(evt) {
        evt.preventDefault();
        currWeekEnd.add('week',1);
        $.getJSON(
            window.apiDomain + 'wards/' + wardNum + '/counts.json?count=7&service_code=' + serviceCode + '&end_date=' + currWeekEnd.format(dateFormat) + '&callback=?',
            function(response) {redrawChart(response);}
        );
    });

    $('.prev-week a').click(function(evt) {
        evt.preventDefault();
        currWeekEnd.subtract('week',1);
        $.getJSON(
            window.apiDomain + 'wards/' + wardNum + '/counts.json?count=7&service_code=' + serviceCode + '&end_date=' + currWeekEnd.format(dateFormat) + '&callback=?',
            function(response) {redrawChart(response);}
        );
    });

    // CHART

    $.getJSON(
        window.apiDomain + 'wards/' + wardNum + '/counts.json?count=7&service_code=' + serviceCode + '&end_date=' + currWeekEnd.format(dateFormat) + '&callback=?',
        function(response) {redrawChart(response);}
    );

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

    var countsChart = new Highcharts.Chart({
        chart: {
            renderTo: 'counts-chart'
        },
        series: [{
            name: "Ward " + wardNum
        },{
            name: "City average",
            data: [5, 6, 7, 8, 4, 3, 9],
            type: 'line',
            dashStyle: 'longdash'
        }]
    });

    $scope.wardNum = wardNum;
});

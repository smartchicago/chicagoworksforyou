// JQUERY

$(function () {
    var centroids = [null,[-87.683231,41.914304],[-87.648803,41.867456],[-87.62964,41.813566],[-87.603125,41.816155],[-87.586705,41.776022],[-87.618591,41.748371],[-87.560683,41.738814],[-87.588308,41.735931],[-87.610453,41.678697],[-87.55646,41.687401],[-87.653323,41.829887],[-87.69005,41.827565],[-87.733226,41.769769],[-87.711642,41.807895],[-87.684336,41.779294],[-87.668058,41.791246],[-87.654714,41.761173],[-87.697921,41.749512],[-87.68919,41.700988],[-87.62324,41.78518],[-87.645912,41.73037],[-87.723363,41.838713],[-87.762182,41.792351],[-87.723381,41.86401],[-87.662184,41.855632],[-87.703834,41.906883],[-87.676908,41.892697],[-87.724581,41.879511],[-87.770902,41.897949],[-87.740537,41.930753],[-87.743533,41.928236],[-87.669543,41.92199],[-87.705979,41.955858],[-87.643502,41.687812],[-87.709059,41.931693],[-87.815517,41.939613],[-87.751139,41.905722],[-87.769601,41.952115],[-87.729676,41.978135],[-87.686528,41.985198],[-87.864869,41.984042],[-87.625883,41.890473],[-87.641061,41.919772],[-87.652198,41.942085],[-87.763801,41.975583],[-87.65149,41.960896],[-87.68137,41.959399],[-87.658782,41.982538],[-87.671545,42.011679],[-87.697288,42.002965]];
    var wardCentroid = centroids[wardNum];
    var wardCenter = [wardCentroid[1], wardCentroid[0]];

    // WARD MAP

    var map = L.map('map', {scrollWheelZoom: false}).setView(wardCenter, 13);
    L.tileLayer('http://{s}.tile.cloudmade.com/{key}/997/256/{z}/{x}/{y}.png', {
        attribution: 'Map data &copy; <a href="http://openstreetmap.org">OpenStreetMap</a> contributors, <a href="http://creativecommons.org/licenses/by-sa/2.0/">CC-BY-SA</a>, Imagery © <a href="http://cloudmade.com">CloudMade</a>',
        key: '302C8A713FF3456987B21FAAE639A13B',
        maxZoom: 18
    }).addTo(map);
    map.zoomControl.setPosition('bottomleft');
    var polygon = L.polygon(wardPaths[wardNum - 1]).addTo(map);

    // MAKE SUBNAV STICK

    $(".filter").affix({
        offset: { top: 530 }
    });
    $(".subnav-wrap").affix({
        offset: { top: 70 }
    });

    // EVENT CONTROL

    $('.this-week a').click(function(evt) {
        evt.preventDefault();
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

wardApp.factory('Data', function () {
    return {};
});

wardApp.controller("sidebarCtrl", function ($scope, Data, $http, $location) {
    $http.get('/data/services.json').success(function(response) {
        Data.services = response;
    });

    $scope.data = Data;

    $scope.prevWeek = function () {
        $location.path(Data.currServiceSlug + "/" + Data.prevWeek);
    };

    $scope.nextWeek = function () {
        $location.path(Data.currServiceSlug + "/" + Data.nextWeek);
    };
});

wardApp.controller("wardCtrl", function ($scope, Data, $http, $routeParams) {
    var date = moment().subtract('days', 1).startOf('day'); // Last Saturday
    if ($routeParams.date) {
        date = moment($routeParams.date);
    }

    var serviceObj = window.lookupSlug($routeParams.serviceSlug);

    Data.wardNum = window.wardNum;
    Data.currServiceSlug = $routeParams.serviceSlug;
    Data.currServiceName = serviceObj.name;
    Data.dateFormatted = date.format(dateFormat);
    Data.prevWeek = moment(date).subtract('week',1).format(dateFormat);
    Data.nextWeek = moment(date).add('week',1).format(dateFormat);
    Data.thisWeek = weekDuration.beforeMoment(date,true).format({implicitYear: false});

    $scope.data = Data;

    var serviceCode = serviceObj.code;
    var ticketsURL = window.apiDomain + 'wards/' + window.wardNum + '/counts.json?count=7&service_code=' + serviceCode + '&end_date=' + Data.dateFormatted + '&callback=JSON_CALLBACK';
    var ttcURL = window.apiDomain + 'requests/time_to_close.json?count=7&service_code=' + serviceCode + '&end_date=' + Data.dateFormatted + '&callback=JSON_CALLBACK';

    // CHARTS

    $http.jsonp(ticketsURL).
        success(function(response, status, headers, config) {
            var categories = [];
            var counts = [];
            var cityAverages = [];

            for (var d in response) {
                categories.push(moment(d).format("MMM DD"));
                counts.push(response[d].Count);
                cityAverages.push(response[d].CityAverage);
            }

            var countsChart = new Highcharts.Chart({
                chart: {
                    renderTo: 'counts-chart'
                },
                xAxis: {
                    categories: categories
                },
                series: [{
                    name: "Ward " + wardNum,
                    data: counts
                },{
                    name: "City average",
                    data: cityAverages,
                    type: 'line',
                    dashStyle: 'longdash'
                }]
            });
        }
    );

    $http.jsonp(ttcURL).
        success(function(response, status, headers, config) {
            var values = _.rest(_.values(response));
            var sorted = _.sortBy(values, function (obj) { return obj.Time; });
            var wards = _.pluck(sorted, 'Ward');
            var position = _.indexOf(wards, Data.wardNum);
            Data.wardRank = window.getOrdinal(position + 1);
            Data.wardTime = Math.round(sorted[position].Time * 10) / 10;
            Data.maxTTC = _.last(sorted);
            Data.ttc = sorted;
        }
    );
});

// HIGHCHARTS

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
        formatter: function() {
            return '<b>' + this.y + '</b> ' + ' ticket' + (this.y > 1 ? 's' : '');
        },
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

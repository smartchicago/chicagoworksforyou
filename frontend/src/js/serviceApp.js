// JQUERY

$(function () {
    $('.pagination-wrap').affix({
        offset: {top: $('.pagination').position().top}
    });
});

// ANGULAR

var serviceApp = angular.module('serviceApp', []).value('$anchorScroll', angular.noop);

serviceApp.config(function($routeProvider) {
    $routeProvider.
        when('/', {
            action: "overview"
        }).
        when('/:date/', {
            action: "detail"
        }).
        otherwise({
            redirectTo: '/'
        });
});

serviceApp.factory('Data', function () {
    var data = {};

    data.setDate = function(date) {
        data.date = date.format(dateFormat);
        data.dateFormatted = date.format('MMM D, YYYY');

        data.startDate = date.clone().day(0);
        data.endDate = date.clone().day(6).max(window.yesterday);
        data.duration = data.endDate.diff(data.startDate, 'days');
        data.thisDate = moment.duration(data.duration,"days").beforeMoment(data.endDate,true).format({implicitYear: false});

        data.prevDate = data.startDate.clone().subtract('day',1);
        data.nextDate = data.endDate.clone().add('day',7);
        data.isLatest = data.nextDate.clone().day(0).isAfter(window.yesterday);
    };

    return data;
});

serviceApp.controller("sidebarCtrl", function ($scope, Data, $http, $location) {
    $scope.data = Data;

    $scope.goToPrevDate = function() {
        if (Data.prevDate.clone().day(0).isBefore(window.earliestDate)) {
            return false;
        }
        $location.path(Data.prevDate.format(dateFormat) + "/");
    };

    $scope.goToNextDate = function() {
        if (Data.isLatest) {
            return false;
        }
        $location.path(Data.nextDate.format(dateFormat) + "/");
    };
});

serviceApp.controller("serviceCtrl", function ($scope, Data, $http, $location, $route, $routeParams) {
    var stCode = window.currServiceType;
    var stSlug = window.lookupCode(stCode).slug;
    var chart = $('#chart').highcharts();

    var renderChart = function() {
        var url = window.apiDomain + 'requests/' + stCode + '/counts.json?end_date=' + Data.endDate.format(dateFormat) + '&count=' + (Data.duration + 1) + '&callback=JSON_CALLBACK';

        $http.jsonp(url).
            success(function(response, status, headers, config) {
                Data.cityCount = response.CityData.Count;
                Data.cityAverage = response.CityData.Count / 50;
                var wardData = response.WardData;
                var categories = _.map(_.keys(wardData), function (wardNum) { return '<a href="/ward/' + wardNum + '/#/' + Data.endDate.format(dateFormat) + '/' + stSlug + '">Ward ' + wardNum + '</a>'; });
                var days = [[],[],[],[],[],[],[]];
                for (var ward in wardData) {
                    var i = 0;
                    for (var count in wardData[ward].Counts) {
                        days[i++].push(wardData[ward].Counts[count]);
                    }
                }

                var weekdays = ["Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"];
                var series = [];
                for (var day in days) {
                    if (days[day].length > 0) {
                        series.push({
                            name: weekdays[day],
                            data: days[day],
                            stack: 0,
                            legendIndex: day + 1
                        });
                    }
                }

                new Highcharts.Chart({
                    chart: {
                        renderTo: 'chart'
                    },
                    colors: [
                        '#37c0b9',
                        '#37acc3',
                        '#3790c7',
                        '#3973c9',
                        '#3a56ca',
                        '#403ccc',
                        '#603fce'
                    ].reverse(),
                    series: series.reverse(),
                    xAxis: {
                        categories: categories
                    },
                    yAxis: {
                        opposite: true,
                        plotLines: [{
                            id: 'avg',
                            value: Data.cityAverage,
                            color: 'black',
                            width: 3,
                            zIndex: 5
                        }]
                    }
                });
            }
        );
    };

    $scope.data = Data;

    $scope.$on(
        "$routeChangeSuccess",
        function ($e, $currentRoute, $previousRoute) {
            Data.setDate(parseDate($routeParams.date, window.yesterday, $location));
            Data.currURL = "#/" + Data.date + "/";
            renderChart();
        }
    );
});

// HIGHCHARTS

Highcharts.setOptions({
    chart: {
        marginBottom: 40,
        type: 'bar'
    },
    title: {
        text: ''
    },
    xAxis: {
        tickmarkPlacement: 'between',
        labels: {
            style: {
                fontFamily: 'Monda, Helvetica, sans-serif',
                fontSize: '13px'
            },
            y: 5
        }
    },
    yAxis: {
        title: {
            text: ''
        },
        minPadding: 0.1,
        labels: {
            style: {
                fontFamily: 'Monda, Helvetica, sans-serif',
                fontWeight: 'bold'
            }
        }
    },
    plotOptions: {
        bar: {
            animation: false,
            borderWidth: 0,
            groupPadding: 0.08,
            dataLabels: {
                enabled: false,
                color: "#000000",
                style: {
                    fontFamily: "Monda, Helvetica, sans-serif",
                    fontSize: '13px',
                    fontWeight: 'bold'
                }
            },
            pointPadding: 0,
            stacking: 'normal'
        }
    },
    tooltip: {
        headerFormat: '',
        // pointFormat: '<b>{point.y:,.0f}</b> requests',
        shadow: false,
        style: {
            fontFamily: 'Monda, Helvetica, sans-serif',
            fontSize: '15px'
        },
        formatter: function() {
            return this.series.name + ': <b>' + this.y + '</b>';
        }
    },
    legend: {
        enabled: true,
        borderWidth: 0,
        backgroundColor: "#f7f7f7",
        padding: 10,
        verticalAlign: 'top',
        y: 10
    }
});

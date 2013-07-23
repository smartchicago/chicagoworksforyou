// JQUERY

$(function () {
    $('.pagination-wrap').affix({
        offset: {top: $('.pagination').position().top}
    });
});

// ANGULAR

var serviceApp = angular.module('serviceApp', []).value('$anchorScroll', angular.noop);

serviceApp.factory('Data', function () {
    return {};
});

serviceApp.config(function($routeProvider) {
    $routeProvider.
        when('/', {
            controller: "serviceCtrl",
            templateUrl: "/views/service_chart.html"
        }).
        when('/:date', {
            controller: "serviceCtrl",
            templateUrl: "/views/service_chart.html"
        }).
        otherwise({
            redirectTo: '/'
        });
});

serviceApp.controller("sidebarCtrl", function ($scope, Data, $http, $location) {
    $scope.data = Data;

    $scope.prevWeek = function () {
        $location.path(Data.prevWeek);
    };

    $scope.nextWeek = function () {
        $location.path(Data.nextWeek);
    };
});

serviceApp.controller("serviceCtrl", function ($scope, Data, $http, $location, $routeParams) {
    var date = parseDate($routeParams.date, window.prevSaturday, $location);

    Data.dateFormatted = date.format(dateFormat);
    Data.prevWeek = moment(date).subtract('week',1).format(dateFormat);
    Data.nextWeek = moment(date).add('week',1).format(dateFormat);
    Data.thisWeek = weekDuration.beforeMoment(date,true).format({implicitYear: false});

    $scope.data = Data;

    var stCode = window.currServiceType;
    var stSlug = window.lookupCode(stCode).slug;
    var numOfDays = 28;
    var url = window.apiDomain + 'requests/' + stCode + '/counts.json?end_date=' + Data.dateFormatted + '&count=' + numOfDays + '&callback=JSON_CALLBACK';
    var chart = $('#chart').highcharts();

    $http.jsonp(url).
        success(function(response, status, headers, config) {
            var cityAverage = response['0'].Count / 50;
            var counts = _.rest(_.pluck(response, 'Count'));
            var categories = _.map(_.rest(_.keys(response)), function (wardNum) { return '<a href="/ward/' + wardNum + '/#/' + stSlug + '">Ward ' + wardNum + '</a>'; });
            var averages = _.map(_.rest(_.pluck(response, 'Average')), function (avg) { return Math.round(avg * numOfDays); });

            new Highcharts.Chart({
                chart: {
                    renderTo: 'chart'
                },
                series: [{
                    data: counts,
                    id: 'counts',
                    index: 2,
                    dataLabels: {
                        style: {
                            fontWeight: 'bold'
                        }
                    }
                },{
                    data: averages,
                    index: 1
                }],
                xAxis: {
                    categories: categories
                },
                yAxis: {
                    plotLines: [{
                        id: 'avg',
                        value: cityAverage,
                        color: 'brown',
                        width: 3,
                        zIndex:5
                    }]
                }
            });
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
                fontFamily: 'Roboto, sans-serif',
                fontSize: '13px'
            }
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
            }
        }
    },
    plotOptions: {
        bar: {
            borderWidth: 0,
            groupPadding: 0.08,
            dataLabels: {
                enabled: true,
                color: "#000000",
                style: {
                    fontFamily: "Roboto, sans-serif",
                    fontSize: '13px'
                }
            },
            pointPadding: 0
        }
    },
    tooltip: {
        headerFormat: '',
        pointFormat: '<b>{point.y:,.0f}</b> requests',
        shadow: false,
        style: {
            fontFamily: 'Roboto, sans-serif',
            fontSize: '15px'
        }
    },
    legend: {
        enabled: false
    }
});

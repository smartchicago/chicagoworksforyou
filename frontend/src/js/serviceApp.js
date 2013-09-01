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
    var serviceObj = window.lookupCode(window.currServiceType);
    var data = {
        stSlug: serviceObj.slug,
        stName: serviceObj.name,
        dayColors: [
            '#1A3E50',
            '#22546C',
            '#2B6A88',
            '#3380A4',
            '#3A95C1',
            '#54A5CC',
            '#6FB4D5'
        ]
    };

    data.setDate = function(date) {
        data.date = date.format(dateFormat);
        data.dateFormatted = date.format('MMM D, YYYY');

        data.startDate = date.clone().day(0);
        data.endDate = date.clone().day(6).max(window.yesterday);
        data.duration = data.endDate.diff(data.startDate, 'days');
        data.thisDate = moment.duration(data.duration,"days").beforeMoment(data.endDate,true).format({implicitYear: false});
        data.pageTitle = data.thisDate + ' | ' + data.stName + ' | Chicago Works For You';

        var today = moment().startOf('day');

        data.days = _.map(data.dayColors, function(color, i) {
            var day = data.startDate.clone().startOf('day').add('day',i);
            var inFuture = !day.isBefore(today);
            return {
                'i': i,
                'date': day.format(),
                'inFuture': inFuture,
                'color': inFuture ? "#dddddd" : color
            };
        });

        data.prevDate = data.startDate.clone().subtract('day',1);
        data.nextDate = data.endDate.clone().add('day',7);
        data.isLatest = data.nextDate.clone().day(0).isAfter(window.yesterday);
    };

    return data;
});

serviceApp.controller("headCtrl", function ($scope, Data) {
    $scope.data = Data;
});

serviceApp.controller("headerCtrl", function ($scope, Data, $location) {
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

serviceApp.controller("sidebarCtrl", function ($scope, Data) {
    $scope.data = Data;
});

serviceApp.controller("serviceCtrl", function ($scope, Data, $http, $location, $route, $routeParams) {
    var stCode = window.currServiceType;
    var chart = $('#chart').highcharts();

    var renderChart = function (categories, requests, closes) {
        var series = _.clone(requests);
        if (closes) {
            series.push(closes);
        }
        var chart = new Highcharts.Chart({
            chart: {
                renderTo: 'chart'
            },
            colors: _.clone(Data.dayColors).slice(0, requests.length).reverse(),
            series: series.reverse(),
            xAxis: {
                categories: categories
            },
            yAxis: {
                opposite: true
            }
        });
    };

    var buildChart = function() {
        var endDate = Data.endDate.format(dateFormat);
        var count = Data.duration + 1;

        var requestsURL = window.apiDomain + 'requests/' + stCode + '/counts.json?end_date=' + endDate + '&count=' + count + '&callback=JSON_CALLBACK';
        var closesURL = window.apiDomain + 'requests/time_to_close.json?service_code=' + stCode + '&end_date=' + endDate + '&count=' + count + '&callback=JSON_CALLBACK';

        $http.jsonp(requestsURL).
            success(function(response, status, headers, config) {
                Data.cityCount = response.CityData.Count;

                var wardData = _.sortBy(response.WardData, function(ward, wardNum) {
                    ward.Ward = wardNum;
                    return parseInt(wardNum, 10);
                });

                var categories = _.map(wardData, function (ward) {
                    return '<a href="/ward/' + ward.Ward + '/#/' + Data.endDate.format(dateFormat) + '/' + Data.stSlug + '">Ward ' + ward.Ward + '</a>';
                });

                var days = [[],[],[],[],[],[],[]];
                _.each(wardData, function(ward) {
                    var i = 0;
                    for (var count in ward.Counts) {
                        days[i++].push(ward.Counts[count]);
                    }
                });

                var weekdays = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"];
                var requestSeries = [];
                for (var day in days) {
                    if (days[day].length > 0) {
                        requestSeries.push({
                            name: weekdays[day],
                            data: days[day],
                            stack: 0,
                            legendIndex: day + 1
                        });
                    }
                }

                $http.jsonp(closesURL).
                    success(function(response, status, headers, config) {
                        Data.cityCloseCount = response.CityData.Count;
                        var nonZeroWards = _.reject(response.WardData, function(ward, key) { return key === '0'; });
                        var wardCloseData = _.sortBy(nonZeroWards, function(ward, wardNum) {
                            ward.Ward = wardNum;
                            return parseInt(wardNum, 10);
                        });
                        var closeCounts = _.map(_.pluck(wardCloseData, 'Count'), function(val) { return val || null; });
                        var closeSeries = {
                            data: closeCounts,
                            type: 'scatter',
                            name: 'Requests closed',
                            color: 'black',
                            stack: 0,
                            index: 10,
                            legendIndex: 100,
                            marker: {
                                symbol: 'url(/img/check-rotated.png)'
                            }
                        };
                        renderChart(categories, requestSeries, closeSeries);
                    }).
                    error(function(data, status, headers, config) {
                        renderChart(categories, requestSeries);
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
            buildChart();
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
                fontSize: '13px',
            },
            useHTML: true,
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
                fontWeight: 'bold',
                color: '#222',
                fontSize: '12px'
            }
        },
        stackLabels: {
            enabled: true,
            color: "#000000",
            // format: '◯ {total}',
            format: '{total}',
            style: {
                fontFamily: "Monda, Helvetica, sans-serif",
                fontSize: '13px'
            }
        }
    },
    plotOptions: {
        bar: {
            animation: false,
            borderWidth: 0,
            groupPadding: 0.08,
            dataLabels: {
                enabled: false
            },
            pointPadding: 0,
            stacking: 'normal',
            tooltip: {
                headerFormat: '',
                pointFormat: '{series.name}: <b>{point.y} opened</b>'
            }
        },
        scatter: {
            animation: false,
            tooltip: {
                headerFormat: '',
                pointFormat: '<b>{point.y} closed</b> this week'
            }
        }
    },
    tooltip: {
        headerFormat: '',
        shadow: false,
        style: {
            fontFamily: 'Monda, Helvetica, sans-serif',
            fontSize: '15px'
        }
    },
    legend: {
        enabled: false,
        borderWidth: 0,
        backgroundColor: "#f7f7f7",
        padding: 10,
        verticalAlign: 'top',
        y: 10
    }
});

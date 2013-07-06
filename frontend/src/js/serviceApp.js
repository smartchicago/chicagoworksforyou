function redrawChart() {
    var stCode = currServiceType;
    var numOfDays = 7;
    var url = window.apiDomain + 'requests/' + stCode + '/counts.json?end_date=' + currWeekEnd.format(dateFormat) + '&count=' + numOfDays + '&callback=?';
    var chart = $('#chart').highcharts();

    $.getJSON(
        url,
        function(response) {
            var cityAverage = response['0'].Count / 50;
            var counts = _.rest(_.pluck(response, 'Count'));

            // var sorted = _.sortBy(counts,function(pair) { return pair[1]; }).reverse();
            // var sortedCategories = _.map(sorted, function(pair) { return "Ward " + pair[0]; });
            // var sortedFakeWardAverages = _.map(sorted, function(pair) { return Math.max(pair[1] - Math.ceil((Math.random() - 0.5) * 20),0); });

            if (chart) {
                chart.get('counts').setData(counts);
                chart.yAxis[0].removePlotLine('avg');
                chart.yAxis[0].addPlotLine({
                    id: 'avg',
                    value: cityAverage,
                    color: 'brown',
                    width: 3,
                    zIndex:5
                });
            } else {
                var categories = _.map(_.rest(_.pluck(response, 'Ward')), function (wardNum) { return "Ward " + wardNum; });
                var averages = _.map(_.rest(_.pluck(response, 'Average')), Math.round);

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

            var currWeek = weekDuration.beforeMoment(currWeekEnd,true);
            $('.this-week a').text(currWeek.format({implicitYear: false}));
        }
    );
}

$(function () {
    // WEEKNAV

    $('.this-week a').click(function(evt) {
        evt.preventDefault();
    });

    $('.next-week a').click(function(evt) {
        evt.preventDefault();
        currWeekEnd.add('week',1);
        redrawChart();
    });

    $('.prev-week a').click(function(evt) {
        evt.preventDefault();
        currWeekEnd.subtract('week',1);
        redrawChart();
    });

    // CHART SETTINGS

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
            pointFormat: '<b>{point.y:,.0f}</b> tickets',
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

    // SET UP INITIAL CHART

    redrawChart();
});

// ANGULAR

var serviceApp = angular.module('serviceApp', []);

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

serviceApp.controller("serviceListCtrl", function ($scope, $http, $location) {
    $http.get('/data/services.json').success(function(data) {
        $scope.services = data;
    });
    $scope.orderProp = 'name';
});

serviceApp.controller("serviceCtrl", function ($scope, $http, $routeParams) {
});

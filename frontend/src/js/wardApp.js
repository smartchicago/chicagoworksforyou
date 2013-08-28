var centroids = [null,[-87.683231,41.914304],[-87.648803,41.867456],[-87.62964,41.813566],[-87.603125,41.816155],[-87.586705,41.776022],[-87.618591,41.748371],[-87.560683,41.738814],[-87.588308,41.735931],[-87.610453,41.678697],[-87.55646,41.687401],[-87.653323,41.829887],[-87.69005,41.827565],[-87.733226,41.769769],[-87.711642,41.807895],[-87.684336,41.779294],[-87.668058,41.791246],[-87.654714,41.761173],[-87.697921,41.749512],[-87.68919,41.700988],[-87.62324,41.78518],[-87.645912,41.73037],[-87.723363,41.838713],[-87.762182,41.792351],[-87.723381,41.86401],[-87.662184,41.855632],[-87.703834,41.906883],[-87.676908,41.892697],[-87.724581,41.879511],[-87.770902,41.897949],[-87.740537,41.930753],[-87.743533,41.928236],[-87.669543,41.92199],[-87.705979,41.955858],[-87.643502,41.687812],[-87.709059,41.931693],[-87.815517,41.939613],[-87.751139,41.905722],[-87.769601,41.952115],[-87.729676,41.978135],[-87.686528,41.985198],[-87.864869,41.984042],[-87.625883,41.890473],[-87.641061,41.919772],[-87.652198,41.942085],[-87.763801,41.975583],[-87.65149,41.960896],[-87.68137,41.959399],[-87.658782,41.982538],[-87.671545,42.011679],[-87.697288,42.002965]];
var wardCentroid = centroids[wardNum];
var wardCenter = [wardCentroid[1], wardCentroid[0]];

// JQUERY

$(function () {
    // MAKE FILTER STICK

    $(".filter").affix({
        offset: { top: 530 }
    });
});

// ANGULAR

var wardApp = angular.module('wardApp', []).value('$anchorScroll', angular.noop);

wardApp.config(function($routeProvider) {
    $routeProvider.
        when('/', {
            action: "overview"
        }).
        when('/:date/', {
            action: "overview"
        }).
        when('/:date/:serviceSlug/', {
            action: "detail"
        }).
        otherwise({
            redirectTo: '/'
        });
});

wardApp.factory('Data', function () {
    var data = {
        wardNum: window.wardNum
    };

    if (!window.chicagoMap) {
        window.chicagoMap = L.map('map', {scrollWheelZoom: false}).setView(wardCenter, 13);
        L.tileLayer('http://{s}.tile.cloudmade.com/{key}/{styleId}/256/{z}/{x}/{y}.png', window.mapOptions)
            .addTo(window.chicagoMap);
        window.chicagoMap.zoomControl.setPosition('bottomleft');
        var polygon = L.polygon(wardPaths[wardNum - 1],
            {
                opacity: 1,
                weight: 2,
                dashArray: '3',
                color: '#182A35',
                fillOpacity: 0.7,
                fillColor: '#4888AF'
            }
        ).addTo(window.chicagoMap);
    }

    data.setDate = function(date) {
        data.date = date.format(dateFormat);
        data.dateObj = date;
        data.dateFormatted = date.format('MMM D, YYYY');

        data.startDate = date.clone().day(0);
        data.endDate = date.clone().day(6).max(window.yesterday);
        data.duration = data.endDate.diff(data.startDate, 'days');
        data.thisDate = moment.duration(data.duration,"days").beforeMoment(data.endDate,true).format({implicitYear: false});
        data.pageTitle = data.thisDate + ' | Ward ' + window.wardNum + ' | Chicago Works For You';

        data.prevDate = data.startDate.clone().subtract('day',1);
        data.nextDate = data.endDate.clone().add('day',7);
        data.isLatest = data.nextDate.clone().day(0).isAfter(window.yesterday);
    };

    return data;
});

wardApp.controller("headCtrl", function ($scope, Data) {
    $scope.data = Data;
});

wardApp.controller("headerCtrl", function ($scope, Data, $location) {
    $scope.data = Data;

    var urlSuffix = function() {
        return Data.serviceObj.slug ? Data.serviceObj.slug + '/' : '';
    };

    $scope.goToPrevDate = function() {
        if (Data.prevDate.clone().day(0).isBefore(window.earliestDate)) {
            return false;
        }
        $location.path(Data.prevDate.format(dateFormat) + "/" + urlSuffix());
    };

    $scope.goToNextDate = function() {
        if (Data.isLatest) {
            return false;
        }
        $location.path(Data.nextDate.format(dateFormat) + "/" + urlSuffix());
    };
});

wardApp.controller("sidebarCtrl", function ($scope, Data, $http, $location) {
    $http.get('/data/services.json').success(function(response) {
        Data.services = response;
    });

    $scope.data = Data;
});

wardApp.controller("wardChartCtrl", function ($scope, Data, $http, $location, $route, $routeParams) {
    var highchartsDefaults = {
        chart: {
            marginBottom: 30
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
                    fontFamily: 'Monda, sans-serif',
                    fontSize: '13px'
                },
                useHTML: true,
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
                    fontFamily: 'Monda, sans-serif',
                    fontWeight: 'bold'
                },
                align: 'left',
                x: 0,
                y: 3
            },
            offset: 30
        },
        tooltip: {
            headerFormat: '',
            shadow: false,
            style: {
                fontFamily: 'Monda, sans-serif',
                fontSize: '15px'
            }
        },
        legend: {
            enabled: false,
            borderWidth: 0,
            backgroundColor: "#f7f7f7",
            padding: 10
        }
    };

    var renderTTCchart = function(ttcURL) {
        $http.jsonp(ttcURL).
            success(function(response, status, headers, config) {
                var threshold = Math.round(Math.max(response.Threshold,1));
                var extended = _.map(response.WardData, function(val, key) { return _.extend(val,{'Ward':parseInt(key,10)}); });
                var filtered = _.filter(extended, function(ward) { return ward.Count >= threshold && ward.Ward > 0; });
                var sorted = _.sortBy(filtered, 'Time');
                var wards = _.pluck(sorted, 'Ward');
                var times = _.pluck(sorted, 'Time');
                var position = _.indexOf(wards, Data.wardNum);
                var colors = _.map(wards, function(ward) { return ward == Data.wardNum ? 'black' : '#BED0DE'; });

                Data.inTTCchart = position >= 0;
                Data.totalTTCWards = wards.length;
                Data.minTTCcount = threshold;

                if (Data.inTTCchart) {
                    Data.wardRank = window.getOrdinal(position + 1);
                    Data.wardTime = Math.round(sorted[position].Time * 100) / 100;
                }

                Highcharts.setOptions(highchartsDefaults);
                var ttcChart = new Highcharts.Chart({
                    chart: {
                        type: 'column',
                        renderTo: 'ttc-chart'
                    },
                    xAxis: {
                        labels: {
                            enabled: false
                        },
                        minPadding: 0.03,
                        maxPadding: 0
                    },
                    plotOptions: {
                        column: {
                            animation: false,
                            groupPadding: 0,
                            pointPadding: 0,
                            borderWidth: 0,
                            colorByPoint: true,
                            colors: colors
                        }
                    },
                    series: [{
                        name: "Time-to-close for " + Data.thisWeek,
                        data: times
                    }],
                    tooltip: {
                        formatter: function() {
                            var text = [
                                '<b>' + 'Ward ' + sorted[this.x].Ward + '<b>',
                                Math.round(this.y * 10) / 10 + ' day' + window.pluralize(this.y),
                                sorted[this.x].Count + ' request' + window.pluralize(sorted[this.x].Count)
                            ];
                            return text.join('<br>');
                        }
                    }
                });
            }
        );
    };

    var renderOverview = function(render) {
        var DAY_COUNT = 1;
        var highsURL = window.apiDomain + 'wards/' + window.wardNum + '/historic_highs.json?include_date=' + Data.date + '&count=' + DAY_COUNT + '&callback=JSON_CALLBACK';
        var ttcURL = window.apiDomain + 'requests/time_to_close.json?count=' + (Data.duration + 1) + '&end_date=' + Data.date + '&callback=JSON_CALLBACK';

        renderTTCchart(ttcURL);

        $http.jsonp(highsURL).
            success(function(response, status, headers, config) {
                var historicHighs = [];
                _.each(response.Highs, function(val, key) {
                    historicHighs.push({
                        'service': lookupCode(key).name,
                        'y': val ? val[0].Count: 0,
                        'name': val ? moment(val[0].Date).format("MMM D, 'YY") : '',
                        'current': response.Current[key].Count
                    });
                });
                historicHighs = _.sortBy(historicHighs, 'service');

                var categories = _.pluck(historicHighs, 'service');
                var current = _.pluck(historicHighs, 'current');

                var weekReviewChart = new Highcharts.Chart({
                    chart: {
                        type: 'line',
                        renderTo: 'weekReview-chart',
                        marginBottom: 50
                    },
                    series: [{
                        data: [4,6,8,5,3,5,2],
                        name: "Requests opened",
                        id: 1
                    },{
                        data: [14,8,1,3,9,2,7],
                        name: "Requests closed",
                        id: 1
                    }],
                    xAxis: {
                        categories: ['sun','sun','sun','sun','sun','sun','sun']
                    },
                    plotOptions: {
                        line: {
                            animation: false
                        }
                    },
                    legend: {
                        enabled: false
                    },
                    title: {
                        text: ''
                    },
                    tooltip: {
                        headerFormat: '',
                        shadow: false,
                        style: {
                            fontFamily: 'Monda, sans-serif',
                            fontSize: '15px'
                        }
                    }
                });

                if (render) {
                    var countsChart = new Highcharts.Chart({
                        chart: {
                            type: 'bar',
                            renderTo: 'highs-chart',
                            marginBottom: 30
                        },
                        series: [{
                            data: historicHighs,
                            name: "Historic high",
                            id: 1
                        }],
                        xAxis: {
                            categories: categories,
                            tickmarkPlacement: 'between',
                            labels: {
                                style: {
                                    fontFamily: 'Monda, sans-serif',
                                    fontSize: '15px'
                                },
                                useHTML: true,
                                y: 5
                            }
                        },
                        yAxis: {
                            opposite: true,
                            title: {
                                text: ''
                            }
                        },
                        plotOptions: {
                            bar: {
                                animation: false,
                                borderWidth: 0,
                                groupPadding: 0.08,
                                dataLabels: {
                                    color: "#000000",
                                    enabled: true,
                                    format: "{point.name}",
                                    style: {
                                        fontFamily: "Monda, Helvetica, sans-serif",
                                        fontSize: '12px'
                                    }
                                },
                                pointPadding: 0
                            }
                        },
                        legend: {
                            enabled: false
                        },
                        title: {
                            text: ''
                        },
                        tooltip: {
                            headerFormat: '',
                            shadow: false,
                            style: {
                                fontFamily: 'Monda, sans-serif',
                                fontSize: '15px'
                            }
                        }
                    });
                }
            });
    };

    var renderDetail = function (render) {
        var DAY_COUNT = 6;
        var highsURL = window.apiDomain + 'wards/' + window.wardNum + '/historic_highs.json?service_code=' + Data.serviceObj.code + '&include_date=' + Data.date + '&count=' + DAY_COUNT + '&callback=JSON_CALLBACK';
        var ttcURL = window.apiDomain + 'requests/time_to_close.json?count=7&service_code=' + Data.serviceObj.code + '&end_date=' + Data.date + '&callback=JSON_CALLBACK';

        renderTTCchart(ttcURL);

        $http.jsonp(highsURL).
            success(function(response, status, headers, config) {
                var todaysCount = _.last(response).Count;
                var highs = _.initial(response);
                var highCounts = _.pluck(highs, "Count");
                var categories = _.map(highs, function(d) {
                    var m = moment(d.Date);
                    return "<a href='/#/" + m.format(dateFormat) + "/" + Data.serviceObj.slug + "'>" + m.format("MMM D<br>YYYY") + "</a>";
                });

                if (render) {
                    Highcharts.setOptions(highchartsDefaults);
                    var countsChart = new Highcharts.Chart({
                        chart: {
                            type: 'column',
                            renderTo: 'counts-chart',
                            marginBottom: 70
                        },
                        plotOptions: {
                            column: {
                                animation: false,
                                groupPadding: 0.05,
                                pointPadding: 0,
                                color: "#4897F1"
                            }
                        },
                        series: [{
                            name: "All-time highs for Ward " + wardNum,
                            data: highCounts
                        }],
                        tooltip: {
                            formatter: function() {
                                return '<b>' + this.y + '</b> ' + ' request' + window.pluralize(this.y);
                            }
                        },
                        xAxis: {
                            categories: categories
                        }
                    });
                }
            });
    };

    var render = function (render) {
        if (Data.action == "overview") {
            renderOverview(render);
        } else if (Data.action == "detail") {
            renderDetail(render);
        }
    };

    $scope.data = Data;

    $scope.$on(
        "$routeChangeSuccess",
        function ($e, $currentRoute, $previousRoute) {
            Data.setDate(parseDate($routeParams.date, window.yesterday, $location));
            Data.action = $route.current.action;
            if (!$previousRoute || $previousRoute.redirectTo || $currentRoute.pathParams.serviceSlug != $previousRoute.pathParams.serviceSlug) {
                Data.serviceObj = {};
                if ($currentRoute.pathParams.serviceSlug) {
                    Data.serviceObj = window.lookupSlug($currentRoute.pathParams.serviceSlug);
                }
                render(true);
            } else {
                render(false);
            }
        }
    );
});

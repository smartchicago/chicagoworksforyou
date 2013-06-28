'use strict';

/* Controllers */

servicesMapApp.controller("servicesMapCtrl", function($scope, $http, $route, $routeParams) {
    $http.get('/data/services.json').success(function(data) {
        $scope.services = data;
    });

    $scope.orderProp = 'name';
    $scope.serviceTypeSlug = $routeParams.serviceSlug;
    $scope.serviceType = serviceTypesJSON[$routeParams.serviceSlug];
    $scope.date = $routeParams.date;

    $scope.calculateLayerSettings = function(wardNum, highest, lowest) {
        var fillOp = 0.1;
        var col = '#0873AD';

        if (wardNum == lowest[0]) {
            fillOp = 1;
        } else if (wardNum == highest[0]) {
            fillOp = 1;
            col = 'black';
        }

        var settings = {
            color: col,
            fillOpacity: fillOp
        };

        return settings;
    }

    $scope.updateST = function(isRedraw) {
        var st = $scope.serviceType;
        var numOfDays = 7;
        var url = window.apiDomain + 'requests/' + st.code + '/counts.json?end_date=' + currWeekEnd.format(dateFormat) + '&count=' + numOfDays + '&callback=?';

        $.getJSON(
            url,
            function(response) {
                var counts = _.pairs(response).slice(1,51);
                var sorted = _.sortBy(counts,function(pair) { return pair[1]; });

                var lowest = sorted[0];
                var highest = sorted[49];

                if (!isRedraw) {
                    window.allWards = L.layerGroup();

                    for (var path in wardPaths) {
                        var wardNum = parseInt(path, 10);
                        var poly = L.polygon(
                            wardPaths[path],
                            $.extend({
                                id: wardNum,
                                opacity: 1,
                                weight: 2
                            }, $scope.calculateLayerSettings(wardNum, highest, lowest))
                        );
                        poly.bindPopup('<a href="/wards/' + wardNum + '/">Ward ' + wardNum + '</a>');
                        window.allWards.addLayer(poly);
                    }

                    window.allWards.addTo(window.map);
                } else {
                    window.allWards.eachLayer(function(layer) {
                        layer.setStyle(calculateLayerSettings(layer.options.id, highest, lowest));
                    });
                }
            }
        );
    }

    drawChicagoMap();
    buildWardPaths();
    $scope.updateST(false);
});

//ServiceMapCtrl.$inject = ['$scope', '$http'];


function ServiceChartCtrl($scope, $routeParams) {
    $scope.serviceSlug = $routeParams.serviceSlug;
}

//PhoneDetailCtrl.$inject = ['$scope', '$routeParams'];

wardMapApp.controller("wardMapCtrl", function ($scope, $http) {
    $http.get('/data/services.json').success(function(data) {
        $scope.services = data;
    });

    $scope.orderProp = 'name';

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
        poly.bindPopup('<a href="/ward/' + wardNum + '/">Ward ' + wardNum + '</a>');
    }
});

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
    var serviceType = serviceTypesJSON[$routeParams.serviceSlug];
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

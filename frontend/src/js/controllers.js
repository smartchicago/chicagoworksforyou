// 'use strict';

// HOMEPAGE

// SERVICE MAP

serviceMapApp.controller("servicesMapCtrl", function($scope, $http, $route, $routeParams) {
    $http.get('/data/services.json').success(function(data) {
        $scope.services = data;
    });

    $scope.orderProp = 'name';
    $scope.serviceTypeSlug = $routeParams.serviceSlug;
    $scope.serviceType = window.lookupSlug($routeParams.serviceSlug);
    $scope.date = $routeParams.date;
    $scope.prevST = window.prevST($scope.serviceTypeSlug);
    $scope.nextST = window.nextST($scope.serviceTypeSlug);

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
                var counts = _.rest(_.pairs(response));
                var sorted = _.sortBy(counts,function(pair) { return pair[1].Count; });

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
                        poly.bindPopup('<a href="/ward/' + wardNum + '/">Ward ' + wardNum + '</a>');
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

    $scope.updateST(false);
});

//ServiceMapCtrl.$inject = ['$scope', '$http'];

// SERVICE DETAIL

serviceApp.controller("serviceListCtrl", function ($scope, $http, $location) {
    $http.get('/data/services.json').success(function(data) {
        $scope.services = data;
    });
    $scope.orderProp = 'name';
});

serviceApp.controller("serviceCtrl", function ($scope, $http, $routeParams) {
});

// WARD MAP

wardMapApp.controller("serviceListCtrl", function ($scope, $http, $location) {
    $http.get('/data/services.json').success(function(data) {
        $scope.services = data;
    });
    $scope.orderProp = 'name';

    $scope.isActive = function(slug) {
        var currServiceSlug = $location.path().substr(1);
        return slug == currServiceSlug;
    };
});

wardMapApp.controller("wardMapCtrl", function ($scope, $http) {

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

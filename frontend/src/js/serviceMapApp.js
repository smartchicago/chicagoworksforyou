// JQUERY

$(function () {
    drawChicagoMap();
});

// ANGULAR

var serviceMapApp = angular.module('serviceMapApp', []);

serviceMapApp.config(function($routeProvider) {
    $routeProvider.
        when('/:serviceSlug', {
            controller: "servicesMapCtrl",
            templateUrl: "/views/service_map_info.html"
        }).
        when('/:serviceSlug/:date', {
            controller: "servicesMapCtrl",
            templateUrl: "/views/service_map_info.html"
        }).
        otherwise({
            redirectTo: '/graffiti_removal'
        });
});

serviceMapApp.controller("servicesMapCtrl", function($scope, $http, $routeParams) {
    var date = moment().subtract('days', 1).startOf('day'); // Last Saturday
    if ($routeParams.date) {
        date = moment($routeParams.date);
    }

    $scope.serviceTypeSlug = $routeParams.serviceSlug;
    $scope.serviceType = window.lookupSlug($routeParams.serviceSlug);
    $scope.date = date.format('MMMM DD, YYYY');
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
    };

    var st = $scope.serviceType;
    var numOfDays = 7;
    var url = window.apiDomain + 'requests/' + st.code + '/counts.json?end_date=' + currWeekEnd.format(dateFormat) + '&count=' + numOfDays + '&callback=JSON_CALLBACK';

    $http.jsonp(url).
        success(function(data, status, headers, config) {
            var counts = _.rest(_.pairs(data));
            var sorted = _.sortBy(counts,function(pair) { return pair[1].Count; });

            var lowest = sorted[0];
            var highest = sorted[49];

            if (window.allWards) {
                window.allWards.clearLayers();
            } else {
                window.allWards = L.layerGroup();
            }

            for (var path in wardPaths) {
                var wardNum = parseInt(path, 10) + 1;
                var poly = L.polygon(
                    wardPaths[path],
                    $.extend({
                        id: wardNum,
                        opacity: 1,
                        weight: 2
                    }, $scope.calculateLayerSettings(wardNum + 1, highest, lowest))
                );
                poly.bindPopup('<a href="/ward/' + wardNum + '/">Ward ' + wardNum + '</a>');
                window.allWards.addLayer(poly);
            }

            window.allWards.addTo(window.map);
        }
    );
});

//ServiceMapCtrl.$inject = ['$scope', '$http'];

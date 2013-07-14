// JQUERY

$(function () {
    drawChicagoMap();
});

// ANGULAR

var wardMapApp = angular.module('wardMapApp', []);

wardMapApp.factory('Data', function () {
    return {};
});

wardMapApp.config(function($routeProvider) {
    $routeProvider.
        when('/:serviceSlug', {
            controller: "wardMapCtrl",
            templateUrl: "/views/ward_map_info.html"
        }).
        when('/:serviceSlug/:date', {
            controller: "wardMapCtrl",
            templateUrl: "/views/ward_map_info.html"
        }).
        otherwise({
            redirectTo: '/graffiti_removal'
        });
});

wardMapApp.controller("serviceListCtrl", function ($scope, Data, $http, $location) {
    $http.get('/data/services.json').success(function(data) {
        Data.services = data;
    });

    $scope.data = Data;

    $scope.selectService = function (service) {
        $location.path(service);
    };
});

wardMapApp.controller("wardMapCtrl", function ($scope, Data, $http, $routeParams) {
    var date = moment().subtract('days', 1).startOf('day'); // Last Saturday
    if ($routeParams.date) {
        date = moment($routeParams.date);
    }

    Data.currServiceSlug = $routeParams.serviceSlug;
    Data.dateFormatted = date.format(dateFormat);
    Data.serviceType = window.lookupSlug($routeParams.serviceSlug);

    $scope.data = Data;

    var numOfDays = 7;
    var url = window.apiDomain + 'requests/' + Data.serviceType.code + '/counts.json?end_date=' + Data.dateFormatted + '&count=' + numOfDays + '&callback=JSON_CALLBACK';

    $http.jsonp(url).
        success(function(response, status, headers, config) {
            if (window.allWards) {
                window.allWards.clearLayers();
            } else {
                window.allWards = L.layerGroup();
            }

            for (var path in wardPaths) {
                var wardNum = parseInt(path,10);
                var poly = L.polygon(
                    wardPaths[path],
                    {
                        color: '#0873AD',
                        opacity: 1,
                        weight: 2,
                        fillOpacity: (((wardNum % 5) + 2) / 10)
                    }
                );
                poly.bindPopup('<a href="/ward/' + wardNum + '/">Ward ' + wardNum + '</a>');
                window.allWards.addLayer(poly);
            }

            window.allWards.addTo(window.map);
        }
    );
});

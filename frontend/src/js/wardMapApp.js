// JQUERY

$(function () {
    drawChicagoMap();
    buildWardPaths();

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
        ).addTo(window.map);
        poly.bindPopup('<a href="/ward/' + wardNum + '/">Ward ' + wardNum + '</a>');
    }
});

// ANGULAR

var wardMapApp = angular.module('wardMapApp', []);

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

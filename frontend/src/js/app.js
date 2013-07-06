// 'use strict';


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

var wardApp = angular.module('wardApp', []);

wardApp.config(function($routeProvider) {
    $routeProvider.
        when('/:serviceSlug', {
            controller: "wardCtrl",
            templateUrl: "/views/ward_charts.html"
        }).
        when('/:serviceSlug/:date', {
            controller: "wardCtrl",
            templateUrl: "/views/ward_charts.html"
        }).
        otherwise({
            redirectTo: '/graffiti_removal'
        });
});

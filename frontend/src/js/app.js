// 'use strict';


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

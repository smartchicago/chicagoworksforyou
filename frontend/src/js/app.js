'use strict';

var servicesMapApp = angular.module('servicesMapApp', []);

servicesMapApp.config(function($routeProvider) {
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

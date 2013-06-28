'use strict';

var app = angular.module('services', []);

app.config(['$routeProvider', function($routeProvider) {
    $routeProvider.
        when('/:serviceSlug/map', {templateUrl: '/views/service_map.html',  controller: ServiceMapCtrl}).
        when('/:serviceSlug', {templateUrl: '/views/service_chart.html', controller: ServiceChartCtrl}).
        otherwise({redirectTo: '/graffiti_removal/map'});
}]);

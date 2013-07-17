// ANGULAR

var mediaApp = angular.module('mediaApp', []);

mediaApp.config(function($routeProvider) {
    $routeProvider.
        when('/:mediaID', {
            controller: "mediaMapCtrl",
            templateUrl: "/views/media.html"
        }).
        when('/', {
            controller: "mediaMapCtrl",
            templateUrl: "/views/media.html"
        }).
        otherwise({
            redirectTo: '/'
        });
});

var url = window.apiDomain + 'requests/media.json?callback=JSON_CALLBACK';

mediaApp.controller("mediaMapCtrl", function ($scope, $http) {
    $http.jsonp(url).
        success(function(response, status, headers, config) {
            $scope.media = response;
        }
    );
});

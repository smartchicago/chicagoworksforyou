// ANGULAR

var mediaApp = angular.module('mediaApp', []);

mediaApp.factory('Data', function () {
    return {};
});

mediaApp.config(function($routeProvider) {
    $routeProvider.
        when('/:serviceSlug', {
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

mediaApp.controller("sidebarCtrl", function ($scope, Data, $http, $location) {
    $http.get('/data/services.json').success(function(response) {
        Data.services = response;
    });

    $scope.data = Data;
});


mediaApp.controller("mediaMapCtrl", function ($scope, $http, Data, $routeParams) {
    Data.currServiceSlug = "";
    Data.search = {};
    var url = window.apiDomain + 'requests/media.json?callback=JSON_CALLBACK';
    var serviceObj = window.lookupSlug($routeParams.serviceSlug);

    if (serviceObj) {
        Data.currServiceSlug = $routeParams.serviceSlug;
        Data.search.Service_name = serviceObj.name;
    }

    $scope.data = Data;

    $http.jsonp(url).
        success(function(response, status, headers, config) {
            $scope.media = response;
        }
    );
});

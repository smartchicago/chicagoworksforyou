// ANGULAR

var mediaApp = angular.module('mediaApp', []);

mediaApp.factory('Data', function () {
    var data = {
        currServiceSlug: "",
        search: {}
    };

    data.setService = function(slug, name) {
        data.currServiceSlug = slug;
        data.search.Service_name = name;
    };

    return data;
});

mediaApp.controller("sidebarCtrl", function ($scope, Data, $http, $location) {
    $http.get('/data/services.json').success(function(response) {
        Data.services = response;
    });

    $scope.data = Data;

    $scope.filterByService = function(service) {
        if (!service) {
            service = {slug:'', name:''};
        }
        $location.path(service.slug);
        Data.setService(service.slug, service.name);
    };
});

mediaApp.controller("mediaCtrl", function ($scope, $http, Data, $location) {
    var url = window.apiDomain + 'requests/media.json?callback=JSON_CALLBACK';
    var slug = $location.path().split("/")[1];
    var serviceObj = window.lookupSlug(slug);

    if (serviceObj) {
        Data.setService(slug, serviceObj.name);
    }

    $scope.data = Data;

    $http.jsonp(url).
        success(function(response, status, headers, config) {
            $scope.media = response;
        }
    );
});

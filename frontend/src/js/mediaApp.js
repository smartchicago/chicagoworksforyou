// ANGULAR

var mediaApp = angular.module('mediaApp', []).value('$anchorScroll', angular.noop);

mediaApp.filter('escape', function() {
    return window.encodeURIComponent;
});

mediaApp.factory('Data', function () {
    var defaultTitle = "Media | Chicago Works For You";

    var data = {
        currServiceSlug: "",
        search: {},
        pageTitle: defaultTitle
    };

    data.setService = function(slug, name) {
        data.currServiceSlug = slug;
        data.search.service_name = name;
        data.pageTitle = (data.search.service_name ? data.search.service_name + " | " : '') + defaultTitle;
        data.currURL = window.urlBase + (slug ? slug + '/' : '');
        data.shareText = 'Recent ' + (data.search.service_name ? data.search.service_name : '311') + ' photos in Chicago';
    };

    return data;
});

mediaApp.controller("headCtrl", function ($scope, Data) {
    $scope.data = Data;
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

    $scope.serviceCount = function(name) {
        if (Data.mediaGroupped) {
            return Data.mediaGroupped[name];
        }
    };
});

mediaApp.controller("mediaCtrl", function ($scope, $http, Data, $location) {
    var url = window.apiDomain + 'requests/media.json?days=14&callback=JSON_CALLBACK';
    var slug = $location.path().split("/")[1];
    var serviceObj = window.lookupSlug(slug);

    Data.setService(slug, serviceObj ? serviceObj.name : '');

    $scope.data = Data;

    $http.jsonp(url).
        success(function(response, status, headers, config) {
            Data.media = response;
            Data.mediaGroupped = _.countBy(response,'service_name');
        }
    );
});

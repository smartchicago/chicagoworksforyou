// ANGULAR

var mediaApp = angular.module('mediaApp', []).value('$anchorScroll', angular.noop);

mediaApp.filter('escape', function() {
    return window.encodeURIComponent;
});

mediaApp.factory('Data', ['dateFilter', function (dateFilter) {
    var defaultTitle = "Media | Chicago Works For You";

    var data = {
        currServiceSlug: "",
        search: {}
    };
    data.pageTitle = getDescription(defaultTitle);

    function getDescription(defaultDescription) {
        var modifiers = [];
        if (data.search.requested_date) { modifiers.push( dateFilter(data.search.requested_date, 'MMMM d, yyyy')); }
        if (data.search.ward) { modifiers.push("Ward " + data.search.ward); }
        if (data.search.service_name) { modifiers.push(data.search.service_name); }

        return modifiers.length ? modifiers.join(", ") + " " + defaultDescription : defaultDescription;
    }

    data.encodeFilters = function(obj, prefix) {
      var str = [];
      for(var p in obj) {
        var k = prefix ? prefix + "[" + p + "]" : p, v = obj[p];
        if (!v) continue;
        str.push(typeof v == "object" ?
          serialize(v, k) :
          encodeURIComponent(k) + "=" + encodeURIComponent(v));
      }
      return str.join("&");
    };

    data.decodeFilters = function(query) {
        var match,
            pl     = /\+/g,  // Regex for replacing addition symbol with a space
            search = /([^&=]+)=?([^&]*)/g,
            decode = function (s) { return decodeURIComponent(s.replace(pl, " ")); };

        urlParams = {};
        while (match = search.exec(query))
           urlParams[decode(match[1])] = decode(match[2]);
        return urlParams;
    };

    data.update = function() {
        var slug = data.encodeFilters(data.search);
        data.pageTitle = getDescription(defaultTitle);
        data.shareText = 'Recent ' + getDescription('311') + ' photos in Chicago';
        data.currServiceSlug = slug;
        data.currURL = window.urlBase + (slug ? slug + '/' : '');
    };

    return data;
}]);

mediaApp.controller("headCtrl", function ($scope, Data) {
    $scope.data = Data;
});

mediaApp.controller("sidebarCtrl", function ($scope, Data, $http, $location) {
    $scope.data = Data;

    $scope.$watch('data', function (newValue) {
        Data.update();
        $location.path(Data.currServiceSlug);
    }, true);
});

mediaApp.controller("mediaCtrl", function ($scope, $http, Data, $location) {
    var url = window.apiDomain + 'requests/media.json?days=14&callback=JSON_CALLBACK';
    var slug = $location.path().split("/")[1];

    $scope.data = Data;

    $http.jsonp(url).
        success(function(response, status, headers, config) {
            Data.media = response;
            Data.wards = _.chain(response).groupBy('ward').keys().value();
            Data.serviceTypes = _.chain(response).groupBy('service_name').keys().value().sort();
            Data.requestedDates = _.chain(response).groupBy('requested_date').keys().value().sort();

            if (slug) {
                Data.search = Data.decodeFilters(slug);
                Data.update();
            }
        }
    );
});

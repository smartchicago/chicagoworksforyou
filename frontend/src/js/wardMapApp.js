// JQUERY

$(function () {
    drawChicagoMap();

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

wardMapApp.controller("wardMapCtrl", function ($scope, $http, $routeParams) {
    var date = moment().subtract('days', 1).startOf('day'); // Last Saturday
    if ($routeParams.date) {
        date = moment($routeParams.date);
    }

    $scope.serviceType = window.lookupSlug($routeParams.serviceSlug);
    $scope.date = date.format('MMMM DD, YYYY');

    var st = $scope.serviceType;
    var numOfDays = 7;
    var url = window.apiDomain + 'requests/' + st.code + '/counts.json?end_date=' + currWeekEnd.format(dateFormat) + '&count=' + numOfDays + '&callback=JSON_CALLBACK';


    $http.jsonp(url).
        success(function(data, status, headers, config) {
        }).
        error(function(data, status, headers, config) {
            // called asynchronously if an error occurs
            // or server returns response with an error status.
        }
    );
});

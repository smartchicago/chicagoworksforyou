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
                fillOpacity: 0.1
            }
        ).addTo(window.map);
        poly.bindPopup('<a href="/ward/' + wardNum + '/">Ward ' + wardNum + '</a>');
    }
});

// ANGULAR

var dateMapApp = angular.module('dateMapApp', []);

dateMapApp.config(function($routeProvider) {
    $routeProvider.
        when('/:date/:serviceSlug', {
            controller: "dateMapCtrl",
            templateUrl: "/views/date_info.html"
        }).
        when('/:date', {
            controller: "dateMapCtrl",
            templateUrl: "/views/date_info.html"
        }).
        when('/', {
            controller: "dateMapCtrl",
            templateUrl: "/views/date_info.html"
        }).
        otherwise({
            redirectTo: '/'
        });
});

dateMapApp.controller("dateMapCtrl", function ($scope, $http, $routeParams) {
    var date = moment().subtract('days', 1).startOf('day');
    if ($routeParams.date) {
        date = moment($routeParams.date);
    }

    window.date = date;

    $scope.date = date.format(dateFormat);
    $scope.dateFormatted = date.format('MMM D, YYYY');
    $scope.nextDay = "#/" + moment(date).add('days', 1).format(dateFormat);
    $scope.prevDay = "#/" + moment(date).subtract('days', 1).format(dateFormat);
    $scope.currURL = "#/" + date.format('YYYY-MM-DD');

    var url = window.apiDomain + 'requests/counts_by_day.json?day=' + date.format(dateFormat) + '&callback=JSON_CALLBACK';

    $http.jsonp(url).
        success(function(data, status, headers, config) {
            var mapped = _.map(_.pairs(data), function(pair) {
                service = _.find(serviceTypesJSON,function(obj) {return obj.code == pair[0];});
                return _.extend(pair[1], {
                    "Code": pair[0],
                    "Slug": service.slug,
                    "Name": service.name,
                    "Diff": Math.round(pair[1].Count - pair[1].Average)
                });
            });

            var split = _.groupBy(mapped, function(obj) {
                return obj.Diff > 0;
            });

            var aboveAverage = _.sortBy(split['true'], function(obj) {
                return obj.Diff;
            }).reverse();

            var belowAverage = _.sortBy(split['false'], function(obj) {
                return obj.Diff;
            }).reverse();

            $scope.aboveAverage = aboveAverage;
            $scope.belowAverage = belowAverage;

            if (window.allWards) {
                window.allWards.clearLayers();
            } else {
                window.allWards = L.layerGroup();
            }

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
                window.allWards.addLayer(poly);
            }

            window.allWards.addTo(window.map);
        }).
        error(function(data, status, headers, config) {
        // called asynchronously if an error occurs
        // or server returns response with an error status.
        }
    );
});

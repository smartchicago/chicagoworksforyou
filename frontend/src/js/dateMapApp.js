// JQUERY

$(function () {
    drawChicagoMap();
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

dateMapApp.controller("dateMapCtrl", function ($scope, $http, $location, $routeParams) {
    var date = parseDate($routeParams.date, window.yesterday, $location);
    var prevDay = moment(date).subtract('days', 1);
    var nextDay = moment(date).add('days', 1);
    var countsURL = window.apiDomain + 'requests/counts_by_day.json?day=' + date.format(dateFormat) + '&callback=JSON_CALLBACK';

    $scope.date = date.format(dateFormat);
    $scope.dateFormatted = date.format('MMM D, YYYY');
    $scope.prevDayFormatted = prevDay.format('MMM D');
    $scope.nextDayFormatted = nextDay.format('MMM D');
    $scope.serviceSlug = $routeParams.serviceSlug;
    $scope.currURL = "#/" + date.format(window.dateFormat);

    $scope.goToPrevDay = function() {
        if (prevDay.isBefore(window.earliestDate)) {
            return false;
        }
        $location.path(prevDay.format(dateFormat) + ($routeParams.serviceSlug ? '/' + $scope.serviceSlug : ''));
    };

    $scope.goToNextDay = function() {
        if (nextDay.isAfter(window.yesterday)) {
            return false;
        }
        $location.path(nextDay.format(dateFormat) + ($routeParams.serviceSlug ? '/' + $scope.serviceSlug : ''));
    };

    var calculateLayerSettings = function(ward, serviceData) {
        // serviceData is in form:
        // { Average, Code, Count, Diff, Name, Slug, Wards}
        // console.log("ward: %d", ward)
        // console.log("serviceData %o", serviceData)

        var maxFillOp = 0.9;
        var col = '#0873AD';

        var max = _.max(serviceData.Wards);
        var opac = ((serviceData.Wards[ward]) / (max)) * maxFillOp;

        return {
            color: col,
            fillOpacity: opac
        };
    };

    $http.jsonp(countsURL).
        success(function(data, status, headers, config) {
            var mapped = _.map(_.pairs(data), function(pair) {
                service = _.find(serviceTypesJSON, function(obj) { return obj.code == pair[0]; });
                return _.extend(pair[1], {
                    "Code": pair[0],
                    "Slug": service.slug,
                    "Name": service.name,
                    "Diff": Math.round(pair[1].Count - pair[1].Average)
                });
            });

            // mapped is of form:
            // [ { Average, Code, Count, Diff, Name, Slug, Wards}, ...]

            var serviceList = _.sortBy(mapped, function(obj) {
                return obj.Slug;
            });

            if (!$scope.serviceSlug) {
                //  FIXME: this should actually look at the average of the set?
                var max = _.max(serviceList, function(obj) { return obj.Diff; });
                $scope.serviceSlug = max.Slug;
            }

            $scope.serviceList = serviceList;

            if (window.allWards) {
                window.allWards.clearLayers();
            } else {
                window.allWards = L.layerGroup();
            }

            for (var path in wardPaths) {
                var wardNum = parseInt(path,10) + 1;
                var poly = L.polygon(
                    wardPaths[path],
                    _.extend({
                        id: wardNum,
                        opacity: 1,
                        weight: 2
                    }, calculateLayerSettings(wardNum, _.find(mapped, function(o) { return o.Slug == $scope.serviceSlug; })))
                ).addTo(window.map);

                poly.bindPopup('<a href="/ward/' + wardNum + '/#/' + $scope.serviceSlug + '/' + $scope.date + '">Ward ' + wardNum + '</a>');
                window.allWards.addLayer(poly);
            }

            window.allWards.addTo(window.map);
        }
    );
});

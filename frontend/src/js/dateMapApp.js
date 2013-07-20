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
    var serviceSlug = $routeParams.serviceSlug;

    $scope.date = date.format(dateFormat);
    $scope.dateFormatted = date.format('MMM D, YYYY');
    $scope.prevDayFormatted = prevDay.format('MMM D');
    $scope.nextDayFormatted = nextDay.format('MMM D');
    $scope.currURL = "#/" + date.format(window.dateFormat);

    $scope.goToPrevDay = function() {
        if (prevDay.isBefore(window.earliestDate)) {
            return false;
        }
        $location.path(prevDay.format(dateFormat));
    };

    $scope.goToNextDay = function() {
        if (nextDay.isAfter(window.yesterday)) {
            return false;
        }
        $location.path(nextDay.format(dateFormat));
    };

    var url = window.apiDomain + 'requests/counts_by_day.json?day=' + date.format(dateFormat) + '&callback=JSON_CALLBACK';

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

    $http.jsonp(url).
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

            var split = _.groupBy(mapped, function(obj) {
                //  FIXME: this should actually look at the average of the set
                return obj.Diff > 0;
            });

            var aboveAverage = _.sortBy(split['true'], function(obj) {
                return obj.Diff;
            }).reverse();

            var belowAverage = _.sortBy(split['false'], function(obj) {
                return obj.Diff;
            }).reverse();

            if (!serviceSlug) {
                var featuredService = aboveAverage[0] || belowAverage[0];
                $location.path(date.format(window.dateFormat) + "/" + featuredService.Slug);
            }

            $scope.aboveAverage = aboveAverage;
            $scope.belowAverage = belowAverage;

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
                    }, calculateLayerSettings(wardNum, _.find(mapped, function(o) { return o.Slug == serviceSlug; })))
                ).addTo(window.map);

                poly.bindPopup('<a href="/ward/' + wardNum + '/">Ward ' + wardNum + '</a>');
                window.allWards.addLayer(poly);
            }

            window.allWards.addTo(window.map);
        }
    );
});

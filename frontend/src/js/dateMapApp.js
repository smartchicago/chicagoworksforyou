// ANGULAR

var dateMapApp = angular.module('dateMapApp', []).value('$anchorScroll', angular.noop);

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

dateMapApp.controller("dateMapCtrl", function ($scope, $http, $location, $routeParams, $timeout) {
    var date = parseDate($routeParams.date, window.yesterday, $location, '');
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

    $scope.serviceClass = function(service) {
        var classes = [];
        if (service.Slug == $scope.serviceSlug) {
            classes.push('active');
            if (service.Slug == $routeParams.serviceSlug) {
                classes.push('in-url');
            }
        }
        if (service.Slug == $scope.maxService.Slug) {
            classes.push('max');
        }
        if (service.Count > service.Average) {
            classes.push('up');
        } else if (service.Count < service.Average) {
            classes.push('down');
        }
        return classes.join(" ");
    };

    $http.jsonp(countsURL).
        success(function(data, status, headers, config) {
            var mapped = _.map(_.pairs(data), function(pair) {
                service = _.find(serviceTypesJSON, function(obj) { return obj.code == pair[0]; });
                return _.extend(pair[1], {
                    "Code": pair[0],
                    "Slug": service.slug,
                    "Name": service.name,
                    "AvgRounded": Math.round(pair[1].Average * 10) / 10,
                    "Percent": Math.min(Math.round((pair[1].Count - pair[1].Average) * 100 / pair[1].Average), 100)
                });
            });

            // mapped is of form:
            // [ { Average, Code, Count, Percent, Name, Slug, Wards}, ...]

            var serviceList = _.sortBy(mapped, function(obj) {
                return obj.Slug;
            });

            $scope.maxService = _.max(serviceList, function(obj) { return obj.Percent; });
            if (!$scope.serviceSlug) {
                //  FIXME: this should actually look at the average of the set?
                $scope.serviceSlug = $scope.maxService.Slug;
            }

            $scope.serviceList = serviceList;

            var serviceObj = _.find(mapped, function(obj) {
                return obj.Slug == $scope.serviceSlug;
            });

            if (window.allWards) {
                window.allWards.clearLayers();
            } else {
                window.allWards = L.layerGroup();
            }

            var jenks_cutoffs = jenks(_.toArray(serviceObj.Wards), 5);
            jenks_cutoffs[0] = 0; // ensure the bottom value is 0
            jenks_cutoffs.pop(); // last item is the max value, so dont use it
            var wardColors = [
                '#265F7A',
                '#2F799B',
                '#3893BC',
                '#52A6CC',
                '#72B7D7',
                '#93C8E1'
            ].reverse();

            $timeout(function() {
                window.map = L.map('map',{scrollWheelZoom: false}).setView([41.83, -87.81], 11);

                L.tileLayer('http://{s}.tile.cloudmade.com/{key}/{styleId}/256/{z}/{x}/{y}.png', {
                    attribution: 'Map data &copy; <a href="http://openstreetmap.org">OpenStreetMap</a> contributors, <a href="http://creativecommons.org/licenses/by-sa/2.0/">CC-BY-SA</a>, Imagery © <a href="http://cloudmade.com">CloudMade</a>',
                    key: '302C8A713FF3456987B21FAAE639A13B',
                    maxZoom: 18,
                    styleId: 82946
                }).addTo(map);
                map.zoomControl.setPosition('bottomright');

                for (var path in wardPaths) {
                    var wardNum = parseInt(path,10) + 1;
                    var wardCount = serviceObj.Wards[wardNum];
                    var poly = L.polygon(
                        wardPaths[path],
                        {
                            id: wardNum,
                            opacity: 1,
                            weight: 1,
                            dashArray: '3',
                            color: 'white',
                            fillOpacity: 0.7,
                            fillColor: wardColors[_.sortedIndex(jenks_cutoffs, wardCount)]
                        }
                    ).addTo(window.map);
                    var requestCount = wardCount;
                    poly.bindPopup('<a href="/ward/' + wardNum + '/#/' + $scope.serviceSlug + '/' + $scope.date + '">Ward ' + wardNum + '</a>' + requestCount + ' request' + (requestCount > 1 ? 's' : ''));
                    window.allWards.addLayer(poly);
                }

                window.allWards.addTo(window.map);
            });
        }
    );
});

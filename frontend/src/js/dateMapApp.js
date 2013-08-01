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

    if (!window.chicagoMap) {
        window.chicagoMap = L.map('map',{scrollWheelZoom: false}).setView([41.83, -87.81], 11);

        L.tileLayer('http://{s}.tile.cloudmade.com/{key}/{styleId}/256/{z}/{x}/{y}.png', {
            attribution: 'Map data &copy; <a href="http://openstreetmap.org">OpenStreetMap</a> contributors, <a href="http://creativecommons.org/licenses/by-sa/2.0/">CC-BY-SA</a>, Imagery © <a href="http://cloudmade.com">CloudMade</a>',
            key: '302C8A713FF3456987B21FAAE639A13B',
            maxZoom: 18,
            styleId: 82946
        }).addTo(window.chicagoMap);
        window.chicagoMap.zoomControl.setPosition('bottomright');
    }

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

            var wardColors = [
                '#182A35',
                '#244153',
                '#315971',
                '#3C7090',
                '#4888AF',
                '#629CBF'
            ].reverse();

            var allCounts = _.toArray(serviceObj.Wards);
            var minCount = _.min(allCounts);
            var maxCount = _.max(allCounts);
            var hasRanges = maxCount >= wardColors.length;

            var grades = _.range(0, Math.min(wardColors.length, maxCount + 1));
            if (hasRanges) {
                grades = _.map(grades, function (grade) {
                    return Math.round(grade * maxCount / (grades.length - 1)) - 0.00001;
                });
            }

            var allColors = _.map(allCounts, function(count) {
                var pos = Math.max(_.sortedIndex(grades, count) - 1, 0);
                if (count == _.last(grades)) {
                    pos = grades.length - 1;
                }
                return wardColors[pos];
            });

            if (window.allWards) {
                window.allWards.clearLayers();
            } else {
                window.allWards = L.layerGroup();
            }

            var wardClick = function(e) {
                document.location = 'ward/' + e.target.options.id + '/#' + $location.path();
            };

            var pluralize = function(n) {
                return n == 1 ? '' : 's';
            };

            $timeout(function() {
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
                            fillOpacity: 1,
                            fillColor: allColors[wardNum-1]
                        }
                    )
                    .bindLabel('<h4>Ward ' + wardNum + '</h4>' + wardCount + ' request' + pluralize(wardCount))
                    .on('click', wardClick);
                    window.allWards.addLayer(poly);
                }

                if (window.legend) {
                    window.legend.removeFrom(window.chicagoMap);
                }
                window.legend = L.control({position: 'topright'});

                window.legend.onAdd = function(map) {
                    var div = L.DomUtil.create('div', 'legend');
                    var labels = _.map(grades, function (grade, i) {
                        var actualGrade = Math.round(grade);
                        return '<i style="background:' + wardColors[i] + '"></i> <b>' + actualGrade + (hasRanges && actualGrade < _.last(grades) ? '+': '') + "</b> request" + (grade == 1 && !hasRanges ? '' : 's');
                    });

                    div.innerHTML = labels.join('<br>');
                    return div;
                };

                window.allWards.addTo(window.chicagoMap);
                legend.addTo(window.chicagoMap);
            });
        }
    );
});

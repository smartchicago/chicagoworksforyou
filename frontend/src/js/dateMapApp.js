// ANGULAR

var dateMapApp = angular.module('dateMapApp', []).value('$anchorScroll', angular.noop);

dateMapApp.config(function($routeProvider) {
    $routeProvider.
        when('/', {
            action: "yesterday"
        }).
        when('/:date/', {
            action: "date"
        }).
        when('/:date/:serviceSlug/', {
            action: "date.service"
        }).
        otherwise({
            redirectTo: '/'
        });
});

dateMapApp.factory('Data', function () {
    var data = {
    };

    if (!window.chicagoMap) {
        window.chicagoMap = L.map('map',{scrollWheelZoom: false}).setView([41.80, -87.815], 11);

        L.tileLayer('http://{s}.tile.cloudmade.com/{key}/{styleId}/256/{z}/{x}/{y}.png', window.mapOptions)
            .addTo(window.chicagoMap);
        window.chicagoMap.zoomControl.setPosition('bottomright');
    }

    data.setDate = function(date) {
        data.date = date.format(dateFormat);
        data.dateISO = date.format();
        data.dateFormatted = date.format('MMM D, YYYY');

        data.prevDay = moment(date).subtract('day',1);
        data.nextDay = moment(date).add('day',1);
        data.prevDayFormatted = data.prevDay.format('MMM D');
        data.nextDayFormatted = data.nextDay.format('MMM D');
        data.isLatest = data.nextDay.isAfter(window.yesterday);
    };

    return data;
});

dateMapApp.controller("headCtrl", function ($scope, Data) {
    $scope.data = Data;
});

dateMapApp.controller("dateCtrl", function ($scope, Data, $http, $location, $route, $routeParams, $timeout) {
    var wardColors = [
        '#182A35',
        '#244153',
        '#315971',
        '#3C7090',
        '#4888AF',
        '#629CBF'
    ].reverse();

    var urlSuffix = function() {
        return $routeParams && $routeParams.serviceSlug ? Data.serviceObj.slug + '/' : '';
    };

    $scope.goToPrevDay = function() {
        if (Data.prevDay.isBefore(window.earliestDate)) {
            return false;
        }
        $location.path(Data.prevDay.format(dateFormat) + '/' + urlSuffix());
    };

    $scope.goToNextDay = function() {
        if (Data.isLatest) {
            return false;
        }
        $location.path(Data.nextDay.format(dateFormat) + '/' + urlSuffix());
    };

    $scope.serviceClass = function(service) {
        var classes = [];
        if (service.Slug == Data.serviceObj.slug) {
            classes.push('active');
            if (service.Slug == $routeParams.serviceSlug) {
                classes.push('in-url');
            }
        }
        if (service.Slug == Data.maxService.Slug) {
            classes.push('max');
        }
        if (service.Count > service.Average) {
            classes.push('up');
        } else if (service.Count < service.Average) {
            classes.push('down');
        }
        return classes.join(" ");
    };

    var renderMap = function() {
        var serviceObj = _.find(Data.serviceList, function(obj) {
            return obj.Slug == Data.serviceObj.slug;
        });

        var allCounts = _.toArray(serviceObj.Wards);
        var maxCount = _.max(allCounts);
        var hasRanges = maxCount >= wardColors.length;

        var grades = _.range(0, Math.min(wardColors.length, maxCount + 1));
        if (hasRanges) {
            grades = _.map(grades, function (grade) {
                return Math.round(grade * maxCount / (grades.length - 1));
            });
        }

        var allColors = _.map(allCounts, function(count) {
            // Shift the count to be between grades, find the index (which'll never be 0), then move it back one slot.
            var pos = _.sortedIndex(grades, count + 0.1) - 1;
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
                        fillOpacity: 0.8,
                        fillColor: allColors[wardNum-1]
                    }
                )
                .bindLabel('<h4>Ward ' + wardNum + '</h4>' + wardCount + ' request' + window.pluralize(wardCount))
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
                    return '<i style="background:' + wardColors[i] + '"></i> <b>' + grade + (hasRanges && grade < _.last(grades) ? '+': '') + "</b> request" + (grade == 1 && !hasRanges ? '' : 's');
                });

                div.innerHTML =
                    '<h4>' + Data.serviceObj.name + '</h4>' +
                    labels.join('<br>');
                return div;
            };

            window.allWards.addTo(window.chicagoMap);
            legend.addTo(window.chicagoMap);
        });
    };

    var changeDate = function() {
        var countsURL = window.apiDomain + 'requests/counts_by_day.json?day=' + Data.date + '&callback=JSON_CALLBACK';

        $http.jsonp(countsURL).
            success(function(response, status, headers, config) {
                Data.reponse = response;

                var serviceCollection = _.map(_.pairs(response), function(pair) {
                    service = _.find(serviceTypesJSON, function(obj) { return obj.code == pair[0]; });
                    return _.extend(pair[1], {
                        "Code": pair[0],
                        "Slug": service.slug,
                        "Name": service.name,
                        "AvgRounded": Math.round(pair[1].Average * 10) / 10,
                        "Percent": Math.min(Math.round((pair[1].Count - pair[1].Average) * 100 / pair[1].Average), 100)
                    });
                });

                Data.serviceList = _.sortBy(serviceCollection, function(obj) {
                    return obj.Slug;
                });

                Data.maxService = _.max(Data.serviceList, function(obj) { return obj.Percent; });
                if (_.isEmpty(Data.serviceObj)) {
                    Data.serviceObj = lookupSlug(Data.maxService.Slug);
                }

                renderMap();
            }
        );
    };

    $scope.data = Data;

    $scope.$on(
        "$routeChangeSuccess",
        function ($e, $currentRoute, $previousRoute) {
            Data.setDate(parseDate($routeParams.date, window.yesterday, $location));
            Data.serviceObj = {};
            Data.action = $route.current.action;
            if ($currentRoute.pathParams.serviceSlug) {
                if (Data.action == 'date.service') {
                    Data.serviceObj = window.lookupSlug($currentRoute.pathParams.serviceSlug);
                }
            }
            Data.currURL = "#/" + Data.date + "/" + urlSuffix();

            var titleParts = [];
            if (_.isEmpty($currentRoute.pathParams)) {
                titleParts.push('Chicago Works For You');
                titleParts.push('The citywide dashboard with ward-by-ward views of service delivery in Chicago');
            } else {
                titleParts.push(Data.dateFormatted);
                if ($currentRoute.pathParams.serviceSlug) {
                    titleParts.push(Data.serviceObj.name);
                }
                titleParts.push('Chicago Works For You');
            }
            Data.pageTitle = titleParts.join(' | ');

            if (!$previousRoute) {
                // First load
                changeDate();
            } else if ($currentRoute.pathParams.date != $previousRoute.pathParams.date) {
                // Going from one date to another
                changeDate();
            } else if (!$currentRoute.pathParams.serviceSlug) {
                // Going from defined date and service to the same date, but no service
                changeDate();
            } else if ($currentRoute.pathParams.serviceSlug != $previousRoute.pathParams.serviceSlug) {
                // Going from one service to another, with the same date
                renderMap();
            } else {
                changeDate();
            }
        }
    );
});

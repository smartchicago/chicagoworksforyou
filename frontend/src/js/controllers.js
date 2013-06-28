'use strict';

/* Controllers */

function ServiceListCtrl($scope, $http) {
    $http.get('/data/services.json').success(function(data) {
        $scope.services = data;
    });

    $scope.orderProp = 'name';
}

//ServiceListCtrl.$inject = ['$scope', '$http'];

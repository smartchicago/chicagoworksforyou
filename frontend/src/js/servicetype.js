var currWeekEnd = moment().day(-1).startOf('day');
var dateFormat = 'YYYY-MM-DD';
var weekDuration = moment.duration(6,"days");

function redrawChart() {
    var stCode = currServiceType;
    var numOfDays = 7;
    var url = 'http://cwfy-api-staging.smartchicagoapps.org/requests/' + stCode + '/counts.json?end_date=' + currWeekEnd.format(dateFormat) + '&count=' + numOfDays + '&callback=?';
    var chart = $('#chart').highcharts();

    $.getJSON(
        url,
        function(response) {
            var cityAverage = response['0'] / 50;
            var counts = _.pairs(response).slice(1,51);

            var sorted = _.sortBy(counts,function(pair) { return pair[1]; }).reverse();
            var sortedCategories = _.map(sorted, function(pair) { return "Ward " + pair[0]; });
            var sortedFakeWardAverages = _.map(sorted, function(pair) { return Math.max(pair[1] - Math.ceil((Math.random() - 0.5) * 20),0); });

            if (chart) {
                chart.get('counts').setData(counts);
                chart.yAxis[0].removePlotLine('avg');
                chart.yAxis[0].addPlotLine({
                    id: 'avg',
                    value: cityAverage,
                    color: 'brown',
                    width: 3,
                    zIndex:5
                });
            } else {
                var categories = _.map(counts, function(pair) { return "Ward " + pair[0]; });
                var fakeWardAverages = _.map(counts, function(pair) { return Math.max(pair[1] - Math.ceil((Math.random() - 0.5) * 20),0); });

                new Highcharts.Chart({
                    chart: {
                        renderTo: 'chart'
                    },
                    series: [{
                        data: counts,
                        id: 'counts',
                        index: 2,
                        dataLabels: {
                            style: {
                                fontWeight: 'bold'
                            }
                        }
                    },{
                        data: fakeWardAverages,
                        index: 1
                    }],
                    xAxis: {
                        categories: categories
                    },
                    yAxis: {
                        plotLines: [{
                            id: 'avg',
                            value: cityAverage,
                            color: 'brown',
                            width: 3,
                            zIndex:5
                        }]
                    }
                });
            }

            var currWeek = weekDuration.beforeMoment(currWeekEnd,true);
            $('.this-week a').text(currWeek.format({implicitYear: false}));
        }
    );
}

$(function () {
    // WEEKNAV

    $('.this-week a').click(function(evt) {
        evt.preventDefault();
    });

    $('.next-week a').click(function(evt) {
        evt.preventDefault();
        currWeekEnd.add('week',1);
        redrawChart();
    });

    $('.prev-week a').click(function(evt) {
        evt.preventDefault();
        currWeekEnd.subtract('week',1);
        redrawChart();
    });

    // CHART SETTINGS

    Highcharts.setOptions({
        chart: {
            marginBottom: 40,
            type: 'bar'
        },
        title: {
            text: ''
        },
        xAxis: {
            tickmarkPlacement: 'between',
            labels: {
                style: {
                    fontFamily: 'Roboto, sans-serif',
                    fontSize: '13px'
                }
            }
        },
        yAxis: {
            title: {
                text: ''
            },
            minPadding: 0.1,
            labels: {
                style: {
                    fontFamily: 'Roboto, sans-serif',
                    fontWeight: 'bold'
                }
            }
        },
        plotOptions: {
            bar: {
                borderWidth: 0,
                groupPadding: 0.08,
                dataLabels: {
                    enabled: true,
                    color: "#000000",
                    style: {
                        fontFamily: "Roboto, sans-serif",
                        fontSize: '13px'
                    }
                },
                pointPadding: 0
            }
        },
        tooltip: {
            headerFormat: '',
            pointFormat: '<b>{point.y:,.0f}</b> tickets',
            shadow: false,
            style: {
                fontFamily: 'Roboto, sans-serif',
                fontSize: '15px'
            }
        },
        legend: {
            enabled: false
        }
    });

    // SET UP INITIAL CHART

    redrawChart();
});

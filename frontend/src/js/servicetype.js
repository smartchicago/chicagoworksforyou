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
            if (chart) {
                chart.destroy();
            }

            var avg = response['0'] / 50;
            var sorted = _.sortBy(_.pairs(response),function(pair) { return pair[1]; }).slice(0,50).reverse();
            var fakeAvgs = _.map(sorted, function(pair) { return Math.max(pair[1] - Math.ceil((Math.random() - 0.5) * 20),0); });
            var cats = _.map(sorted, function(pair) { return "Ward " + pair[0]; });

            chart = new Highcharts.Chart({
                chart: {
                    renderTo: 'chart'
                },
                series: [{
                    data: sorted,
                    index: 2,
                    dataLabels: {
                        style: {
                            fontWeight: 'bold'
                        }
                    }
                },{
                    data: fakeAvgs,
                    index: 1
                }],
                xAxis: {
                    categories: cats
                },
                yAxis: {
                    plotLines: [{
                        id: 'avg',
                        value: avg,
                        color: 'brown',
                        width: 3,
                        label: {
                            align: 'center',
                            style: {
                                color: 'gray'
                            }
                        },
                        zIndex:5
                    }]
                }
            });

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

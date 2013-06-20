var currWeekEnd = moment().day(-1).startOf('day');
var dateFormat = 'YYYY-MM-DD';
var weekDuration = moment.duration(6,"days");

function redrawChart() {
    var stCode = currServiceType;
    var numOfDays = 7;
    var url = 'http://cwfy-api-staging.smartchicagoapps.org/requests/' + stCode + '/counts.json?end_date=' + currWeekEnd.format(dateFormat) + '&count=' + numOfDays + '&callback=?';
    $.getJSON(
        url,
        function(response) {
            var avg = response['0'] / 50;
            var sorted = _.sortBy(_.pairs(response),function(pair) { return pair[1]; }).slice(0,50).reverse();
            var cats = _.map(sorted, function(pair) { return pair[0]; });

            chart.series[0].setData(sorted);
            chart.xAxis[0].setCategories(cats);
            chart.yAxis[0].removePlotLine('avg');
            chart.yAxis[0].addPlotLine({
                id: 'avg',
                value: avg,
                color: 'black',
                width: 3,
                label: {
                    align: 'center',
                    style: {
                        color: 'gray'
                    }
                },
                zIndex:5
            });

            var currWeek = weekDuration.beforeMoment(currWeekEnd,true);
            $('.this-week a').text(currWeek.format({implicitYear: false}));
        }
    );
}

$(function () {
    // CHARTS

    // CHARTS WEEKNAV

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

    Highcharts.setOptions({
        chart: {
            marginBottom: 80,
            type: 'bar'
        },
        title: {
            text: ''
        },
        xAxis: {
            minPadding: 0.05,
            maxPadding: 0.05,
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
                groupPadding: 0.1,
                dataLabels: {
                    enabled: true,
                    color: "#000000",
                    style: {
                        fontFamily: "Roboto, sans-serif",
                        fontSize: '13px'
                    }
                }
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
            enabled: true,
            borderWidth: 0,
            backgroundColor: "#f7f7f7",
            padding: 10
        }
    });

    window.chart = new Highcharts.Chart({
        chart: {
            renderTo: 'chart'
        },
        series: [{}]
    });

    redrawChart();
});

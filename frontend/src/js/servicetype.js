$(function () {
    window.currWeekEnd = moment(new Date(2013,05,15));
    var dateFormat = 'YYYY-MM-DD';
    var weekDuration = moment.duration(6,"days");
    var week = 7 * 24 * 60 * 60 * 1000;

    // CHARTS

    function redrawChart(response) {
        var currWeek = weekDuration.beforeMoment(currWeekEnd,true);
        $('.this-week a').text(currWeek.format({implicitYear: false}));
    }

    redrawChart();

    // CHARTS WEEKNAV

    $('.this-week a').click(function(evt) {
        evt.preventDefault();
    });

    $('.next-week a').click(function(evt) {
        evt.preventDefault();
        currWeekEnd.add('week',1);
        $.getJSON(
            // 'http://cwfy-api-staging.smartchicagoapps.org/wards/' + ward + '/counts.json?count=7&service_code=' + currServiceType + '&end_date=' + currWeekEnd.format(dateFormat) + '&callback=?',
            function(response) {redrawChart(response);}
        );
    });

    $('.prev-week a').click(function(evt) {
        evt.preventDefault();
        currWeekEnd.subtract('week',1);
        $.getJSON(
            // 'http://cwfy-api-staging.smartchicagoapps.org/wards/' + ward + '/counts.json?count=7&service_code=' + currServiceType + '&end_date=' + currWeekEnd.format(dateFormat) + '&callback=?',
            function(response) {redrawChart(response);}
        );
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
            },
            categories: [1,2,3]
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
                    inside: true,
                    align: 'right',
                    color: "#ffffff",
                    x: -5,
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

    var countsChart = new Highcharts.Chart({
        chart: {
            renderTo: 'chart'
        },
        series: [{
            data: [50, 61, 72, 81, 42, 39, -15, -44, -60, -100]
        }]
    });
});

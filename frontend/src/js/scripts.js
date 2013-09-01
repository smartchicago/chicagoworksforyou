window.yesterday = moment().subtract('days', 1);
window.prevSaturday = moment().day(-1).startOf('day');
window.earliestDate = moment('2008-01-01');
window.dateFormat = 'YYYY-MM-DD';

window.serviceTypesJSON = [
    {
        "slug": "graffiti_removal",
        "code": "4fd3b167e750846744000005",
        "name": "Graffiti Removal"
    },
    {
        "slug": "restaurant_complaint",
        "code": "4fd6e4ece750840569000019",
        "name": "Restaurant Complaint"
    },
    {
        "slug": "rodent_baiting_rat_complaint",
        "code": "4fd3b9bce750846c5300004a",
        "name": "Rodent Baiting / Rat Complaint"
    },
    {
        "slug": "tree_debris",
        "code": "4fd3bbf8e750846c53000069",
        "name": "Tree Debris"
    },
    {
        "slug": "abandoned_vehicle",
        "code": "4ffa4c69601827691b000018",
        "name": "Abandoned Vehicle"
    },
    {
        "slug": "street_light_1_out",
        "code": "4ffa9f2d6018277d400000c8",
        "name": "Street Light 1 / Out"
    },
    {
        "slug": "pavement_cavein_survey",
        "code": "4ffa971e6018277d4000000b",
        "name": "Pavement Cave-In Survey"
    },
    {
        "slug": "alley_light_out",
        "code": "4ffa9cad6018277d4000007b",
        "name": "Alley Light Out"
    },
    {
        "slug": "building_violation",
        "code": "4fd3bd72e750846c530000cd",
        "name": "Building Violation"
    },
    {
        "slug": "traffic_signal_out",
        "code": "4ffa9db16018277d400000a2",
        "name": "Traffic Signal Out"
    },
    {
        "slug": "street_cut_complaint",
        "code": "4ffa995a6018277d4000003c",
        "name": "Street Cut Complaints"
    },
    {
        "slug": "sanitation_code_violation",
        "code": "4fd3b750e750846c5300001d",
        "name": "Sanitation Code Violation"
    },
    {
        "slug": "pothole_in_street",
        "code": "4fd3b656e750846c53000004",
        "name": "Pothole in Street"
    },
    {
        "slug": "street_lights_all_out",
        "code": "4fd3bd3de750846c530000b9",
        "name": "Street Lights All / Out"
    }
];

window.mapOptions = {
    attribution: 'Map data &copy; <a href="http://openstreetmap.org">OpenStreetMap</a> contributors, <a href="http://creativecommons.org/licenses/by-sa/2.0/">CC-BY-SA</a>, Imagery © <a href="http://cloudmade.com">CloudMade</a>',
    key: '302C8A713FF3456987B21FAAE639A13B',
    maxZoom: 18,
    styleId: 82946
};

window.stSlugs = _.pluck(serviceTypesJSON, 'slug');

window.parseDate = function(passedDate, defaultDate, locationModule) {
    var date = defaultDate;
    if (passedDate) {
        date = moment(passedDate);
        if (!date.isValid()) {
            document.location = "./#/";
        } else if (date.isBefore(window.earliestDate)) {
            locationModule.path(window.earliestDate.format(dateFormat));
        } else if (date.isAfter(window.yesterday)) {
            locationModule.path(window.yesterday.format(dateFormat));
        }
    }
    return date;
};

window.lookupSlug = function(slug) {
    return _.find(serviceTypesJSON, function(obj) {return obj.slug == slug;});
};

window.lookupCode = function(code) {
    return _.find(serviceTypesJSON, function(obj) {return obj.code == code;});
};

window.getOrdinal = function(n) {
    var s = ["th","st","nd","rd"];
    var v = n % 100;
    return n + (s[(v - 20) % 10] || s[v] || s[0]);
};

window.pluralize = function(n) {
    return n == 1 ? '' : 's';
};

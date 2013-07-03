window.wardPaths = [];
window.wardPolys = [];

window.currWeekEnd = moment().day(-1).startOf('day');
window.dateFormat = 'YYYY-MM-DD';
window.weekDuration = moment.duration(6,"days");

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

window.stSlugs = _.pluck(serviceTypesJSON, 'slug');

window.lookupSlug = function(slug) {
    return _.find(serviceTypesJSON,function(obj) {return obj.slug == slug;});
};

window.buildWardPaths = function() {
    for (var ward in WARDS) {
        var points = WARDS[ward].simple_shape[0][0];
        var wardPath = [];
        for (var p in points) {
            var latlong = [points[p][1], points[p][0]];
            wardPath.push(latlong);
        }
        wardPaths[ward] = wardPath;
    }
};

window.drawChicagoMap = function() {
    window.map = L.map('map',{scrollWheelZoom: false}).setView([41.83, -87.81], 11);
    L.tileLayer('http://{s}.tile.cloudmade.com/{key}/997/256/{z}/{x}/{y}.png', {
        attribution: 'Map data &copy; <a href="http://openstreetmap.org">OpenStreetMap</a> contributors, <a href="http://creativecommons.org/licenses/by-sa/2.0/">CC-BY-SA</a>, Imagery © <a href="http://cloudmade.com">CloudMade</a>',
        key: '302C8A713FF3456987B21FAAE639A13B',
        maxZoom: 18
    }).addTo(map);
    map.zoomControl.setPosition('bottomright');
};

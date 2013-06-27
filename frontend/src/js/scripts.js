window.wardPaths = [];
window.wardPolys = [];

window.currWeekEnd = moment().day(-1).startOf('day');
window.dateFormat = 'YYYY-MM-DD';
window.weekDuration = moment.duration(6,"days");

window.serviceTypesJSON = {
    "graffiti_removal": {
        "code": "4fd3b167e750846744000005",
        "name": "Graffiti Removal"
    },
    "restaurant_complaint": {
        "code": "4fd6e4ece750840569000019",
        "name": "Restaurant Complaint"
    },
    "rodent_baiting_rat_complaint": {
        "code": "4fd3b9bce750846c5300004a",
        "name": "Rodent Baiting / Rat Complaint"
    },
    "tree_debris": {
        "code": "4fd3bbf8e750846c53000069",
        "name": "Tree Debris"
    },
    "abandoned_vehicle": {
        "code": "4ffa4c69601827691b000018",
        "name": "Abandoned Vehicle"
    },
    "street_light_1_out": {
        "code": "4ffa9f2d6018277d400000c8",
        "name": "Street Light 1 / Out"
    },
    "pavement_cavein_survey": {
        "code": "4ffa971e6018277d4000000b",
        "name": "Pavement Cave-In Survey"
    },
    "alley_light_out": {
        "code": "4ffa9cad6018277d4000007b",
        "name": "Alley Light Out"
    },
    "building_violation": {
        "code": "4fd3bd72e750846c530000cd",
        "name": "Building Violation"
    },
    "traffic_signal_out": {
        "code": "4ffa9db16018277d400000a2",
        "name": "Traffic Signal Out"
    },
    "street_cut_complaint": {
        "code": "4ffa995a6018277d4000003c",
        "name": "Street Cut Complaints"
    },
    "sanitation_code_violation": {
        "code": "4fd3b750e750846c5300001d",
        "name": "Sanitation Code Violation"
    },
    "pothole_in_street": {
        "code": "4fd3b656e750846c53000004",
        "name": "Pothole in Street"
    },
    "street_lights_all_out": {
        "code": "4fd3bd3de750846c530000b9",
        "name": "Street Lights All / Out"
    }
};

function buildWardPaths() {
    for (var ward in WARDS) {
        var points = WARDS[ward].simple_shape[0][0];
        var wardPath = [];
        for (var p in points) {
            var latlong = [points[p][1], points[p][0]];
            wardPath.push(latlong);
        }
        wardPaths[ward] = wardPath;
    }
}

function drawChicagoMap() {
    window.map = L.map('map',{scrollWheelZoom: false}).setView([41.83, -87.81], 11);
    L.tileLayer('http://{s}.tile.cloudmade.com/{key}/997/256/{z}/{x}/{y}.png', {
        attribution: 'Map data &copy; <a href="http://openstreetmap.org">OpenStreetMap</a> contributors, <a href="http://creativecommons.org/licenses/by-sa/2.0/">CC-BY-SA</a>, Imagery © <a href="http://cloudmade.com">CloudMade</a>',
        key: '302C8A713FF3456987B21FAAE639A13B',
        maxZoom: 18
    }).addTo(map);
    map.zoomControl.setPosition('bottomright');
}

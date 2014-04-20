var CENTRAL_SERVER_ADDR = "http://localhost:15340"

function getConnectionFromCServ(data, status) {
    console.log("managed to get a response from the server");
    if (data.Status != "OK") {
        console.log("central server not ready: retrying...");
        setTimeout(function(){$.get(CENTRAL_SERVER_ADDR, getConnectionFromCServ)}, 1000);
    } else {
        console.log("central server is ready");


    }
}

function ConnectionManager() {
    console.log("hello from connectionManager");
    $.get(CENTRAL_SERVER_ADDR, getConnectionFromCServ);


}


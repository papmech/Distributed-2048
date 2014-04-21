var CENTRAL_SERVER_ADDR = "http://localhost:15340"

function ConnectionManager() {
    console.log("hello from connectionManager");


    $.get(CENTRAL_SERVER_ADDR, this.getConnectionFromCServ);



}

ConnectionManager.prototype.connectToGameServer = function(hostport) {
    var connectionString = 'ws://' + hostport + "/";
    this.connection = new WebSocket(connectionString);
    this.connection.onopen = this.connectionOpenHandler;
    this.connection.onerror = this.connectionErrorHandler;
    this.connection.onmessage = this.connectionMessageHandler;

};

ConnectionManager.prototype.getConnectionFromCServ = function(data, status) {
    var key = "Status";
    var unpacked = JSON && JSON.parse(data) || $.parseJSON(data);

    if (unpacked.Status !== "OK") {
        console.log("status is " + data["Status"]);
        console.log("central server not ready: retrying...");
        var callback = arguments.callee;
        setTimeout(function(){$.get(CENTRAL_SERVER_ADDR, callback)}, 1000);
    } else {
        console.log("central server is ready");
        ConnectionManager.prototype.connectToGameServer(unpacked.Hostport);
    }
};

ConnectionManager.prototype.connectionOpenHandler = function() {
    console.log("connection to the gameserver open");
};

ConnectionManager.prototype.connectionErrorHandler = function() {
    console.log("error from connection with gameserver");
};

ConnectionManager.prototype.connectionMessageHandler = function() {
    console.log("message received from gameserver");
};
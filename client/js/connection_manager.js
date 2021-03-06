function ConnectionManager() {
    console.log("hello from connectionManager");
    this.events = {};
    this.connected = false;
    this.boardHasBeenSet = false;
}

ConnectionManager.prototype.on = function (event, callback) {
    console.log('on was called');
    if (!this.events[event]) {
        this.events[event] = [];
    }
    this.events[event].push(callback);
};

ConnectionManager.prototype.emit = function (event, data) {
    console.log('emit was called');
    var callbacks = this.events[event];
    if (callbacks) {
        callbacks.forEach(function (callback) {
            callback(data);
        });
    }
};

ConnectionManager.prototype.connectToGameServer = function (hostport) {
    var self = this;
    if ('WebSocket' in window) {
        console.log('Websocket supported');
    } else {
        alert('WebSocket notch supported');
    }
    var connectionString = 'ws://' + hostport + "/abc";
    console.log('connection string is' + connectionString);
    console.log('this is ' + this)
    this.connection = new WebSocket(connectionString);

    this.connection.onopen = function() {
        console.log("connection to the gameserver open");
        console.log(self);
        $("#connected-server").text(hostport);
        self.emit("connectionMade");
    }
    this.connection.onerror = function(error) {
        console.log("error from connection with gameserver");
    }
    this.connection.onmessage = function(e) {
        console.log("message received from gameserver");
        var data = JSON && JSON.parse(e.data) || $.parseJSON(e.data);
        console.log(data);
        if (!self.boardHasBeenSet) {
        self.boardHasBeenSet = true;
        $(".load-wrapper").css( "display", "none" );
        }
        self.emit("update", data);
    }
    this.connection.onclose = function() {
        console.log("Connection to the game server has been lost(Oh no, whyy?).");
        self.getConnectionFromCServ();
    }
};

ConnectionManager.prototype.getConnectionFromCServ = function () {
    console.log(this);

    var xmlHttp = null;
    xmlHttp = new XMLHttpRequest();
    //  xmlHttp.open( "GET", "http://128.237.201.5:25340", false );
    xmlHttp.open( "GET", CENTRAL_HOSTPORT, false );
    xmlHttp.send( null );
    data = xmlHttp.responseText;

    var key = "Status";
    var unpacked = JSON && JSON.parse(data) || $.parseJSON(data);

    if (unpacked.Status !== "OK") {
        console.log("status is " + data["Status"]);
        console.log("central server not ready: retrying...");
        var callback = arguments.callee;
        setTimeout(function(){$.get(CENTRAL_SERVER_ADDR, callback)}, 1000);
    } else {
        console.log("central server is ready");
        this.connectToGameServer(unpacked.Hostport);
    }
};

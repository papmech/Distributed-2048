function ConnectionManager() {
    console.log("hello from connectionManager");
    $.get("http://localhost:15340", function(data, status) {
        console.log("managed to get a response from the server")
        console.log(data)
    });
    console.log("Completed Get Request")
}
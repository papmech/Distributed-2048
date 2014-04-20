function ConnectionManager() {
    console.log("hello from connectionManager");
    $.get("http://localhost:15340", function(data, status) {
        alert("managed to get a response from the server")
        alert(data)
    });
    console.log("Completed Get Request")
}
function ConnectionManager() {
    console.log("hello from connectionManager");
    var testingscope = 5;
    $.get("http://localhost:15340", function(data, status) {
        console.log("managed to get a response from the server")
        console.log(data)
        testingscope = 6;
    });
    console.log("Completed Get Request")
    console.log(testingscope)
}
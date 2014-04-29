package main

import (
	"net/http"
)

func main() {
	http.Handle("/", http.FileServer(http.Dir("client")))
	http.ListenAndServe(":8888", nil)
}

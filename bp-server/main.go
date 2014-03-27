// +build !appengine

package main

import (
	"net/http"
	"strconv"
)

func main() {
	http.ListenAndServe(":"+strconv.Itoa(*port), nil)
}

// bp-server implements a Martini based server to serve static files generated from standard Go html templates
package main

import (
	"flag"
	"github.com/ArtnerC/blueprint"
	"github.com/codegangsta/martini"
	"net/http"
	"strconv"
	"strings"
)

var m *martini.Martini
var port *int

func init() {
	master := flag.String("m", "Master.html", "filename of the master template file")
	extra := flag.String("x", "", "comma separated list of extra files to use with templates")
	tdir := flag.String("dir", "templates", "directory that contains template files")
	sdir := flag.String("static", "static", "directory to serve static files from")
	notfound := flag.String("nf", "static/404.html", "location and name of the 404 Not Found page")
	port = flag.Int("port", 80, "port for the server to listen on")
	flag.Parse()

	extras := strings.Split(*extra, ",")

	if *extra == "" {
		blueprint.MustCompileDir(*master, *tdir)
	} else {
		blueprint.MustCompileDir(*master, *tdir, extras...)
	}
	blueprint.SaveGenerated("generated")

	m = martini.New()
	m.Use(martini.Recovery())
	m.Use(martini.Logger())
	m.Use(martini.Static("generated"))
	m.Use(martini.Static(*sdir))

	r := martini.NewRouter()
	r.NotFound(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusNotFound)
		http.ServeFile(rw, req, *notfound)
	})
	m.Use(r.Handle)

	http.Handle("/", m)
}

func main() {
	http.ListenAndServe(":"+strconv.Itoa(*port), nil)
}

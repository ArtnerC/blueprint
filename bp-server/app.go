// bp-server implements a Martini based server to serve static files generated from standard Go html templates
package main

import (
	"flag"
	"github.com/ArtnerC/blueprint"
	"github.com/codegangsta/martini"
	"log"
	"net/http"
	"strings"
)

var m *martini.Martini
var port *int

func init() {
	log.SetFlags(0)
	log.SetPrefix("[bp-server] ")

	master := flag.String("m", "Master.html", "filename of the master template file")
	extra := flag.String("x", "", "comma separated list of extra files to use with templates")
	tdir := flag.String("dir", "templates", "directory that contains template files")
	sdir := flag.String("static", "static", "directory to serve static files from")
	notfound := flag.String("nf", "static/404.html", "location and name of the 404 Not Found page")
	port = flag.Int("port", 80, "port for the server to listen on")
	flag.Parse()

	extras := strings.Split(*extra, ",")
	if *extra == "" {
		extras = nil
	}

	if err := blueprint.CompileDir(*master, *tdir, extras...); err != nil {
		log.Fatal(err)
	}
	blueprint.SaveGenerated("generated")
	blueprint.BeginWatching()

	//martini.Env = martini.Prod
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
	m.Action(r.Handle)

	http.Handle("/", m)
}

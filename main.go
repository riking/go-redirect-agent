package main

import (
	"flag"
	"log"
	"net/http"
	"net/url"
)

// "go" = 103 111
var listenAddr = flag.String("listen", "127.0.103.111:80", "ipv4 listen address")
var destination = flag.String("d", "", "redirect destination for go links")
var destGolinks = flag.Bool("golinks", false, "use www.golinks.io")

func main() {
	flag.Parse()

	if *destGolinks && *destination == "" {
		*destination = "https://www.golinks.io"
	}
	if *destination == "" {
		log.Fatalln("go link handler not set, provide a -d argument")
	}

	destTmpl, err := url.Parse(*destination)
	if err != nil {
		log.Fatalln("could not parse destination url:", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		u := *destTmpl
		u.Path = r.URL.Path
		u.RawPath = r.URL.RawPath
		u.RawQuery = r.URL.RawQuery

		http.Redirect(w, r, u.String(), 302)
	})

	err = http.ListenAndServe(*listenAddr, nil)
	log.Fatalln(err)
}

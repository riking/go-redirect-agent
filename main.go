package main

import (
	"flag"
	"log"
	"net/http"
	"net/url"
	"path"
)

// "go" = 103 111
var listenAddr = flag.String("listen", "127.0.103.111:80", "ipv4 listen address")
var listenAddr6 = flag.String("listen6", "[fd00:6874:7470:676f::1]:80", "ipv6 listen address")
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
	tmplHasPath := false
	if destTmpl.Path != "" && destTmpl.Path != "/" {
		tmplHasPath = true
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		u := *destTmpl
		if tmplHasPath {
			u.Path = path.Join(u.Path, r.URL.Path)
			if r.URL.RawPath != "" {
				u.RawPath = path.Join(u.Path, r.URL.RawPath)
			}
		} else {
			u.Path = r.URL.Path
			u.RawPath = r.URL.RawPath
		}
		u.RawQuery = r.URL.RawQuery

		http.Redirect(w, r, u.String(), 302)
	})

	exitCh := make(chan struct{})

	go func() {
		err = http.ListenAndServe(*listenAddr, nil)
		log.Fatalln(err)
		exitCh <- struct{}{}
	}()
	go func() {
		err = http.ListenAndServe(*listenAddr6, nil)
		log.Println(err)
	}()

	<-exitCh
}

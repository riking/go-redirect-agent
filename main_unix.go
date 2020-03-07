// +build !windows

package main

import (
	"flag"
	"log"
	"net/http"
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

	err := Setup(nil, *destination)
	if err != nil {
		log.Fatalln(err)
	}

	err = http.ListenAndServe(*listenAddr, nil)
	log.Fatalln(err)
}

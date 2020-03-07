package main

import (
	"net/http"
	"net/url"
	"path"
)

func Setup(mux *http.ServeMux, destination string) error {
	if mux == nil {
		mux = http.DefaultServeMux
	}

	destTmpl, err := url.Parse(destination)
	if err != nil {
		return err
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

	return nil
}

package main

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

const port = 8080

//go:embed static/* post/* templates/*
var f embed.FS

func main() {
	static, _ := fs.Sub(f, "static")

	fs := http.FileServer(http.FS(static))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/post/", postHandler)
	http.HandleFunc("/", serveTemplate)

	log.Println("listening on http://localhost:" + strconv.Itoa(port))
	err := http.ListenAndServe("localhost:"+strconv.Itoa(port), nil)
	if err != nil {
		log.Fatal(err)
	}
}

func serveTemplate(w http.ResponseWriter, r *http.Request) {
	tp, _ := fs.Sub(f, "templates")

	lp := "layout.html"
	if r.URL.Path == "/" {
		r.URL.Path = "index.html"
	}
	fp := strings.TrimPrefix(filepath.Clean(r.URL.Path), `\`)

	_, err := tp.Open(fp)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	tmpl, err := template.ParseFS(tp, lp, fp)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(500), 500)
		return
	}

	err = tmpl.ExecuteTemplate(w, "layout", nil)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(500), 500)
	}
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	tp, _ := fs.Sub(f, "templates")

	lp := "post-layout.html"
	fp := filepath.Join(strings.TrimPrefix(filepath.Clean(r.URL.Path), "/"))
	fp = strings.TrimPrefix(fp, `\`)
	fp = strings.ReplaceAll(fp, `\`, `/`)

	// load the named .txt file for processing
	txt := strings.ReplaceAll(fp, ".html", ".txt")

	_, err := f.ReadFile(txt)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	content, err := f.ReadFile(txt)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(500), 500)
		return
	}

	tmpl, err := template.ParseFS(tp, lp)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(500), 500)
		return
	}

	err = tmpl.ExecuteTemplate(w, "post-layout", string(content))
	if err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(500), 500)
	}
}

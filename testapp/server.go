package main

import (
	"github.com/gorilla/mux"
	"github.com/pilu/fresh/runner/runnerutils"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func main() {
	r := mux.NewRouter()
	r.PathPrefix("/js").Handler(http.FileServer(http.Dir("../")))
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if runnerutils.HasErrors() {
			runnerutils.RenderError(w)
		}

		fc, err := os.Open("index.html")
		if err != nil {
			panic(err)
		}
		ct, _ := ioutil.ReadAll(fc)
		w.Write(ct)
	})
	http.Handle("/", r)
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

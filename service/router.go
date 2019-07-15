// -*- Mode: Go; indent-tabs-mode: t -*-

package service

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// Router creates a Gorilla mux router
func Router() *mux.Router {

	router := mux.NewRouter()

	path := []string{".", "/static/"}
	fs := http.StripPrefix("/static/", http.FileServer(http.Dir(strings.Join(path, ""))))
	router.PathPrefix("/static/").Handler(fs)
	router.Handle("/ajax", http.HandlerFunc(IndexHandler)).Methods("GET")
	router.Handle("/", http.HandlerFunc(StaticHandler)).Methods("GET")
	router.Handle("/form", http.HandlerFunc(StaticFormHandler)).Methods("POST")

	return router
}

// -*- Mode: Go; indent-tabs-mode: t -*-

package service

import (
	"log"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/slimjim777/football/datastore"
)

const limit = 40

// Page is the page details for the web application
type Page struct {
	Date     string
	Bookings []datastore.Booking
	Limit    int
}

// BookingRequest is the JSON request to create or update a booking
type BookingRequest struct {
	Name    string `json:"name"`
	Date    string `json:"date"`
	Playing bool   `json:"playing"`
}

var indexTemplate = "/static/app.html"
var staticTemplate = "/static/static.html"

// IndexHandler is the front page of the web application
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	page := Page{}

	path := []string{".", indexTemplate}
	t, err := template.ParseFiles(strings.Join(path, ""))
	if err != nil {
		log.Printf("Error loading the application template: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, page)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// StaticHandler is the front page of the static web page
func StaticHandler(w http.ResponseWriter, r *http.Request) {
	page := Page{Limit: limit}

	path := []string{".", staticTemplate}
	t, err := template.ParseFiles(strings.Join(path, ""))
	if err != nil {
		log.Printf("Error loading the application template: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the bookings for next Monday
	d := getDate()
	page.Date = d.Format("Mon 2 Jan 2006")
	bookings, err := datastore.BookingList(d.Format(time.RFC3339)[:10])
	if err != nil {
		log.Println("Error fetching bookings:", bookings)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	page.Bookings = []datastore.Booking{}
	for _, b := range bookings {
		if b.Playing {
			page.Bookings = append(page.Bookings, b)
		}
	}

	err = t.Execute(w, page)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// StaticFormHandler is the POST-ed form
func StaticFormHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("---", r.Header)

	if len(r.FormValue("name")) == 0 {
		http.Redirect(w, r, "/vintage", http.StatusFound)
		return
	}

	d := getDate()
	name := strings.TrimSpace(r.FormValue("name"))

	err := datastore.BookingUpsert(name, d.Format(time.RFC3339)[:10], r.FormValue("playing") == "playing", r.Header.Get("X-Forwarded-For"))
	if err != nil {
		log.Printf("Error with booking: %v\n", err)
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func getDate() time.Time {
	t := time.Now()

	day := (1 + 7 - int(t.Weekday())) % 7

	return t.Add(time.Duration(day*24) * time.Hour)
}

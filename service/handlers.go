// -*- Mode: Go; indent-tabs-mode: t -*-

package service

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/gorilla/mux"
	"github.com/slimjim777/football/datastore"
)

const limit = 20

// Page is the page details for the web application
type Page struct {
	Date     string
	Bookings []datastore.Booking
	Limit    int
}

// VersionResponse is the JSON response from the API Version method
type VersionResponse struct {
	Version string `json:"version"`
}

// BookingRequest is the JSON request to create or update a booking
type BookingRequest struct {
	Name    string `json:"name"`
	Date    string `json:"date"`
	Playing bool   `json:"playing"`
}

// StandardResponse is the JSON response from the API
type StandardResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// BookingResponse is the JSON response from the API for the booking list
type BookingResponse struct {
	StandardResponse
	Bookings []datastore.Booking `json:"bookings"`
}

var indexTemplate = "/static/app.html"
var staticTemplate = "/static/static.html"

// VersionHandler is the API method to return the version of the service
func VersionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	response := VersionResponse{Version: datastore.Version}

	// Encode the response as JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding the version response: %v\n", err)
	}
}

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
	r.ParseForm()

	if len(r.FormValue("name")) == 0 {
		http.Redirect(w, r, "/vintage", http.StatusFound)
		return
	}

	d := getDate()

	err := datastore.BookingUpsert(r.FormValue("name"), d.Format(time.RFC3339)[:10], r.FormValue("playing") == "playing")
	if err != nil {
		log.Printf("Error with booking: %v\n", err)
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

// BookingHandler creates or updates a booking for an individual
func BookingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	booking := BookingRequest{}
	defer r.Body.Close()

	err := json.NewDecoder(r.Body).Decode(&booking)
	if err != nil {
		log.Printf("Error in decoding JSON booking: %v\n", err)
		standardResponse(false, "Error in the request", w)
		return
	}

	err = datastore.BookingUpsert(booking.Name, booking.Date, booking.Playing)
	if err != nil {
		log.Printf("Error in decoding JSON booking: %v\n", err)
		standardResponse(false, "Error in saving the response", w)
		return
	}

	standardResponse(true, "", w)
}

// BookingListHandler fetches the bookings for a date
func BookingListHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	vars := mux.Vars(r)

	bookings, err := datastore.BookingList(vars["date"])
	if err != nil {
		log.Printf("Error in getting the bookings: %v\n", err)
		standardResponse(false, "Error in getting the bookings", w)
		return
	}

	std := StandardResponse{true, ""}
	response := BookingResponse{std, bookings}

	// Encode the response as JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error forming the booking response.")
	}
}

func standardResponse(success bool, message string, w http.ResponseWriter) {
	if !success {
		w.WriteHeader(http.StatusBadRequest)
	}
	response := StandardResponse{Success: success, Message: message}

	// Encode the response as JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error forming the standard response.")
	}
}

func getDate() time.Time {
	t := time.Now()

	day := (1 + 7 - int(t.Weekday())) % 7

	return t.Add(time.Duration(day*24) * time.Hour)
}

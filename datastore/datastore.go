// -*- Mode: Go; indent-tabs-mode: t -*-

package datastore

import (
	"database/sql"
	"log"
	"os"
	"time"

	// PostgreSQL adapter
	_ "github.com/lib/pq"
)

// DB local database interface with our custom methods.
type DB struct {
	*sql.DB
}

var dbConnection *DB

// Version holds the version of the service
const Version = "0.1"

const (
	createBooking = `CREATE TABLE IF NOT EXISTS booking (
		id serial primary key not null,
		created timestamp DEFAULT NOW(),
		book_date date not null,
		name varchar(200) not null,
		playing boolean,
		modified timestamp DEFAULT NOW(),
		ip_address varchar(200) DEFAULT ''
	)`

	createBookingIndex = `CREATE UNIQUE INDEX IF NOT EXISTS booking_idx ON booking (book_date DESC, name)`

	upsertBooking = `
		INSERT INTO booking (book_date, name, playing, modified, ip_address)
		VALUES ($1, $2, $3, NOW(), $4)
		ON CONFLICT (book_date, name)
			DO UPDATE SET playing = $3, modified=NOW() WHERE booking.book_date=$1 AND booking.name=$2
	`

	listBookingsForDate = `SELECT id, created, book_date, name, playing, modified, ip_address from booking where book_date=$1 ORDER BY modified`
)

// Booking holds a booking
type Booking struct {
	ID        int       `json:"id"`
	Created   time.Time `json:"created"`
	Date      string    `json:"date"`
	Name      string    `json:"name"`
	Playing   bool      `json:"playing"`
	Modified  time.Time `json:"modified"`
	IPAddress string    `json:"ip_address"`
}

// CreateDatabase creates the database tables
func CreateDatabase() error {

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Error opening database: %q", err)
	}

	dbConnection = &DB{db}

	if _, err := dbConnection.Exec(createBooking); err != nil {
		return nil
	}

	if _, err := dbConnection.Exec(createBookingIndex); err != nil {
		return nil
	}

	return nil
}

// BookingUpsert upserts a booking for a date
func BookingUpsert(name, date string, playing bool, ipAddress string) error {

	if _, err := dbConnection.Exec(upsertBooking, date, name, playing, ipAddress); err != nil {
		log.Println("Upsert error:", err)
		return err
	}

	return nil
}

// BookingList fetches bookings for a date
func BookingList(date string) ([]Booking, error) {

	bookings := []Booking{}
	var rows *sql.Rows

	rows, err := dbConnection.Query(listBookingsForDate, date)
	if err != nil {
		log.Printf("Error retrieving bookings: %v\n", err)
		return bookings, err
	}

	defer rows.Close()
	for rows.Next() {
		b := Booking{}
		err = rows.Scan(&b.ID, &b.Created, &b.Date, &b.Name, &b.Playing, &b.Modified, &b.IPAddress)
		if err != nil {
			return bookings, err
		}
		bookings = append(bookings, b)
	}

	return bookings, nil
}

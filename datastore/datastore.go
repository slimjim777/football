// -*- Mode: Go; indent-tabs-mode: t -*-

package datastore

import (
	"database/sql"
	"log"
	"os"

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
		playing boolean
	)`

	createBookingIndex = `CREATE UNIQUE INDEX IF NOT EXISTS booking_idx ON booking (book_date DESC, name)`

	upsertBooking = `
		INSERT INTO booking (book_date, name, playing)
		VALUES ($1, $2, $3)
		ON CONFLICT (book_date, name)
			DO UPDATE SET playing = $3 WHERE booking.book_date=$1 AND booking.name=$2
	`
)

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
func BookingUpsert(name, date string, playing bool) error {

	if _, err := dbConnection.Exec(upsertBooking, date, name, playing); err != nil {
		return err
	}

	return nil
}

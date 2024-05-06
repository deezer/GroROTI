package model

import (
	"database/sql"
	"os"
	"testing"

	"github.com/rs/zerolog/log"
)

func TestInitDatabase(t *testing.T) {
	// Delete the existing database file if it exists
	if _, err := os.Stat("data/sqlite-database.db"); err == nil {
		err := os.Remove("data/sqlite-database.db")
		if err != nil {
			t.Fatal(err)
		}
	}

	// Check twice to make sure that both case (existing and non existing DB)
	for i := 0; i < 2; i++ {
		// Call the function to initialize the database
		db := InitDatabase()

		// Check if the returned database instance is not nil
		if db == nil {
			t.Errorf("InitDatabase returned nil database instance on iteration %d", i)
		}
	}

	// Cleanup
	err := os.Remove("data/sqlite-database.db")
	if err != nil {
		t.Fatal(err)
	}

}

func TestInitTables(t *testing.T) {
	log.Info().Msg("Open an in-memory SQLite database for testing tables creation")
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Call the function to initialize tables
	initTables(db)

	// Check if the "roti" table exists
	roti, err := db.Query("SELECT name FROM sqlite_schema WHERE type='table' AND name='roti'")
	if err != nil {
		t.Fatal(err)
	}
	defer roti.Close()

	if !roti.Next() {
		t.Error("Table 'roti' does not exist")
	}

}

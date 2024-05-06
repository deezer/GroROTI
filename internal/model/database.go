package model

import (
	"database/sql"
	"errors"
	"os"
	"strings"

	"github.com/rs/zerolog/log"

	_ "github.com/mattn/go-sqlite3"
)

var (
	ErrNoROTIMatchingThisID = errors.New("no ROTI matching this ID")
	ErrNoFreeIDs            = errors.New("couldn't find a free ID for this ROTI")
	sqliteDatabase          *sql.DB
)

func InitDatabase() *sql.DB {
	// TODO: add a flag to trash DB if it exists
	// os.Remove("sqlite-database.db")

	// checks if database exists first
	if _, err := os.Stat("data/sqlite-database.db"); errors.Is(err, os.ErrNotExist) {
		if _, err := os.Stat("data/"); errors.Is(err, os.ErrNotExist) {
			log.Info().Msg("Creating data/ dir")
			if err := os.MkdirAll("data", os.ModePerm); err != nil {
				log.Fatal().Msg(err.Error())
			}
			log.Info().Msg("Directory data/ created successfully")
		}
		log.Info().Msg("Creating data/sqlite-database.db database file")
		file, err := os.Create("data/sqlite-database.db")
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
		file.Close()
		log.Info().Msg("data/sqlite-database.db database file created")

		sqliteDatabase, _ = sql.Open("sqlite3", "data/sqlite-database.db")

		initTables(sqliteDatabase) // Create Database Tables
	} else {
		log.Info().Msg("Re-using existing data/sqlite-database.db database file")
		sqliteDatabase, _ = sql.Open("sqlite3", "data/sqlite-database.db")

		addMissingColumns(sqliteDatabase)
	}
	return sqliteDatabase
}

func initTables(db *sql.DB) {
	createROTITable := `CREATE TABLE roti (
		"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"rotiid" INTEGER NOT NULL,	
		"description" TEXT,
		"hide" INTEGER,
		"feedback" INTEGER DEFAULT 0,
        "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	  );`

	createVoteTable := `CREATE TABLE vote (
		"id" TEXT NOT NULL PRIMARY KEY,		
		"value" INTEGER,
		"roti" INTEGER,
		"feedback" TEXT
	  );`

	rotiStatement, err := db.Prepare(createROTITable)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	_, err = rotiStatement.Exec()
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	log.Info().Msg("'roti' table created")

	voteStatement, err := db.Prepare(createVoteTable)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	_, err = voteStatement.Exec()
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	log.Info().Msg("'vote' table created")
}

// addMissingColumns allows to update old databases missing columns (new features)
func addMissingColumns(db *sql.DB) {
	if !columnExists(db, "roti", "feedback") {
		addFeedbackToROTITable := `ALTER TABLE roti ADD COLUMN "feedback" INTEGER DEFAULT 0;`
		dbStatement, err := db.Prepare(addFeedbackToROTITable)
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
		_, err = dbStatement.Exec()
		if err != nil {
			log.Fatal().Msg(err.Error())
		}

		log.Info().Msg("'feedback' column added to 'roti' table")
	}

	if !columnExists(db, "vote", "feedback") {
		addFeedbackToVoteTable := `ALTER TABLE vote ADD COLUMN "feedback" TEXT;`
		dbStatement, err := db.Prepare(addFeedbackToVoteTable)
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
		_, err = dbStatement.Exec()
		if err != nil {
			log.Fatal().Msg(err.Error())
		}

		log.Info().Msg("'feedback' column added to 'vote' table")
	}

	if !columnExists(db, "roti", "created_at") {
		addCreatedAtToROTITable := `ALTER TABLE roti ADD COLUMN "created_at" TIMESTAMP;`
		dbStatement, err := db.Prepare(addCreatedAtToROTITable)
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
		_, err = dbStatement.Exec()
		if err != nil {
			log.Fatal().Msg(err.Error())
		}

		log.Info().Msg("'created_at' column added to 'roti' table")
	}

	// look for rows that don't have a value for created_at
	dbStatement := `UPDATE roti SET created_at = CURRENT_DATE WHERE created_at IS NULL;`
	_, err := db.Exec(dbStatement)
	if err != nil {
		log.Fatal().Msgf("Error setting today's date for empty created_at values: %s", err.Error())
	}
}

func columnExists(db *sql.DB, tableName, columnName string) bool {
	query := `
		SELECT sql
		FROM sqlite_master
		WHERE type = 'table' AND name = ?;
	`
	row := db.QueryRow(query, tableName)
	var createTableStmt string
	if err := row.Scan(&createTableStmt); err != nil {
		log.Fatal().Msg(err.Error())
	}

	return strings.Contains(createTableStmt, columnName)
}

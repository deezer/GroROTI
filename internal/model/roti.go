package model

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
)

var (
	ErrInvalidROTIID = errors.New("invalid ROTI ID")
)

type ROTIEntity struct {
	id          ROTIID
	description string
	hide        bool
	feedback    bool
}

type ROTIID int

type ShortROTIInfo struct {
	ID   ROTIID
	Desc string
}

func (id ROTIID) Int() int {
	return int(id)
}

func NewROTIID() (rotiID ROTIID) {
	//Use a trivial seed as there is no real cryptographic work here
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	min := 10000
	max := 99999

	rotiID = ROTIID(r.Intn(max-min+1) + min)

	return
}

func NewROTIEntity(id ROTIID, description string, hide, feedback bool) (roti ROTIEntity) {
	roti.id = id
	roti.description = description
	roti.hide = hide
	roti.feedback = feedback

	return
}

func GetROTI(rotiid ROTIID) (roti ROTIEntity, err error) {
	var description string
	var hide, feedback bool

	row, err := sqliteDatabase.Query("SELECT description,hide,feedback FROM roti WHERE rotiid =" + strconv.Itoa(int(rotiid)))
	if err != nil {
		log.Fatal().Msgf(err.Error())
	}
	defer row.Close()

	row.Next()
	err = row.Scan(&description, &hide, &feedback)
	if err != nil {
		return ROTIEntity{}, err
	}
	return NewROTIEntity(rotiid, description, hide, feedback), nil
}

func insertROTI(db *sql.DB, roti ROTIEntity) {
	id := int(roti.GetID())
	log.Info().Msgf("inserting ROTI record %d (%s) hidden:%t feedback:%t", id, roti.description, roti.hide, roti.feedback)
	insertROTISQL := `INSERT INTO ROTI(rotiid, description, hide, feedback) VALUES (?, ?, ?, ?)`
	statement, err := db.Prepare(insertROTISQL)

	if err != nil {
		log.Fatal().Msgf(err.Error())
	}
	_, err = statement.Exec(id, roti.description, roti.hide, roti.feedback)
	if err != nil {
		log.Fatal().Msgf(err.Error())
	}
}

func CreateROTI(description string, hide, feedback bool, clean int) (rotiID ROTIID) {
	// before doing anything, run a check to see if we can clean some old ROTIs
	log.Info().Msgf("searching for opportunistic cleaning on the ROTI database")
	cleanOldROTIs(sqliteDatabase, clean)

	// create a new ID, hope it's free
	rotiID = NewROTIID()
	roti, _ := GetROTI(rotiID)

	// if roti is nil, it means this rotiID is free
	// if not, we loop and try again
	loop := 0
	for roti != (ROTIEntity{}) && loop < 1000 {
		// iterate until we find a free ID
		rotiID = NewROTIID()
		roti, _ = GetROTI(rotiID)
		loop++
	}
	if loop == 1000 {
		//we couldn't find a free ID after 1000 tries
		panic(ErrNoFreeIDs)
	}

	insertROTI(sqliteDatabase, NewROTIEntity(rotiID, description, hide, feedback))

	return
}

func cleanOldROTIs(db *sql.DB, clean int) {
	// Generate a time.Time from the clean_over_time parameter
	cleanOverTime := time.Now().AddDate(0, 0, -clean)

	// Retrieve ROTIs older than 30 days
	rows, err := db.Query("SELECT id,rotiid FROM roti WHERE created_at < ?", cleanOverTime)
	if err != nil {
		log.Error().Msgf("error retrieving old ROTIs: %s", err.Error())
		return
	}
	defer rows.Close()

	var rotiIDs []struct {
		id     int
		rotiid int
	}

	for rows.Next() {
		var id int
		var rotiid int
		if err := rows.Scan(&id, &rotiid); err != nil {
			log.Error().Msgf("Error scanning ROTI row: %s", err.Error())
			continue
		}
		currentROTI := struct {
			id     int
			rotiid int
		}{
			id:     id,
			rotiid: rotiid,
		}
		rotiIDs = append(rotiIDs, currentROTI)
	}
	if err := rows.Err(); err != nil {
		log.Error().Msgf("error iterating through ROTI rows: %s", err.Error())
		return
	}

	// Delete associated votes
	for _, rotiID := range rotiIDs {
		_, err := db.Exec("DELETE FROM vote WHERE roti = ?", rotiID.id)
		if err != nil {
			log.Error().Msgf("error deleting votes for ROTI ID %d/%d: %s", rotiID.rotiid, rotiID.id, err.Error())
			continue
		}
		log.Info().Msgf("votes deleted for ROTI ID %d/%d", rotiID.rotiid, rotiID.id)
	}

	// Delete old ROTIs
	_, err = db.Exec("DELETE FROM roti WHERE created_at < ?", cleanOverTime)
	if err != nil {
		log.Error().Msgf("error cleaning old ROTIs: %s", err.Error())
		return
	}

	log.Info().Msgf("old ROTIs cleaned up successfully")
}

func ListROTIs() (rotis []ShortROTIInfo) {
	row, err := sqliteDatabase.Query("SELECT rotiid,description FROM roti WHERE hide = FALSE ORDER BY id DESC LIMIT 10;")
	if err != nil {
		log.Fatal().Msgf(err.Error())
	}
	defer row.Close()
	for row.Next() {
		var rotiid int
		var description string
		err := row.Scan(&rotiid, &description)
		if err != nil {
			log.Fatal().Msgf(err.Error())
		}
		shortInfo := ShortROTIInfo{ID: ROTIID(rotiid), Desc: description}
		rotis = append(rotis, shortInfo)
	}
	return
}

func CountROTIs() int {
	var count int
	err := sqliteDatabase.QueryRow("SELECT COUNT(*) FROM roti").Scan(&count)
	if err != nil {
		log.Fatal().Msgf(err.Error())
	}
	return count
}

func GetMaxROTIID() int {
	// Use sql.NullInt64 to handle NULL values
	var maxID sql.NullInt64
	err := sqliteDatabase.QueryRow("SELECT MAX(id) FROM roti").Scan(&maxID)
	if err != nil {
		log.Fatal().Msgf(err.Error())
	}

	if maxID.Valid {
		return int(maxID.Int64)
	} else {
		// Handle case where there are no rows in roti
		return 0
	}
}

func (currentROTI *ROTIEntity) GetID() ROTIID {
	return currentROTI.id
}

func (currentROTI *ROTIEntity) GetDescription() string {
	return currentROTI.description
}

func (currentROTI *ROTIEntity) IsHidden() bool {
	return currentROTI.hide
}

func (currentROTI *ROTIEntity) HasFeedback() bool {
	return currentROTI.feedback
}

func (currentROTI *ROTIEntity) AddVoteToROTI(value float64, feedback string) (err error) {
	currentVote, err := NewVoteEntity(value)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidVoteID, err)
	}
	insertVote(sqliteDatabase, currentVote, currentROTI.id, feedback)
	return nil
}

func (currentROTI *ROTIEntity) CountVotes() int {
	row, err := sqliteDatabase.Query("SELECT COUNT(*) FROM vote WHERE roti =" + strconv.Itoa(int(currentROTI.id)))
	if err != nil {
		log.Fatal().Msgf("couldn't connect to database")
	}
	defer row.Close()
	var count int
	for row.Next() {
		if err := row.Scan(&count); err != nil {
			log.Error().Msgf("couldn't scan values : %s", err.Error())
		}
	}
	return count
}

func (currentROTI *ROTIEntity) GetMinVote() float64 {
	var min sql.NullFloat64

	row, err := sqliteDatabase.Query("SELECT MIN(value) FROM vote WHERE roti =" + strconv.Itoa(int(currentROTI.id)))
	if err != nil {
		log.Fatal().Msgf("couldn't connect to database")
	}
	defer row.Close()

	for row.Next() {
		if err := row.Scan(&min); err != nil {
			log.Error().Msgf("couldn't scan values : %s", err.Error())
		}
	}
	return min.Float64
}

func (currentROTI *ROTIEntity) GetMaxVote() float64 {
	var max sql.NullFloat64

	row, err := sqliteDatabase.Query("SELECT MAX(value) FROM vote WHERE roti =" + strconv.Itoa(int(currentROTI.id)))
	if err != nil {
		log.Fatal().Msgf("couldn't connect to database")
	}
	defer row.Close()

	for row.Next() {
		if err := row.Scan(&max); err != nil {
			log.Error().Msgf("couldn't scan values : %s", err.Error())
		}
	}
	return max.Float64
}

func (currentROTI *ROTIEntity) VotesAverage() float64 {
	var avg sql.NullFloat64

	row, err := sqliteDatabase.Query("SELECT AVG(value) FROM vote WHERE roti =" + strconv.Itoa(int(currentROTI.id)))
	if err != nil {
		log.Fatal().Msgf("couldn't connect to database")
	}
	defer row.Close()

	for row.Next() {
		if err := row.Scan(&avg); err != nil {
			log.Error().Msgf("couldn't scan values : %s", err.Error())
		}
	}
	return math.Ceil(avg.Float64*100) / 100
}

func (currentROTI *ROTIEntity) ListFeedbacks() (feedbacks []string) {
	var feedback string
	var value float32
	row, err := sqliteDatabase.Query("SELECT value, feedback FROM vote WHERE roti =" + strconv.Itoa(int(currentROTI.id)))
	if err != nil {
		log.Fatal().Msgf("couldn't connect to database")
	}
	defer row.Close()
	for row.Next() {
		if err := row.Scan(&value, &feedback); err != nil {
			log.Error().Msgf("couldn't scan values : %s", err.Error())
		}
		if feedback != "" {
			feedbacks = append(feedbacks, fmt.Sprintf("(%.1f) %s", value, feedback))
		}
	}
	return
}

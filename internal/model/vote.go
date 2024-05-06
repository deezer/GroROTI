package model

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/rs/zerolog/log"

	"github.com/google/uuid"
)

var (
	ErrInvalidVoteID = errors.New("invalid vote ID")
	ErrInvalidVote   = errors.New("invalid vote value")
)

type VoteEntity struct {
	id    VoteID
	value float64
}

type VoteID string

func (id VoteID) String() string {
	return string(id)
}

func (currentVote *VoteEntity) ID() VoteID {
	return currentVote.id
}

func (currentVote *VoteEntity) GetVote() float64 {
	return currentVote.value
}

func NewVoteEntity(value float64) (vote VoteEntity, err error) {
	uuid, err := uuid.NewRandom()
	if err != nil {
		err = fmt.Errorf("%w: %s", ErrInvalidVoteID, err)
		return
	}
	vote.id = VoteID(uuid.String())
	vote.value = value

	return
}

func insertVote(db *sql.DB, vote VoteEntity, rotiid ROTIID, feedback string) {
	log.Info().Msgf("Inserting Vote record %s for ROTI %d", vote.id, int(rotiid))
	insertVoteSQL := `INSERT INTO vote(id, value, roti, feedback) VALUES (?, ?, ?, ?)`
	statement, err := db.Prepare(insertVoteSQL)

	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	_, err = statement.Exec(vote.id, vote.value, rotiid, feedback)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
}

func CheckVote(voteString string) (vote float64, err error) {
	vote, err = strconv.ParseFloat(voteString, 64)
	if err != nil {
		return 0, ErrInvalidVote
	} else {
		if vote < 1 || vote > 5 {
			return 0, ErrInvalidVote
		} else {
			return vote, nil
		}
	}
}

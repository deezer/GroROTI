package model

import (
	"errors"
	"testing"
)

func TestNewVoteEntity(t *testing.T) {
	initialVote := 8.1

	vote, err := NewVoteEntity(initialVote)
	if err != nil {
		t.Fatal(err)
	}

	if initialVote != vote.value {
		t.Errorf("Got %v value but expected %v", vote, initialVote)
	}
}

func initVoteTest(values []float64, feedbacks []string) (ROTIEntity, error) {
	if len(values) != len(feedbacks) {
		return ROTIEntity{}, errors.New("Both values and feedbacks lengthes must be equal to fill votes")
	}
	if err := removeData(); err != nil {
		return ROTIEntity{}, err
	}
	InitDatabase()
	defer removeData()

	roti := ROTIEntity{
		id:          NewROTIID(),
		description: "test",
		hide:        false,
		feedback:    true,
	}

	for i, value := range values {
		err := roti.AddVoteToROTI(value, feedbacks[i])
		if err != nil {
			return roti, err
		}
	}

	return roti, nil
}

func TestAddVoteToROTI(t *testing.T) {
	roti, err := initVoteTest([]float64{3.5}, []string{"test"})
	if err != nil {
		t.Fatal(err)
	}

	if roti.CountVotes() != 1 {
		t.Errorf("Got %d vote(s) but expected 1", roti.CountVotes())
	}
}

func TestCheckVote(t *testing.T) {
	testCases := []struct {
		voteString  string
		expected    float64
		expectedErr error
	}{
		{"3", 3.0, nil},            // Valid vote
		{"1.5", 1.5, nil},          // Valid vote with a decimal
		{"0", 0, ErrInvalidVote},   // Invalid vote (less than 1)
		{"6", 0, ErrInvalidVote},   // Invalid vote (greater than 5)
		{"abc", 0, ErrInvalidVote}, // Invalid vote (not a number)
	}

	for _, tc := range testCases {
		t.Run("TestCheckVote : "+tc.voteString, func(t *testing.T) {
			vote, err := CheckVote(tc.voteString)
			if vote != tc.expected {
				t.Errorf("Got %f but expected vote is %f", tc.expected, vote)
			}
			if err != tc.expectedErr {
				t.Errorf("Got %v but expected error is %v", tc.expectedErr, err)
			}
		})
	}
}

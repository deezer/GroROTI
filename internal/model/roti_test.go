package model

import (
	"os"
	"reflect"
	"testing"
)

func removeData() error {
	if _, err := os.Stat("data/sqlite-database.db"); err == nil {
		err := os.Remove("data/sqlite-database.db")
		if err != nil {
			return err
		}
	}
	return nil
}

func TestNewROTIID(t *testing.T) {
	min := 10000
	max := 99999

	rotiid := NewROTIID()

	if rotiid.Int() < min && max < rotiid.Int() {
		t.Errorf("rotiID isn't between min and max value.")
	}
}

func TestGetROTI(t *testing.T) {
	if err := removeData(); err != nil {
		t.Fatal(err)
	}
	InitDatabase()
	defer removeData()

	rotidesc := "test"
	rotiid := CreateROTI(rotidesc, false, false, 30)

	testedRoti, err := GetROTI(rotiid)
	if err != nil {
		t.Fatal(err)
	}

	if testedRoti.id.Int() != rotiid.Int() {
		t.Errorf("Got %d id but expected %d", testedRoti.id.Int(), rotiid.Int())
	}
	if testedRoti.description != "test" {
		t.Errorf("Got %s description but expected %s", testedRoti.description, rotidesc)
	}
}

func TestListROTIs(t *testing.T) {
	if err := removeData(); err != nil {
		t.Fatal(err)
	}
	InitDatabase()
	defer removeData()

	rotiid1 := CreateROTI("test1", false, false, 30)
	rotiid2 := CreateROTI("test2", false, false, 30)

	rotiList := []ShortROTIInfo{
		{ID: rotiid2, Desc: "test2"},
		{ID: rotiid1, Desc: "test1"},
	}

	testedRotiList := ListROTIs()

	if !reflect.DeepEqual(rotiList, testedRotiList) {
		t.Errorf("Got %v slice but expected %v", rotiList, testedRotiList)
	}
}

func TestGetMinVote(t *testing.T) {
	roti, err := initVoteTest([]float64{5, 6}, []string{"test", "test"})
	if err != nil {
		t.Fatal(err)
	}

	min := roti.GetMinVote()

	if min != 5.0 {
		t.Errorf("Got min = %v but expected 5.0", min)
	}
}

func TestGetMaxVote(t *testing.T) {
	roti, err := initVoteTest([]float64{5, 6}, []string{"test", "test"})
	if err != nil {
		t.Fatal(err)
	}

	max := roti.GetMaxVote()

	if max != 6.0 {
		t.Errorf("Got max = %v but expected 6.0", max)
	}
}

func TestVotesAverage(t *testing.T) {
	roti, err := initVoteTest([]float64{5, 6}, []string{"test", "test"})
	if err != nil {
		t.Fatal(err)
	}

	avg := roti.VotesAverage()

	if avg != 5.5 {
		t.Errorf("Got avg = %v but expected 5.5", avg)
	}
}

func TestListFeedbacks(t *testing.T) {
	testCases := []struct {
		values            []float64
		feedbacks         []string
		expectedFeedbacks []string
	}{
		{
			values:            []float64{4.5, 3.0},
			feedbacks:         []string{"Good Roti", "Okay Roti"},
			expectedFeedbacks: []string{"(4.5) Good Roti", "(3.0) Okay Roti"}, // Multiple feedbacks
		},
		{
			values:            []float64{2.5},
			feedbacks:         []string{"Bad Roti"},
			expectedFeedbacks: []string{"(2.5) Bad Roti"}, // One feedback
		},
	}

	for _, tc := range testCases {
		roti, err := initVoteTest(tc.values, tc.feedbacks)

		if err != nil {
			t.Fatal(err)
		}

		testedFeedbacks := roti.ListFeedbacks()

		if !reflect.DeepEqual(testedFeedbacks, tc.expectedFeedbacks) {
			t.Errorf("Got feedback = %v but expected list is %v", testedFeedbacks, tc.expectedFeedbacks)
		}
	}
}

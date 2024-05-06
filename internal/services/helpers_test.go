package services

import (
	"net/http"
	"testing"

	"github.com/deezer/groroti/internal/model"
	"github.com/rs/zerolog/log"
)

func TestGetIDFromURL(t *testing.T) {
	testCases := []struct {
		query          string
		rotiid         string
		legacy         bool
		expectedRotiID int
		expectedError  error
	}{
		{"/roti", "", true, 0, model.ErrInvalidROTIID},                // Without parameter / legacy
		{"/roti?r=aaaaa", "aaaaa", true, 0, nil},                      // Not a number / legacy
		{"/roti?r=99999", "99999", true, 99999, nil},                  // With good number / legacy
		{"/roti?r=100000", "100000", true, 0, model.ErrInvalidROTIID}, // With a BAD number / legacy
		{"/roti/", "", false, 0, model.ErrInvalidROTIID},              // Without parameter
		{"/roti/aaaaa", "aaaaaa", false, 0, nil},                      // Not a number
		{"/roti/99999", "99999", false, 99999, nil},                   // With good number
		{"/roti/99999", "100000", false, 0, model.ErrInvalidROTIID},   // With a BAD number
	}

	for _, tc := range testCases {
		t.Run(tc.query, func(t *testing.T) {
			req, err := http.NewRequest("GET", tc.query, nil)
			req.SetPathValue("rotiid", tc.rotiid)
			if err != nil {
				t.Fatal(err)
			}
			rotid, err := getIDFromURL(req, tc.legacy)

			if rotid != tc.expectedRotiID && err != tc.expectedError {
				log.Warn().Msg(err.Error())
				t.Errorf("Expected %d and got %d", tc.expectedRotiID, rotid)
			}
		})
	}
}

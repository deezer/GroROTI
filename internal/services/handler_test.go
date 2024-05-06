package services

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/deezer/groroti/internal/model"
	"github.com/deezer/groroti/internal/staticEmbed"
)

func initDatabaseAndTemplates() error {
	model.InitDatabase()
	err := staticEmbed.LoadTemplates()
	if err != nil {
		return err
	}
	return nil
}

func testHandler(query string, method string, handlerfunc func(http.ResponseWriter, *http.Request)) (int, error) {
	// Create a GET request
	req, err := http.NewRequest(method, query, nil)
	if err != nil {
		return 0, err
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Call the handler function
	handler := http.HandlerFunc(handlerfunc)
	handler.ServeHTTP(rr, req)

	return rr.Code, nil
}

func testRouter(query string, method string, router *http.ServeMux) (int, error) {
	// Create a GET request
	req, err := http.NewRequest(method, query, nil)
	if err != nil {
		return 0, err
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	return rr.Code, nil
}

func TestHomeHandler(t *testing.T) {
	if err := initDatabaseAndTemplates(); err != nil {
		t.Fatal(err)
	}

	expectedCode := 200
	code, err := testHandler("/", "GET", homeHandler)
	if err != nil {
		t.Fatal(err)
	}
	if code != expectedCode {
		t.Errorf("Handler returned wrong status code: got %d want %d", code, expectedCode)
	}
}

func generateTestsROTIs() (existingROTI int, nonExistingROTI int) {
	// create a roti and get the ID
	rotiID := model.CreateROTI("test", false, false, 30)
	existingROTI = rotiID.Int()

	// then create an id from a roti that doesn't exist
	if existingROTI != 99999 {
		nonExistingROTI = existingROTI + 1
	} else {
		// edge case, rotiID = 99999 so we can't +1 to get a valid ROTIID
		nonExistingROTI = existingROTI - 1
	}
	return
}

func TestDisplayROTIHandler(t *testing.T) {
	if err := initDatabaseAndTemplates(); err != nil {
		t.Fatal(err)
	}

	router := http.DefaultServeMux
	router.HandleFunc("/roti/{rotiid}", displayROTIHandler)

	existingROTI, nonExistingROTI := generateTestsROTIs()

	testCases := []struct {
		query              string
		legacy             bool
		expectedStatusCode int
	}{
		{"/roti", true, 406}, // Without r parameter
		{fmt.Sprintf("/roti?r=%d", nonExistingROTI), true, 308}, // With a roti (r) that doesn't exist, legacy
		{fmt.Sprintf("/roti?r=%d", existingROTI), true, 308},    // With a roti (r) that exists, legacy
		{"/roti/", false, 404},                                  // Without rotiid parameter
		{fmt.Sprintf("/roti/%d", nonExistingROTI), false, 406},  // With a roti (r) that doesn't exist
		{fmt.Sprintf("/roti/%d", existingROTI), false, 200},     // With a roti (r) that exists
	}

	for _, tc := range testCases {
		t.Run(tc.query, func(t *testing.T) {
			var code int
			var err error
			if tc.legacy {
				code, err = testHandler(tc.query, "GET", displayROTIHandlerLegacy)
			} else {
				code, err = testRouter(tc.query, "GET", router)
			}
			if err != nil {
				t.Fatal(err)
			}
			if code != tc.expectedStatusCode {
				t.Errorf("Handler returned wrong status code: got %d want %d", code, tc.expectedStatusCode)
			}
		})
	}
}

func TestDisplayVoteHandler(t *testing.T) {
	if err := initDatabaseAndTemplates(); err != nil {
		t.Fatal(err)
	}

	router := http.DefaultServeMux
	router.HandleFunc("/displayvote/{rotiid}", displayROTIHandler)

	existingROTI, nonExistingROTI := generateTestsROTIs()
	currentROTI, err := model.GetROTI(model.ROTIID(existingROTI))
	if err != nil {
		t.Fatal(err)
	}
	if err := currentROTI.AddVoteToROTI(3.0, "ok"); err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		query              string
		expectedStatusCode int
	}{
		{"/displayvote", 404}, // Without r parameter
		{fmt.Sprintf("/displayvote/%d", nonExistingROTI), 406}, // With a roti (r) that doesn't exist
		{fmt.Sprintf("/displayvote/%d", existingROTI), 200},    // With a roti (r) that exists
	}

	for _, tc := range testCases {
		t.Run(tc.query, func(t *testing.T) {
			code, err := testRouter(tc.query, "GET", router)
			if err != nil {
				t.Fatal(err)
			}
			if code != tc.expectedStatusCode {
				t.Errorf("Handler returned wrong status code: got %d want %d", code, tc.expectedStatusCode)
			}
		})
	}
}

func TestPostROTIHandler(t *testing.T) {
	if err := initDatabaseAndTemplates(); err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		query              string
		form               map[string][]string
		expectedStatusCode int
		expectedHidden     bool
		expectedFeedback   bool
	}{
		{"/newroti", nil, 406, false, false}, // Without form
		{"/newroti", map[string][]string{"rotiname": {"test"}, "hide": {"on"}, "feedback": {"on"}}, 303, true, true},      // With form
		{"/newroti", map[string][]string{"rotiname": {"test2"}, "hide": {"off"}, "feedback": {"off"}}, 303, false, false}, // With form, no options
	}

	for _, tc := range testCases {
		t.Run(tc.query, func(t *testing.T) {
			req, err := http.NewRequest("POST", tc.query, nil)
			if err != nil {
				t.Fatal(err)
			}
			if tc.form != nil {
				req.PostForm = tc.form
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			}

			rr := httptest.NewRecorder()
			postROTIHandler(rr, req)

			if rr.Code != tc.expectedStatusCode {
				t.Errorf("Handler returned wrong status code: got %d want %d", rr.Code, tc.expectedStatusCode)
			}

			if rr.Code == http.StatusSeeOther {
				location := rr.Header().Get("Location")
				if !strings.HasPrefix(location, "/roti/") {
					t.Errorf("Unexpected redirect URL: %s", location)
				}

				re := regexp.MustCompile(`/roti/(\d+)`)
				matches := re.FindStringSubmatch(location)
				rotiID, err := strconv.Atoi(matches[1])
				if err != nil {
					t.Fatal(err)
				}
				currentROTI, _ := model.GetROTI(model.ROTIID(rotiID))
				if err != nil {
					t.Fatal(err)
				}

				if currentROTI.IsHidden() != tc.expectedHidden {
					t.Errorf("ROTI %d has wrong hidden option: got %t want %t", rotiID, currentROTI.IsHidden(), tc.expectedHidden)
				}

				if currentROTI.HasFeedback() != tc.expectedFeedback {
					t.Errorf("ROTI %d has wrong feedback option: got %t want %t", rotiID, currentROTI.HasFeedback(), tc.expectedFeedback)
				}
			}
		})
	}
}

func TestPostVoteHandler(t *testing.T) {
	if err := initDatabaseAndTemplates(); err != nil {
		t.Fatal(err)
	}

	router := http.DefaultServeMux
	router.HandleFunc("/vote/{rotiid}", postVoteHandler)

	rotiID := model.CreateROTI("test", true, true, 30)

	testCases := []struct {
		query              string
		expectedStatusCode int
	}{
		{"/vote", 404},                               // Without form, without roti
		{"/vote/aaaaa", 406},                         // Without form, invalid roti
		{"/vote/99999", 406},                         // Without form, bad roti
		{fmt.Sprintf("/vote/%d", rotiID.Int()), 406}, // Without form, good roti
		{fmt.Sprintf("/vote/%d?vote=%s&feedback=%s", rotiID.Int(), "99.99", "test"), 406}, // With form, bad vote, good roti
		{fmt.Sprintf("/vote/%d?vote=%s&feedback=%s", rotiID.Int(), "2.5", "test"), 302},   // With form, good vote, good roti
	}

	for _, tc := range testCases {
		t.Run(tc.query, func(t *testing.T) {
			code, err := testRouter(tc.query, "POST", router)
			if err != nil {
				t.Fatal(err)
			}
			if code != tc.expectedStatusCode {
				t.Errorf("Handler returned wrong status code: got %d want %d", code, tc.expectedStatusCode)
			}
		})
	}
}

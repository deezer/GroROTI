package services

import (
	"context"
	"errors"
	"fmt"
	"image/png"
	"io/fs"
	"net/http"
	"os"
	"strconv"

	"github.com/deezer/groroti/internal/middlewares"
	"github.com/deezer/groroti/internal/model"
	"github.com/deezer/groroti/internal/staticEmbed"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/trace"
)

var (
	ErrTemplateParseFile = errors.New("error while parsing template file")
	ErrTemplateExecute   = errors.New("error while execution of template file")
	ErrParsingToInt      = errors.New("error while parsing to int")
	Version              string
	tracer               trace.Tracer
)

type existingROTI struct {
	Id           int
	Description  string
	NumVotes     int
	Avg          float64
	Min          float64
	Max          float64
	Url          string
	Feedbacks    []string
	UserHasVoted bool
	Version      string
}

func Register() *http.ServeMux {
	router := http.DefaultServeMux

	// launch the periodic process that collects the metrics
	recordMetrics()

	// Set up OpenTelemetry tracer
	if currentConfig.EnableTracing {
		tracer = middlewares.TP.Tracer("github.com/deezer/groroti/internal/services")
	}

	// Prometheus + liveness/readiness
	router.Handle("GET /-/liveness", NewHealthHandler())
	router.Handle("GET /-/metrics", NewMetricsHandler())
	router.Handle("GET /-/readiness", NewHealthHandler())

	// Real application
	router.Handle("GET /{$}", middlewares.MiddlewareChain("/", http.HandlerFunc(homeHandler)))
	router.Handle("GET /downpng/{rotiid}", middlewares.MiddlewareChain("/downpng", http.HandlerFunc(downloadPNGHandler)))
	router.Handle("GET /downcsv/{rotiid}", middlewares.MiddlewareChain("/downcsv", http.HandlerFunc(downloadCSVHandler)))
	router.Handle("GET /roti/{rotiid}", middlewares.MiddlewareChain("/roti", http.HandlerFunc(displayROTIHandler)))
	router.Handle("GET /roti", middlewares.MiddlewareChain("/roti", http.HandlerFunc(displayROTIHandlerLegacy)))
	router.Handle("POST /displayvote/{rotiid}", middlewares.MiddlewareChain("/displayvote", http.HandlerFunc(displayVoteHandler)))
	router.Handle("POST /newroti", middlewares.MiddlewareChain("/newroti", http.HandlerFunc(postROTIHandler)))
	router.Handle("POST /vote/{rotiid}", middlewares.MiddlewareChain("/vote", http.HandlerFunc(postVoteHandler)))

	// Create a sub-file system for embedded static files
	staticFS, err := fs.Sub(staticEmbed.EmbeddedStatic, "static")
	if err != nil {
		log.Fatal().Err(err)
		return nil
	}
	staticFileServer := http.FileServer(http.FS(staticFS))
	router.Handle("GET /static/", http.StripPrefix("/static/", staticFileServer))

	fsQr := http.FileServer(http.Dir("data/qr"))
	router.Handle("GET /qr/", http.StripPrefix("/qr/", fsQr))

	return router
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	var span trace.Span
	if currentConfig.EnableTracing {
		_, span = tracer.Start(r.Context(), "home")
		defer span.End()
	}

	templateFilePath := "templates/index.html"
	t, ok := staticEmbed.Templates[templateFilePath]
	if !ok {
		log.Error().Msgf("template %s not found", templateFilePath)
		return
	}

	var template struct {
		List    []model.ShortROTIInfo
		Version string
	}
	template.List = model.ListROTIs()
	template.Version = Version

	if currentConfig.EnableTracing {
		span.AddEvent("Start executing template")
	}
	err := t.Execute(w, template)
	if err != nil {
		log.Error().Err(ErrTemplateExecute)
		return
	}
}

func displayROTIHandler(w http.ResponseWriter, r *http.Request) {
	var span trace.Span
	var ctx context.Context
	if currentConfig.EnableTracing {
		ctx, span = tracer.Start(r.Context(), "display ROTI")
		defer span.End()
	}

	var currentROTI model.ROTIEntity

	rotiID, err := getIDFromURL(r, false)
	if err != nil {
		logErrorAndGoBackHome(err, w, r)
		return
	}

	currentROTI, err = model.GetROTI(model.ROTIID(rotiID))
	// protects from IDs that match no existing ROTI
	if err != nil {
		logErrorAndGoBackHome(err, w, r)
		return
	}

	currentConfig, err := GetConfig()
	if err != nil {
		log.Error().Err(err)
		return
	}

	// checks if QRcode exists or not. If not, generates one
	strID := strconv.Itoa(rotiID)
	if _, err := os.Stat("qr/qr" + strID + ".png"); errors.Is(err, os.ErrNotExist) {
		if err := genQRCode(currentConfig.GetURL(), strID, ctx); err != nil {
			log.Warn().Err(ErrQRCodeGeneration)
		}
	}

	hasVoted, _ := hasVotedForROTI(r, rotiID)

	template := existingROTI{
		Id:           rotiID,
		Description:  currentROTI.GetDescription(),
		NumVotes:     currentROTI.CountVotes(),
		Avg:          currentROTI.VotesAverage(),
		Min:          currentROTI.GetMinVote(),
		Max:          currentROTI.GetMaxVote(),
		Url:          currentConfig.GetURL(),
		Feedbacks:    currentROTI.ListFeedbacks(),
		UserHasVoted: hasVoted,
		Version:      Version,
	}

	templateFilePath := "templates/roti.html"
	t, ok := staticEmbed.Templates[templateFilePath]
	if !ok {
		log.Error().Msgf("template %s not found", templateFilePath)
		return
	}

	exportAsPNG(template)

	if currentConfig.EnableTracing {
		span.AddEvent("Start executing template")
	}

	err = t.Execute(w, template)
	if err != nil {
		log.Error().Err(ErrTemplateExecute)
		return
	}
}

func displayROTIHandlerLegacy(w http.ResponseWriter, r *http.Request) {
	rotiID, err := getIDFromURL(r, true)
	if err != nil {
		logErrorAndGoBackHome(err, w, r)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/roti/%d", rotiID), http.StatusPermanentRedirect)
}

func displayVoteHandler(w http.ResponseWriter, r *http.Request) {
	var span trace.Span
	if currentConfig.EnableTracing {
		_, span = tracer.Start(r.Context(), "display vote")
		defer span.End()
	}

	var currentROTI model.ROTIEntity

	rotiID, err := getIDFromURL(r, false)
	if err != nil {
		logErrorAndGoBackHome(err, w, r)
		return
	}

	if currentConfig.EnableTracing {
		span.AddEvent("get current ROTI")
	}

	currentROTI, err = model.GetROTI(model.ROTIID(rotiID))
	// protects from IDs that match no existing ROTI
	if err != nil {
		logErrorAndGoBackHome(err, w, r)
		return
	}

	templateFilePath := "templates/vote.html"
	t, ok := staticEmbed.Templates[templateFilePath]
	if !ok {
		log.Error().Msgf("template %s not found", templateFilePath)
		return
	}

	var template struct {
		RotiID      string
		VoteStep    string
		Description string
		HasFeedback bool
		Version     string
	}
	template.RotiID = strconv.Itoa(rotiID)
	template.VoteStep = fmt.Sprintf("%f", currentConfig.VoteStep)
	template.Description = currentROTI.GetDescription()
	template.HasFeedback = currentROTI.HasFeedback()
	template.Version = Version

	if currentConfig.EnableTracing {
		span.AddEvent("Start executing template")
	}

	err = t.Execute(w, template)
	if err != nil {
		log.Error().Err(ErrTemplateExecute)
		return
	}
}

func downloadPNGHandler(w http.ResponseWriter, r *http.Request) {
	var span trace.Span
	if currentConfig.EnableTracing {
		_, span = tracer.Start(r.Context(), "download PNG")
		defer span.End()
	}

	var currentROTI model.ROTIEntity

	rotiID, err := getIDFromURL(r, false)
	if err != nil {
		logErrorAndGoBackHome(err, w, r)
		return
	}

	currentROTI, err = model.GetROTI(model.ROTIID(rotiID))
	// protects from IDs that match no existing ROTI
	if err != nil {
		logErrorAndGoBackHome(err, w, r)
		return
	}

	template := existingROTI{
		Id:          rotiID,
		Description: currentROTI.GetDescription(),
		NumVotes:    currentROTI.CountVotes(),
		Avg:         currentROTI.VotesAverage(),
		Min:         currentROTI.GetMinVote(),
		Max:         currentROTI.GetMaxVote(),
		Feedbacks:   currentROTI.ListFeedbacks(),
	}

	img := exportAsPNG(template)

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=roti_%d.png", rotiID))
	w.Header().Set("Content-Type", "image/png")

	if err := png.Encode(w, img); err != nil {
		logErrorAndGoBackHome(err, w, r)
		return
	}
}

func downloadCSVHandler(w http.ResponseWriter, r *http.Request) {
	var span trace.Span
	if currentConfig.EnableTracing {
		_, span = tracer.Start(r.Context(), "download CSV")
		defer span.End()
	}

	var currentROTI model.ROTIEntity

	rotiID, err := getIDFromURL(r, false)
	if err != nil {
		logErrorAndGoBackHome(err, w, r)
		return
	}

	currentROTI, err = model.GetROTI(model.ROTIID(rotiID))
	// protects from IDs that match no existing ROTI
	if err != nil {
		logErrorAndGoBackHome(err, w, r)
		return
	}

	template := existingROTI{
		Id:          rotiID,
		Description: currentROTI.GetDescription(),
		NumVotes:    currentROTI.CountVotes(),
		Avg:         currentROTI.VotesAverage(),
		Min:         currentROTI.GetMinVote(),
		Max:         currentROTI.GetMaxVote(),
	}

	csvContent := exportAsCSV(template)

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=roti_%d.csv", rotiID))
	w.Header().Set("Content-Type", "text/csv")

	for _, line := range csvContent {
		_, err := fmt.Fprintln(w, line)
		if err != nil {
			logErrorAndGoBackHome(err, w, r)
			return
		}
	}
}

func postROTIHandler(w http.ResponseWriter, r *http.Request) {
	var span trace.Span
	if currentConfig.EnableTracing {
		_, span = tracer.Start(r.Context(), "create ROTI")
		defer span.End()
	}

	var rotiname string
	var hide, feedback bool

	// get ROTI name from form if present. "" if not
	if err := r.ParseForm(); err != nil {
		logErrorAndGoBackHome(err, w, r)
		return
	}
	rotiname = r.Form.Get("rotiname")
	hide = false
	if r.Form.Get("hide") == "on" {
		hide = true
	}
	feedback = false
	if r.Form.Get("feedback") == "on" {
		feedback = true
	}
	rotiID := model.CreateROTI(rotiname, hide, feedback, currentConfig.CleanOverTime)

	http.Redirect(w, r, "/roti/"+strconv.Itoa(int(rotiID)), http.StatusSeeOther)
}

func postVoteHandler(w http.ResponseWriter, r *http.Request) {
	var span trace.Span
	if currentConfig.EnableTracing {
		_, span = tracer.Start(r.Context(), "create vote")
		defer span.End()
	}

	rotiID, err := getIDFromURL(r, false)
	if err != nil {
		logErrorAndGoBackHome(err, w, r)
		return
	}

	currentROTI, err := model.GetROTI(model.ROTIID(rotiID))
	if err != nil {
		logErrorAndGoBackHome(err, w, r)
		return
	}

	feedback := r.FormValue("feedback")
	// check vote validity
	vote, err := model.CheckVote(r.FormValue("vote"))
	if err != nil {
		log.Error().Msgf(model.ErrInvalidVote.Error())
		http.Redirect(w, r, "/roti/"+strconv.Itoa(rotiID), http.StatusNotAcceptable)
		return
	}

	if hasVoted, _ := hasVotedForROTI(r, rotiID); hasVoted {
		log.Warn().Msgf("User has already voted for ROTI " + strconv.Itoa(rotiID))
		http.Redirect(w, r, "/roti/"+strconv.Itoa(rotiID), http.StatusFound)
		return
	}

	if err := currentROTI.AddVoteToROTI(vote, feedback); err != nil {
		log.Warn().Msgf(err.Error())
		http.Redirect(w, r, "/roti/"+strconv.Itoa(rotiID), http.StatusFound)
		return
	}

	// Put a cookie to mark that the user has voted for this ROTI
	setVotedCookie(w, rotiID, r.Context())

	http.Redirect(w, r, "/roti/"+strconv.Itoa(rotiID), http.StatusFound)
}

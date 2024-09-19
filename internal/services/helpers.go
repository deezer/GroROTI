package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/deezer/groroti/internal/model"
	"go.opentelemetry.io/otel/trace"
	"github.com/rs/zerolog/log"
	qrcode "github.com/skip2/go-qrcode"
)

var (
	ErrQRCodeGeneration = errors.New("error during QRcode generation")
)

// getIDFromURL() takes the id in the URL and checks if it's a valid int comprised
// between 10000 and 99999
func getIDFromURL(r *http.Request, legacy_routing bool) (rotiID int, err error) {
	var span trace.Span
	if currentConfig.EnableTracing {
		_, span = tracer.Start(r.Context(), "getIDFromURL")
	defer span.End()
	}

	var urlRotiId string
	if legacy_routing {
		urlRotiId = r.URL.Query().Get("r")
	} else {
		urlRotiId = r.PathValue("rotiid")
	}

	if urlRotiId == "" {
		return 0, model.ErrInvalidROTIID
	} else {
		rotiID, err = strconv.Atoi(urlRotiId)
	}

	if rotiID < 10000 || rotiID > 99999 {
		return 0, model.ErrInvalidROTIID
	}

	return
}

func setVotedCookie(w http.ResponseWriter, rotiID int, ctx context.Context) {
	var span trace.Span
	if currentConfig.EnableTracing {
		_, span = tracer.Start(ctx, "setVotedCookie")
		defer span.End()
	}

	cookie := http.Cookie{
		Name:     "voted_roti_" + strconv.Itoa(rotiID),
		Value:    "true",
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60,
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
}

func hasVotedForROTI(r *http.Request, rotiID int) (bool, error) {
	var span trace.Span
	if currentConfig.EnableTracing {
		_, span = tracer.Start(r.Context(), "hasVotedForROTI")
		defer span.End()
	}

	cookieName := "voted_roti_" + strconv.Itoa(rotiID)
	_, err := r.Cookie(cookieName)
	if err == nil {
		return true, nil
	} else if errors.Is(err, http.ErrNoCookie) {
		return false, nil
	}
	return false, err
}

func genQRCode(url string, strid string, ctx context.Context) (err error) {
	var span trace.Span
	if currentConfig.EnableTracing {
		_, span = tracer.Start(ctx, "genQRCode")
		defer span.End()
	}

	// check directory tree for data/qr
	qrDir := "data/qr"
	if _, err := os.Stat(qrDir); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(qrDir, os.ModePerm)
		if err != nil {
			log.Error().Err(err)
		}
	}

	// generate qrcode image
	fullUrl := url + "/roti/" + strid

	return qrcode.WriteFile(fullUrl, qrcode.Medium, currentConfig.GetQrCodeSize(),
		fmt.Sprintf("%s/qr%s.png", qrDir, strid))
}

func logErrorAndGoBackHome(err error, w http.ResponseWriter, r *http.Request) {
	log.Error().Msgf(err.Error())
	http.Redirect(w, r, "/", http.StatusNotAcceptable)
}

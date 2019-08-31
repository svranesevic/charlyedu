package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/svranesevic/charlyedu/weatherservice"
	log "go.uber.org/zap"
)

func GetWeather(ws weatherservice.Service, w http.ResponseWriter, r *http.Request) {
	startStr := r.FormValue("start")
	endStr := r.FormValue("end")

	start, err := time.Parse("2006-01-02T15:04:05Z0700", startStr)
	if err != nil {
		http.Error(w, NewErrorResponse("`start` must be an ISO8601 DateTime"), http.StatusBadRequest)
		return
	}

	end, err := time.Parse("2006-01-02T15:04:05Z0700", endStr)
	if err != nil {
		http.Error(w, NewErrorResponse("`end` must be an ISO8601 DateTime"), http.StatusBadRequest)
		return
	}

	temps, err := ws.GetForRange(r.Context(), start, end)
	if err != nil {
		http.Error(w, NewErrorResponse(err.Error()), http.StatusInternalServerError)
		return
	}

	if err = json.NewEncoder(w).Encode(temps); err != nil {
		log.S().Errorf("Failed to marshal response: %+v", err)
		http.Error(w, NewErrorResponse("Woops, something went wrong, try again"), http.StatusInternalServerError)
	}
}

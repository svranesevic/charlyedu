package router

import (
	"context"
	"time"

	"github.com/gorilla/mux"
	"github.com/svranesevic/charlyedu/handler"
	"github.com/svranesevic/charlyedu/temperatureservice"
	"github.com/svranesevic/charlyedu/weatherservice"
	"github.com/svranesevic/charlyedu/windspeedservice"

	//"github.com/svranesevic/charlyedu/handler"
	"net/http"
)

func New(ts temperatureservice.Service, wss windspeedservice.Service, ws weatherservice.Service) *mux.Router {
	router := mux.NewRouter()

	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")
			next.ServeHTTP(w, r)
		})
	})
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
			defer cancel()

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})

	initializeTemperatureRoutes(ts, router)
	initializeWindSpeedRoutes(wss, router)
	initializeWeatherRoutes(ws, router)

	return router
}

func initializeTemperatureRoutes(ts temperatureservice.Service, router *mux.Router) {
	router.
		Path("/temperatures").
		Queries("start", "{start}").
		Queries("end", "{end}").
		Methods("GET").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handler.GetTemperature(ts, w, r)
		}).
		Name("GetTemperature")
}

func initializeWindSpeedRoutes(wss windspeedservice.Service, router *mux.Router) {
	router.
		Path("/speeds").
		Queries("start", "{start}").
		Queries("end", "{end}").
		Methods("GET").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handler.GetWindSpeed(wss, w, r)
		}).
		Name("GetWindSpeed")
}

func initializeWeatherRoutes(ws weatherservice.Service, router *mux.Router) {
	router.
		Path("/weather").
		Queries("start", "{start}").
		Queries("end", "{end}").
		Methods("GET").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handler.GetWeather(ws, w, r)
		}).
		Name("GetWeather")
}

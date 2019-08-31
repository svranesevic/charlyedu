package main

import (
	"fmt"
	"net/http"

	"github.com/kelseyhightower/envconfig"
	"github.com/svranesevic/charlyedu/router"
	"github.com/svranesevic/charlyedu/temperatureservice"
	"github.com/svranesevic/charlyedu/weatherservice"
	"github.com/svranesevic/charlyedu/windspeedservice"
	log "go.uber.org/zap"
)

type config struct {
	Port               uint64 `default:"3000"`
	TemperatureService string `default:"http://localhost:8000/" split_words:"true"`
	WindSpeedService   string `default:"http://localhost:8080/" split_words:"true"`
}

func main() {
	logger, _ := log.NewDevelopment()
	log.ReplaceGlobals(logger)

	var c config
	if err := envconfig.Process("", &c); err != nil {
		log.S().Fatalf("Unable to process ENV config: %v\n", err.Error())
	}

	ts := temperatureservice.New(c.TemperatureService)
	wss := windspeedservice.New(c.WindSpeedService)
	ws := weatherservice.New(ts, wss)

	r := router.New(ts, wss, ws)

	addr := fmt.Sprintf(":%d", c.Port)
	log.S().Infof("Server starting on %s", addr)
	log.S().Fatal(http.ListenAndServe(addr, r).Error())
}

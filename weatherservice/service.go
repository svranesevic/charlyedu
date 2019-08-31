package weatherservice

import (
	"context"
	"errors"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/svranesevic/charlyedu/temperatureservice"
	"github.com/svranesevic/charlyedu/windspeedservice"
	log "go.uber.org/zap"
)

type Service interface {
	GetForRange(ctx context.Context, from time.Time, to time.Time) ([]Weather, error)
	GetForDateTime(ctx context.Context, at time.Time) (*Weather, error)
}

type weatherService struct {
	ts  temperatureservice.Service
	wss windspeedservice.Service
}

func New(ts temperatureservice.Service, wss windspeedservice.Service) Service {
	return weatherService{ts: ts, wss: wss}
}

func (ws weatherService) GetForRange(ctx context.Context, from time.Time, to time.Time) ([]Weather, error) {
	if from.After(to) {
		return []Weather{}, errors.New("`start` must be before `end`")
	}
	to = to.Add(24 * time.Hour)

	numDays := int(math.Ceil(to.Sub(from).Hours() / 24))
	tempChan := make(chan Weather, numDays)
	var wg sync.WaitGroup

	for at := from; at.Before(to); at = at.Add(24 * time.Hour) {
		wg.Add(1)
		go func(at time.Time) {
			defer wg.Done()

			if t, err := ws.GetForDateTime(ctx, at); err != nil {
				log.S().Errorf("failed to obtain weather for datetime %s, %+v", at, err)
			} else if t != nil {
				tempChan <- *t
			}
		}(at)
	}
	wg.Wait()
	close(tempChan)

	temps := make(weathersSlice, 0)
	for t := range tempChan {
		temps = append(temps, t)
	}

	sort.Sort(temps)
	return temps, nil
}

func (ws weatherService) GetForDateTime(ctx context.Context, at time.Time) (*Weather, error) {
	var wg sync.WaitGroup

	wg.Add(1)
	tempRespChan := make(chan temperatureResponse, 1)
	go func() {
		defer wg.Done()
		defer close(tempRespChan)

		temp, err := ws.ts.GetForDateTime(ctx, at)
		tempRespChan <- temperatureResponse{Temperature: temp, Error: err}
	}()

	wg.Add(1)
	windSpeedRespChan := make(chan windSpeedResponse, 1)
	go func() {
		defer wg.Done()
		defer close(windSpeedRespChan)

		windSpeed, err := ws.wss.GetForDateTime(ctx, at)
		windSpeedRespChan <- windSpeedResponse{WindSpeed: windSpeed, Error: err}
	}()

	wg.Wait()

	weather := Weather{Date: at}

	if tempResp := <-tempRespChan; tempResp.Error != nil {
		return nil, tempResp.Error
	} else if tempResp.Temperature == nil {
		return nil, nil
	} else {
		temp := tempResp.Temperature
		weather.Temperature = temp.Temperature
	}

	if windSpeedResp := <-windSpeedRespChan; windSpeedResp.Error != nil {
		return nil, windSpeedResp.Error
	} else if windSpeedResp.WindSpeed == nil {
		return nil, nil
	} else {
		windSpeed := windSpeedResp.WindSpeed
		weather.North = windSpeed.North
		weather.West = windSpeed.West
	}

	return &weather, nil
}

type temperatureResponse struct {
	Temperature *temperatureservice.Temperature
	Error       error
}

type windSpeedResponse struct {
	WindSpeed *windspeedservice.WindSpeed
	Error     error
}

package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/svranesevic/charlyedu/weatherservice"
)

type weatherServiceStub struct {
	Weathers []weatherservice.Weather
}

func (s weatherServiceStub) GetForRange(ctx context.Context, from time.Time, to time.Time) ([]weatherservice.Weather, error) {
	temps := make([]weatherservice.Weather, 0)
	for _, temp := range s.Weathers {
		if temp.Date.After(from.Add(-24*time.Hour)) && temp.Date.Before(to.Add(24*time.Hour)) {
			temps = append(temps, temp)
		}
	}

	return temps, nil
}

func (s weatherServiceStub) GetForDateTime(ctx context.Context, at time.Time) (*weatherservice.Weather, error) {
	for _, temp := range s.Weathers {
		if temp.Date.Equal(at) {
			return &temp, nil
		}
	}

	return nil, nil
}

type phallicWeatherServiceStub struct {
}

func (s phallicWeatherServiceStub) GetForRange(ctx context.Context, from time.Time, to time.Time) ([]weatherservice.Weather, error) {
	return []weatherservice.Weather{}, errors.New("GetForRange error")
}

func (s phallicWeatherServiceStub) GetForDateTime(ctx context.Context, at time.Time) (*weatherservice.Weather, error) {
	return nil, errors.New("GetForRange error")
}

func TestGetWeatherReturnsWeathers(t *testing.T) {
	weatherService := weatherServiceStub{
		Weathers: []weatherservice.Weather{
			{
				Date:        time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
				Temperature: 1.1,
				North:       1.2,
				West:        1.3,
			},
			{
				Date:        time.Date(2019, 1, 2, 0, 0, 0, 0, time.UTC),
				Temperature: 2.1,
				North:       2.2,
				West:        2.3,
			},
			{
				Date:        time.Date(2019, 1, 3, 0, 0, 0, 0, time.UTC),
				Temperature: 3.1,
				North:       3.2,
				West:        3.3,
			},
			{
				Date:        time.Date(2019, 2, 1, 0, 0, 0, 0, time.UTC),
				Temperature: 4.1,
				North:       4.2,
				West:        4.3,
			},
			{
				Date:        time.Date(2019, 2, 2, 0, 0, 0, 0, time.UTC),
				Temperature: 5.1,
				North:       5.2,
				West:        5.3,
			},
		},
	}

	req, err := http.NewRequest("GET", "http://url.handled.by.router/weather?start=2019-01-01T00:00:00Z&end=2019-02-02T00:00:00Z", nil)
	assert.Nil(t, err)

	rec := httptest.NewRecorder()
	GetWeather(weatherService, rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	bodyBytes, err := ioutil.ReadAll(rec.Body)
	assert.Nil(t, err)

	var weathers []weatherservice.Weather
	err = json.Unmarshal(bodyBytes, &weathers)
	assert.Nil(t, err)

	assert.ElementsMatch(t, weatherService.Weathers, weathers)
}

func TestGetWeatherReturnsBadRequestErrorOnMissingStartQueryParam(t *testing.T) {
	weatherService := weatherServiceStub{}

	req, err := http.NewRequest("GET", "http://url.handled.by.router/weather?end=2019-02-02T00:00:00Z", nil)
	assert.Nil(t, err)

	rec := httptest.NewRecorder()
	GetWeather(weatherService, rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetWeatherReturnsBadRequestErrorOnMissingEndQueryParam(t *testing.T) {
	weatherService := weatherServiceStub{}

	req, err := http.NewRequest("GET", "http://url.handled.by.router/weather?start=2019-01-01T00:00:00Z", nil)
	assert.Nil(t, err)

	rec := httptest.NewRecorder()
	GetWeather(weatherService, rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetWeatherReturnsInternalServerErrorOnOnServiceError(t *testing.T) {
	weatherService := phallicWeatherServiceStub{}

	req, err := http.NewRequest("GET", "http://url.handled.by.router/weather?start=2019-01-01T00:00:00Z&end=2019-02-02T00:00:00Z", nil)
	assert.Nil(t, err)

	rec := httptest.NewRecorder()
	GetWeather(weatherService, rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

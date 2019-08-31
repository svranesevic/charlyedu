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
	"github.com/svranesevic/charlyedu/windspeedservice"
)

type windSpeedServiceStub struct {
	WindSpeeds []windspeedservice.WindSpeed
}

func (s windSpeedServiceStub) GetForRange(ctx context.Context, from time.Time, to time.Time) ([]windspeedservice.WindSpeed, error) {
	speeds := make([]windspeedservice.WindSpeed, 0)
	for _, speed := range s.WindSpeeds {
		if speed.Date.After(from.Add(-24*time.Hour)) && speed.Date.Before(to.Add(24*time.Hour)) {
			speeds = append(speeds, speed)
		}
	}

	return speeds, nil
}

func (s windSpeedServiceStub) GetForDateTime(ctx context.Context, at time.Time) (*windspeedservice.WindSpeed, error) {
	for _, speed := range s.WindSpeeds {
		if speed.Date.Equal(at) {
			return &speed, nil
		}
	}

	return nil, nil
}

type phallicWindSpeedServiceStub struct {
}

func (s phallicWindSpeedServiceStub) GetForRange(ctx context.Context, from time.Time, to time.Time) ([]windspeedservice.WindSpeed, error) {
	return []windspeedservice.WindSpeed{}, errors.New("GetForRange error")
}

func (s phallicWindSpeedServiceStub) GetForDateTime(ctx context.Context, at time.Time) (*windspeedservice.WindSpeed, error) {
	return nil, errors.New("GetForRange error")
}

func TestGetWindSpeedReturnsWindSpeeds(t *testing.T) {
	windSpeedService := windSpeedServiceStub{
		WindSpeeds: []windspeedservice.WindSpeed{
			{
				Date:  time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
				West:  1.1,
				North: 1.2,
			},
			{
				Date:  time.Date(2019, 1, 2, 0, 0, 0, 0, time.UTC),
				West:  2.1,
				North: 2.2,
			},
			{
				Date:  time.Date(2019, 1, 3, 0, 0, 0, 0, time.UTC),
				West:  3.1,
				North: 3.2,
			},
			{
				Date:  time.Date(2019, 2, 1, 0, 0, 0, 0, time.UTC),
				West:  4.1,
				North: 4.2,
			},
			{
				Date:  time.Date(2019, 2, 2, 0, 0, 0, 0, time.UTC),
				West:  5.1,
				North: 5.2,
			},
		},
	}

	req, err := http.NewRequest("GET", "http://url.handled.by.router/speeds?start=2019-01-01T00:00:00Z&end=2019-02-02T00:00:00Z", nil)
	assert.Nil(t, err)

	rec := httptest.NewRecorder()
	GetWindSpeed(windSpeedService, rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	bodyBytes, err := ioutil.ReadAll(rec.Body)
	assert.Nil(t, err)

	var windSpeeds []windspeedservice.WindSpeed
	err = json.Unmarshal(bodyBytes, &windSpeeds)
	assert.Nil(t, err)

	assert.ElementsMatch(t, windSpeedService.WindSpeeds, windSpeeds)
}

func TestGetWindSpeedReturnsBadRequestErrorOnMissingStartQueryParam(t *testing.T) {
	windSpeedService := windSpeedServiceStub{}

	req, err := http.NewRequest("GET", "http://url.handled.by.router/speeds?end=2019-02-02T00:00:00Z", nil)
	assert.Nil(t, err)

	rec := httptest.NewRecorder()
	GetWindSpeed(windSpeedService, rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetWindSpeedReturnsBadRequestErrorOnMissingEndQueryParam(t *testing.T) {
	windSpeedService := windSpeedServiceStub{}

	req, err := http.NewRequest("GET", "http://url.handled.by.router/speeds?start=2019-01-01T00:00:00Z", nil)
	assert.Nil(t, err)

	rec := httptest.NewRecorder()
	GetWindSpeed(windSpeedService, rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetWindSpeedReturnsInternalServerErrorOnOnServiceError(t *testing.T) {
	windSpeedService := phallicWindSpeedServiceStub{}

	req, err := http.NewRequest("GET", "http://url.handled.by.router/speeds?start=2019-01-01T00:00:00Z&end=2019-02-02T00:00:00Z", nil)
	assert.Nil(t, err)

	rec := httptest.NewRecorder()
	GetWindSpeed(windSpeedService, rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

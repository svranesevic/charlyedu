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
	"github.com/svranesevic/charlyedu/temperatureservice"
)

type temperatureServiceStub struct {
	Temperatures []temperatureservice.Temperature
}

func (s temperatureServiceStub) GetForRange(ctx context.Context, from time.Time, to time.Time) ([]temperatureservice.Temperature, error) {
	temps := make([]temperatureservice.Temperature, 0)
	for _, temp := range s.Temperatures {
		if temp.Date.After(from.Add(-24*time.Hour)) && temp.Date.Before(to.Add(24*time.Hour)) {
			temps = append(temps, temp)
		}
	}

	return temps, nil
}

func (s temperatureServiceStub) GetForDateTime(ctx context.Context, at time.Time) (*temperatureservice.Temperature, error) {
	for _, temp := range s.Temperatures {
		if temp.Date.Equal(at) {
			return &temp, nil
		}
	}

	return nil, nil
}

type phallicTemperatureServiceStub struct {
}

func (s phallicTemperatureServiceStub) GetForRange(ctx context.Context, from time.Time, to time.Time) ([]temperatureservice.Temperature, error) {
	return []temperatureservice.Temperature{}, errors.New("GetForRange error")
}

func (s phallicTemperatureServiceStub) GetForDateTime(ctx context.Context, at time.Time) (*temperatureservice.Temperature, error) {
	return nil, errors.New("GetForRange error")
}

func TestGetTemperatureReturnsTemperatures(t *testing.T) {
	tempService := temperatureServiceStub{
		Temperatures: []temperatureservice.Temperature{
			{
				Date:        time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
				Temperature: 1.1,
			},
			{
				Date:        time.Date(2019, 1, 2, 0, 0, 0, 0, time.UTC),
				Temperature: 2.2,
			},
			{
				Date:        time.Date(2019, 1, 3, 0, 0, 0, 0, time.UTC),
				Temperature: 3.3,
			},
			{
				Date:        time.Date(2019, 2, 1, 0, 0, 0, 0, time.UTC),
				Temperature: 4.4,
			},
			{
				Date:        time.Date(2019, 2, 2, 0, 0, 0, 0, time.UTC),
				Temperature: 5.5,
			},
		},
	}

	req, err := http.NewRequest("GET", "http://url.handled.by.router/temperatures?start=2019-01-01T00:00:00Z&end=2019-02-02T00:00:00Z", nil)
	assert.Nil(t, err)

	rec := httptest.NewRecorder()
	GetTemperature(tempService, rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	bodyBytes, err := ioutil.ReadAll(rec.Body)
	assert.Nil(t, err)

	var temps []temperatureservice.Temperature
	err = json.Unmarshal(bodyBytes, &temps)
	assert.Nil(t, err)

	assert.ElementsMatch(t, tempService.Temperatures, temps)
}

func TestGetTemperatureReturnsBadRequestErrorOnMissingStartQueryParam(t *testing.T) {
	tempService := temperatureServiceStub{}

	req, err := http.NewRequest("GET", "http://url.handled.by.router/temperatures?end=2019-02-02T00:00:00Z", nil)
	assert.Nil(t, err)

	rec := httptest.NewRecorder()
	GetTemperature(tempService, rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetTemperatureReturnsBadRequestErrorOnMissingEndQueryParam(t *testing.T) {
	tempService := temperatureServiceStub{}

	req, err := http.NewRequest("GET", "http://url.handled.by.router/temperatures?start=2019-01-01T00:00:00Z", nil)
	assert.Nil(t, err)

	rec := httptest.NewRecorder()
	GetTemperature(tempService, rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetTemperatureReturnsInternalServerErrorOnServiceError(t *testing.T) {
	tempService := phallicTemperatureServiceStub{}

	req, err := http.NewRequest("GET", "http://url.handled.by.router/temperatures?start=2019-01-01T00:00:00Z&end=2019-02-02T00:00:00Z", nil)
	assert.Nil(t, err)

	rec := httptest.NewRecorder()
	GetTemperature(tempService, rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

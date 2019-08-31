package temperatureservice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	log "go.uber.org/zap"
)

type temperatureService struct {
	host string
}

type Service interface {
	GetForRange(ctx context.Context, from time.Time, to time.Time) ([]Temperature, error)
	GetForDateTime(ctx context.Context, at time.Time) (*Temperature, error)
}

func New(host string) Service {
	if strings.HasSuffix(host, "/") {
		host = strings.TrimSuffix(host, "/")
	}
	return temperatureService{host: host}
}

func (ts temperatureService) GetForRange(ctx context.Context, from time.Time, to time.Time) ([]Temperature, error) {
	if from.After(to) {
		return []Temperature{}, errors.New("`start` must be before `end`")
	}
	to = to.Add(24 * time.Hour)

	numDays := int(math.Ceil(to.Sub(from).Hours() / 24))
	tempChan := make(chan Temperature, numDays)
	var wg sync.WaitGroup

	for at := from; at.Before(to); at = at.Add(24 * time.Hour) {
		wg.Add(1)
		go func(at time.Time) {
			defer wg.Done()

			if t, err := ts.GetForDateTime(ctx, at); err != nil {
				log.S().Errorf("failed to obtain temperature for datetime %s, %+v", at, err)
			} else if t != nil {
				tempChan <- *t
			}
		}(at)
	}
	wg.Wait()
	close(tempChan)

	temps := make(temperaturesSlice, 0)
	for t := range tempChan {
		temps = append(temps, t)
	}

	sort.Sort(temps)
	return temps, nil
}

func (ts temperatureService) GetForDateTime(ctx context.Context, at time.Time) (*Temperature, error) {
	url := fmt.Sprintf("%s/?at=%s", ts.host, at.Format("2006-01-02T15:04:05Z0700"))
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var t Temperature
	if err = json.Unmarshal(bodyBytes, &t); err != nil {
		return nil, err
	}

	return &t, nil
}

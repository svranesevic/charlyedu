package windspeedservice

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

type windSpeedService struct {
	host string
}

type Service interface {
	GetForRange(ctx context.Context, from time.Time, to time.Time) ([]WindSpeed, error)
	GetForDateTime(ctx context.Context, at time.Time) (*WindSpeed, error)
}

func New(host string) Service {
	if strings.HasSuffix(host, "/") {
		host = strings.TrimSuffix(host, "/")
	}
	return windSpeedService{host: host}
}

func (wss windSpeedService) GetForRange(ctx context.Context, from time.Time, to time.Time) ([]WindSpeed, error) {
	if from.After(to) {
		return []WindSpeed{}, errors.New("`start` must be before `end`")
	}
	to = to.Add(24 * time.Hour)

	numDays := int(math.Ceil(to.Sub(from).Hours() / 24))
	wsChan := make(chan WindSpeed, numDays)
	var wg sync.WaitGroup

	for at := from; at.Before(to); at = at.Add(24 * time.Hour) {
		wg.Add(1)
		go func(at time.Time) {
			defer wg.Done()

			if ws, err := wss.GetForDateTime(ctx, at); err != nil {
				log.S().Errorf("failed to obtain wind speed for datetime %s, %+v", at, err)
			} else if ws != nil {
				wsChan <- *ws
			}
		}(at)
	}
	wg.Wait()
	close(wsChan)

	windSpeeds := make(speedsSlice, 0)
	for t := range wsChan {
		windSpeeds = append(windSpeeds, t)
	}

	sort.Sort(windSpeeds)
	return windSpeeds, nil
}

func (wss windSpeedService) GetForDateTime(ctx context.Context, at time.Time) (*WindSpeed, error) {
	url := fmt.Sprintf("%s/?at=%s", wss.host, at.Format("2006-01-02T15:04:05Z0700"))
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var t WindSpeed
	if err = json.Unmarshal(bodyBytes, &t); err != nil {
		return nil, err
	}

	return &t, nil
}

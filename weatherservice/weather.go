package weatherservice

import "time"

type Weather struct {
	North       float64   `json:"north"`
	West        float64   `json:"west"`
	Temperature float64   `json:"temp"`
	Date        time.Time `json:"date"`
}

type weathersSlice []Weather

func (w weathersSlice) Len() int {
	return len(w)
}

func (w weathersSlice) Less(i, j int) bool {
	return w[i].Date.Before(w[j].Date)
}

func (w weathersSlice) Swap(i, j int) {
	w[i], w[j] = w[j], w[i]
}

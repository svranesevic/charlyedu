package temperatureservice

import "time"

type Temperature struct {
	Temperature float64   `json:"temp"`
	Date        time.Time `json:"date"`
}

type temperaturesSlice []Temperature

func (t temperaturesSlice) Len() int {
	return len(t)
}

func (t temperaturesSlice) Less(i, j int) bool {
	return t[i].Date.Before(t[j].Date)
}

func (t temperaturesSlice) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

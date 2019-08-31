package windspeedservice

import "time"

type WindSpeed struct {
	North float64   `json:"north"`
	West  float64   `json:"west"`
	Date  time.Time `json:"date"`
}

type speedsSlice []WindSpeed

func (s speedsSlice) Len() int {
	return len(s)
}

func (s speedsSlice) Less(i, j int) bool {
	return s[i].Date.Before(s[j].Date)
}

func (s speedsSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
